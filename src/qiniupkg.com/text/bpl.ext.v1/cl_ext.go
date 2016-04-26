package bpl

import (
	"fmt"

	"qiniupkg.com/text/bpl.v1"
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

	var vars map[string]interface{}
	if ctx != nil {
		vars, _ = ctx.Dom().(map[string]interface{})
	}
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
		if ctx != nil {
			ctx = ctx.Parent
		}
		v := p.Eval(ctx, e.start, e.end)
		return v.(int)
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

func (p *Compiler) popRule() bpl.Ruler {

	n := len(p.stk) - 1
	r := p.stk[n].(bpl.Ruler)
	p.stk = p.stk[:n]
	return r
}

func (p *Compiler) fnCase() {

	var defaultR bpl.Ruler
	if p.popArity() != 0 {
		defaultR = p.popRule()
	}

	arity := p.popArity()

	stk := p.stk
	n := len(stk)
	caseRs := clone(stk[n-arity:])
	caseExprs := p.gstk.PopNArgs(arity)
	e := p.popExpr()
	r := func(ctx *bpl.Context) (bpl.Ruler, error) {
		v := p.Eval(ctx, e.start, e.end)
		for i, expr := range caseExprs {
			if eq(v, expr) {
				return caseRs[i], nil
			}
		}
		if defaultR != nil {
			return defaultR, nil
		}
		return nil, fmt.Errorf("case `%v` is not found", v)
	}
	stk[n-arity] = bpl.Dyntype(r)
	p.stk = stk[:n-arity+1]
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnRead() {

	e := p.popExpr()
	stk := p.stk
	i := len(stk) - 1
	n := func(ctx *bpl.Context) int {
		v := p.Eval(ctx, e.start, e.end)
		return v.(int)
	}
	stk[i] = bpl.Read(n, stk[i].(bpl.Ruler))
}

// -----------------------------------------------------------------------------

func (p *Compiler) dostruct(hasCase int, m int, cstyle int) {

	var caseR bpl.Ruler
	if hasCase != 0 {
		caseR = p.popRule()
	}

	if m == 0 {
		if caseR == nil {
			caseR = bpl.Nil
		}
		p.stk = append(p.stk, caseR)
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
	stk[base] = bpl.Struct(members, caseR)
	p.stk = stk[:base+1]
}

func (p *Compiler) cstruct() {

	hasCase := p.popArity()
	m := p.popArity()
	p.dostruct(hasCase, m, 1)
}

func (p *Compiler) gostruct() {

	hasCase := p.popArity()
	m := p.popArity()
	p.dostruct(hasCase, m, 0)
}

// -----------------------------------------------------------------------------
