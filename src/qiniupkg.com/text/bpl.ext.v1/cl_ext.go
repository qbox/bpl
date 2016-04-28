package bpl

import (
	"fmt"
	"strings"

	"qiniupkg.com/text/bpl.v1"
	"qiniupkg.com/text/tpl.v1/interpreter.util"
	"qlang.io/exec.v2"
)

// -----------------------------------------------------------------------------

type executor struct {
	code exec.Code
	estk *exec.Stack
}

func newExecutor() *executor {
	return &executor{
		estk: exec.NewStack(),
	}
}

var (
	nilVars = map[string]interface{}{}
)

func (p *executor) Eval(ctx *bpl.Context, start, end int) interface{} {

	vars, _ := ctx.Dom().(map[string]interface{})
	if vars == nil {
		vars = nilVars
	}
	code := &p.code
	stk := p.estk
	ectx := exec.NewSimpleContext(vars, stk, code)
	code.Exec(start, end, stk, ectx)
	v, _ := stk.Pop()
	return v
}

// -----------------------------------------------------------------------------

type exprBlock struct {
	start int
	end   int
}

func (p *Compiler) istart() {

	p.idxStart = p.code.Len()
}

func (p *Compiler) iend() {

	end := p.code.Len()
	p.gstk.Push(&exprBlock{start: p.idxStart, end: end})
}

func (p *Compiler) popExpr() *exprBlock {

	if v, ok := p.gstk.Pop(); ok {
		if e, ok := v.(*exprBlock); ok {
			return e
		}
	}
	panic("no index expression")
}

func (p *Compiler) array() {

	e := p.popExpr()
	stk := p.stk
	i := len(stk) - 1
	n := func(ctx *bpl.Context) int {
		v := p.Eval(ctx.Parent, e.start, e.end)
		return toInt(v, "index isn't an integer expression")
	}
	stk[i] = bpl.Dynarray(stk[i].(bpl.Ruler), n)
}

func (p *Compiler) array0() {

	stk := p.stk
	i := len(stk) - 1
	stk[i] = bpl.Array0(stk[i].(bpl.Ruler))
}

func (p *Compiler) array1() {

	stk := p.stk
	i := len(stk) - 1
	stk[i] = bpl.Array1(stk[i].(bpl.Ruler))
}

// -----------------------------------------------------------------------------

func (p *Compiler) casei(v int) {

	p.gstk.Push(v)
}

func (p *Compiler) source(v interface{}) {

	p.gstk.Push(v)
}

func (p *Compiler) popRule() bpl.Ruler {

	n := len(p.stk) - 1
	r := p.stk[n].(bpl.Ruler)
	p.stk = p.stk[:n]
	return r
}

func (p *Compiler) popRules(m int) bpl.Ruler {

	if m == 0 {
		return nil
	}
	p.and(m)
	return p.popRule()
}

func sourceOf(engine interpreter.Engine, src interface{}) string {

	b := engine.Source(src)
	return strings.Trim(string(b), " \t\r\n")
}

func (p *Compiler) fnCase(engine interpreter.Engine) {

	var defaultR bpl.Ruler
	if p.popArity() != 0 {
		defaultR = p.popRule()
	}

	arity := p.popArity()

	stk := p.stk
	n := len(stk)
	caseRs := clone(stk[n-arity:])
	caseExprAndSources := p.gstk.PopNArgs(arity << 1)
	e := p.popExpr()
	srcSw, _ := p.gstk.Pop()
	r := func(ctx *bpl.Context) (bpl.Ruler, error) {
		v := p.Eval(ctx, e.start, e.end)
		for i := 0; i < len(caseExprAndSources); i += 2 {
			expr := caseExprAndSources[i]
			if eq(v, expr) {
				if SetCaseType {
					key := sourceOf(engine, srcSw)
					val := sourceOf(engine, caseExprAndSources[i+1])
					ctx.SetVar(key+".kind", val)
				}
				return caseRs[i>>1], nil
			}
		}
		if defaultR != nil {
			return defaultR, nil
		}
		return nil, fmt.Errorf("case `%s(=%v)` is not found", sourceOf(engine, srcSw), v)
	}
	stk[n-arity] = bpl.Dyntype(r)
	p.stk = stk[:n-arity+1]
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnEval() {

	e := p.popExpr()
	stk := p.stk
	i := len(stk) - 1
	expr := func(ctx *bpl.Context) []byte {
		v := p.Eval(ctx, e.start, e.end)
		return v.([]byte)
	}
	stk[i] = bpl.Eval(expr, stk[i].(bpl.Ruler))
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnAssert(src interface{}) {

	e := p.popExpr()
	expr := func(ctx *bpl.Context) bool {
		v := p.Eval(ctx, e.start, e.end)
		return toBool(v, "assert condition isn't a boolean expression")
	}
	msg := sourceOf(p.ipt, src)
	p.stk = append(p.stk, bpl.Assert(expr, msg))
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnRead() {

	e := p.popExpr()
	stk := p.stk
	i := len(stk) - 1
	n := func(ctx *bpl.Context) int {
		v := p.Eval(ctx, e.start, e.end)
		return toInt(v, "read bytes isn't an integer expression")
	}
	stk[i] = bpl.Read(n, stk[i].(bpl.Ruler))
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnIf() {

	e := p.popExpr()
	stk := p.stk
	i := len(stk) - 1
	cond := func(ctx *bpl.Context) bool {
		v := p.Eval(ctx, e.start, e.end)
		return toBool(v, "if condition isn't a boolean expression")
	}
	stk[i] = bpl.If(cond, stk[i].(bpl.Ruler))
}

/*
const (
	lzwArgMsg = "lzw argument isn't an integer expression"
)

func (p *Compiler) fnLzw() {

	e2 := p.popExpr()
	e1 := p.popExpr()
	stk := p.stk
	i := len(stk) - 1
	r := stk[i].(bpl.Ruler)
	dynR := func(ctx *bpl.Context) (bpl.Ruler, error) {
		v1 := p.Eval(ctx, e1.start, e1.end)
		v2 := p.Eval(ctx, e2.start, e2.end)
		return lzw.Type(toInt(v1, lzwArgMsg), toInt(v2, lzwArgMsg), r), nil
	}
	stk[i] = bpl.Dyntype(dynR)
}
*/

// -----------------------------------------------------------------------------

func (p *Compiler) dostruct(nDynExpr int, m int, cstyle int) {

	dynExprR := p.popRules(nDynExpr)
	if m == 0 {
		if dynExprR == nil {
			dynExprR = bpl.Nil
		}
		p.stk = append(p.stk, dynExprR)
		return
	}

	stk := p.stk
	base := len(stk) - (m << 1)
	members := make([]bpl.Member, m)
	for i := 0; i < m; i++ {
		idx := base + (i << 1)
		typ := stk[idx+1-cstyle].(bpl.Ruler)
		name := stk[idx+cstyle].(string)
		members[i] = bpl.Member{Name: name, Type: typ}
	}
	stk[base] = bpl.Struct(members, dynExprR)
	p.stk = stk[:base+1]
}

func (p *Compiler) cstruct() {

	nDynExpr := p.popArity()
	m := p.popArity()
	p.dostruct(nDynExpr, m, 1)
}

func (p *Compiler) gostruct() {

	nDynExpr := p.popArity()
	m := p.popArity()
	p.dostruct(nDynExpr, m, 0)
}

// -----------------------------------------------------------------------------
