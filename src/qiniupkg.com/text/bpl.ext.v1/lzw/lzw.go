package lzw

import (
	"compress/lzw"

	"qiniupkg.com/text/bpl.v1"
	"qiniupkg.com/text/bpl.v1/bufio"
)

// -----------------------------------------------------------------------------

type typeImpl struct {
	r        bpl.Ruler
	order    lzw.Order
	litWidth int
}

func (p typeImpl) Match(in *bufio.Reader, ctx *bpl.Context) (v interface{}, err error) {

	f := lzw.NewReader(in, p.order, p.litWidth)
	defer f.Close()

	in = bufio.NewReader(f)
	return p.r.Match(in, ctx)
}

func (p typeImpl) SizeOf() int {

	return -1
}

// Type is a matching unit that matches R with a lzw decoded stream.
//
func Type(order, litWidth int, R bpl.Ruler) bpl.Ruler {

	return &typeImpl{order: lzw.Order(order), litWidth: litWidth, r: R}
}

// -----------------------------------------------------------------------------
