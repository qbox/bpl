package bpl

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
	stk[i] = bpl.Array(stk[i].(bpl.Ruler), p.code, e.start, p.end)
}

// -----------------------------------------------------------------------------
