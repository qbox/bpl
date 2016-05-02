package bpl

import (
	"bytes"
	"fmt"
	"strings"

	"qiniupkg.com/text/bpl.ext.v1/lzw"
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

func (p *executor) Eval(ctx *bpl.Context, start, end int) interface{} {

	vars, hasDom := ctx.Dom().(map[string]interface{})
	if vars == nil {
		vars = make(map[string]interface{})
	}
	code := &p.code
	stk := p.estk
	var parent *exec.Context
	if len(ctx.Globals) > 0 {
		parent = exec.NewSimpleContext(ctx.Globals, nil, nil, nil)
	}
	ectx := exec.NewSimpleContext(vars, stk, code, parent)
	code.Exec(start, end, stk, ectx)
	if !hasDom && len(vars) > 0 { // update dom
		ctx.SetDom(vars)
	}
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

func (p *Compiler) fnIf() {

	var elseR bpl.Ruler
	if p.popArity() != 0 {
		elseR = p.popRule()
	} else {
		elseR = bpl.Nil
	}

	arity := p.popArity() + 1

	stk := p.stk
	n := len(stk)
	bodyRs := clone(stk[n-arity:])
	condExprs := p.gstk.PopNArgs(arity)
	r := func(ctx *bpl.Context) (bpl.Ruler, error) {
		for i := 0; i < arity; i++ {
			e := condExprs[i].(*exprBlock)
			v := p.Eval(ctx, e.start, e.end)
			if toBool(v, "condition isn't a boolean expression") {
				return bodyRs[i], nil
			}
		}
		return elseR, nil
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

func (p *Compiler) fnDo() {

	e := p.popExpr()
	fn := func(ctx *bpl.Context) error {
		p.Eval(ctx, e.start, e.end)
		return nil
	}
	p.stk = append(p.stk, bpl.Do(fn))
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnLet() {

	e := p.popExpr()
	stk := p.stk
	i := len(stk) - 1
	name := stk[i].(string)
	fn := func(ctx *bpl.Context) error {
		v := p.Eval(ctx, e.start, e.end)
		ctx.SetVar(name, v)
		return nil
	}
	stk[i] = bpl.Do(fn)
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnGlobal() {

	e := p.popExpr()
	stk := p.stk
	i := len(stk) - 1
	name := stk[i].(string)
	fn := func(ctx *bpl.Context) error {
		v := p.Eval(ctx, e.start, e.end)
		ctx.Globals[name] = v
		return nil
	}
	stk[i] = bpl.Do(fn)
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

const (
	lzwArgMsg = "lzw argument isn't an integer expression"
)

func (p *Compiler) fnLzw() {

	e3 := p.popExpr()
	e2 := p.popExpr()
	e1 := p.popExpr()
	stk := p.stk
	i := len(stk) - 1
	r := stk[i].(bpl.Ruler)
	dynR := func(ctx *bpl.Context) (bpl.Ruler, error) {
		v1, ok1 := p.Eval(ctx, e1.start, e1.end).([]byte)
		if !ok1 {
			panic("lzw source, order, litWidth: source isn't a []byte expression")
		}
		v2 := p.Eval(ctx, e2.start, e2.end)
		v3 := p.Eval(ctx, e3.start, e3.end)
		return lzw.Type(bytes.NewReader(v1), toInt(v2, lzwArgMsg), toInt(v3, lzwArgMsg), r), nil
	}
	stk[i] = bpl.Dyntype(dynR)
}

// -----------------------------------------------------------------------------

func (p *Compiler) doret() func(ctx *bpl.Context) (v interface{}, err error) {

	if p.popArity() != 0 {
		e := p.popExpr()
		return func(ctx *bpl.Context) (v interface{}, err error) {
			v = p.Eval(ctx, e.start, e.end)
			return
		}
	}
	return nil
}

func (p *Compiler) dostruct(nDynExpr int, m int, fnRet func(ctx *bpl.Context) (v interface{}, err error), cstyle int) {

	dynExprR := p.popRules(nDynExpr)
	stk := p.stk
	if m > 0 {
		base := len(stk) - (m << 1)
		members := make([]bpl.Member, m)
		for i := 0; i < m; i++ {
			idx := base + (i << 1)
			typ := stk[idx+1-cstyle].(bpl.Ruler)
			name := stk[idx+cstyle].(string)
			members[i] = bpl.Member{Name: name, Type: typ}
		}
		stk[base] = bpl.StructEx(members, dynExprR, fnRet)
		p.stk = stk[:base+1]
	} else {
		p.stk = append(stk, bpl.StructEx(nil, dynExprR, fnRet))
	}
}

func (p *Compiler) cstruct() {

	fnRet := p.doret()
	nDynExpr := p.popArity()
	m := p.popArity()
	p.dostruct(nDynExpr, m, fnRet, 1)
}

func (p *Compiler) gostruct() {

	fnRet := p.doret()
	nDynExpr := p.popArity()
	m := p.popArity()
	p.dostruct(nDynExpr, m, fnRet, 0)
}

// -----------------------------------------------------------------------------
