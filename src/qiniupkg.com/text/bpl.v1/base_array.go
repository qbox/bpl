package bpl

import (
	"errors"
	"io"
	"reflect"
	"unsafe"

	"qiniupkg.com/text/bpl.v1/bufio"
)

// -----------------------------------------------------------------------------

func matchBaseArray(R BaseType, n int, in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	if n == 0 {
		return
	}

	t := baseTypes[R]
	v = t.newn(n)
	data := (*reflect.SliceHeader)(unsafe.Pointer(reflect.ValueOf(v).UnsafeAddr())).Data
	b := (*[1 << 30]byte)(unsafe.Pointer(data))
	_, err = io.ReadFull(in, b[:n*t.sizeOf])
	return
}

// -----------------------------------------------------------------------------

type baseArray struct {
	r BaseType
	n int
}

func (p *baseArray) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	n := p.n
	return matchBaseArray(p.r, n, in, ctx)
}

func (p *baseArray) SizeOf() int {

	return p.r.SizeOf() * p.n
}

// BaseArray returns a matching unit that matches R n times.
//
func BaseArray(r BaseType, n int) Ruler {

	return &baseArray{r: r, n: n}
}

// -----------------------------------------------------------------------------

type baseDynarray struct {
	r BaseType
	n func(ctx *Context) int
}

func (p *baseDynarray) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	defer func() {
		if e := recover(); e != nil {
			switch v := e.(type) {
			case string:
				err = errors.New(v)
			case error:
				err = v
			default:
				panic(e)
			}
		}
	}()

	n := p.n(ctx)
	return matchBaseArray(p.r, n, in, ctx)
}

func (p *baseDynarray) SizeOf() int {

	return -1
}

// BaseDynarray returns a matching unit that matches R n(ctx) times.
//
func BaseDynarray(r BaseType, n func(ctx *Context) int) Ruler {

	return &baseDynarray{r: r, n: n}
}

// -----------------------------------------------------------------------------
