package lzw

import (
	"compress/lzw"
	"io"

	"qiniupkg.com/text/bpl.v1"
	"qiniupkg.com/text/bpl.v1/bufio"
)

// -----------------------------------------------------------------------------

type typeImpl struct {
	r        bpl.Ruler
	src      func(ctx *bpl.Context) io.Reader
	order    lzw.Order
	litWidth int
}

func (p typeImpl) Match(in *bufio.Reader, ctx *bpl.Context) (v interface{}, err error) {

	f := lzw.NewReader(p.src(ctx), p.order, p.litWidth)
	defer f.Close()

	in = bufio.NewReader(f)
	return p.r.Match(in, ctx)
}

func (p typeImpl) SizeOf() int {

	return -1
}

// Type is a matching unit that matches R with a lzw decoded stream.
//
func Type(src func(ctx *bpl.Context) io.Reader, order, litWidth int, R bpl.Ruler) bpl.Ruler {

	return &typeImpl{src: src, order: lzw.Order(order), litWidth: litWidth, r: R}
}

// -----------------------------------------------------------------------------
