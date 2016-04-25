package bpl

import (
	"qiniupkg.com/text/bpl.v1"
	"qlang.io/exec.v2"
)

// -----------------------------------------------------------------------------

type executor struct {
	code exec.Code
	ctx  *exec.Context
}

func newExecutor() *executor {

	ctx := exec.NewContext()
	p := &executor{ctx: ctx}
	ctx.Stack = exec.NewStack()
	ctx.Code = &p.code
	return p
}

func (p *executor) Eval(start, end int) interface{} {

	ctx := p.ctx
	stk := ctx.Stack
	p.code.Exec(start, end, stk, ctx)
	v, _ := stk.Pop()
	return v
}

// -----------------------------------------------------------------------------

type indexExpr struct {
	start int
	end   int
}

func (p *Compiler) istart() {

	p.idxStart = p.code.Len()
}

func (p *Compiler) iend() {

	end := p.code.Len()
	p.gstk.Push(&indexExpr{start: p.idxStart, end: end})
}

func (p *Compiler) popIndex() *indexExpr {

	if v, ok := p.gstk.Pop(); ok {
		if e, ok := v.(*indexExpr); ok {
			return e
		}
	}
	panic("no index expression")
}

func (p *Compiler) array() {

	e := p.popIndex()
	stk := p.stk
	i := len(stk) - 1
	n := func(ctx *bpl.Context) int {
		v := p.Eval(e.start, e.end)
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
