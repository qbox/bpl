package bpl

import (
	"bufio"
	"io"
	"reflect"
	"unsafe"
)

// -----------------------------------------------------------------------------

func matchCharArray(n int, in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	if n == 0 {
		return "", nil
	}

	b := make([]byte, n)
	_, err = io.ReadFull(in, b)
	if err != nil {
		return
	}
	return string(b), nil
}

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

func (p *baseArray) RetType() reflect.Type {

	return reflect.SliceOf(p.r.RetType())
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

	n := p.n(ctx)
	return matchBaseArray(p.r, n, in, ctx)
}

func (p *baseDynarray) RetType() reflect.Type {

	return reflect.SliceOf(p.r.RetType())
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

type charArray int

func (p charArray) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	return matchCharArray(int(p), in, ctx)
}

func (p charArray) RetType() reflect.Type {

	return tyString
}

func (p charArray) SizeOf() int {

	return int(p)
}

// CharArray returns a matching unit that matches `[n]char`.
//
func CharArray(n int) Ruler {

	return charArray(n)
}

// -----------------------------------------------------------------------------

type charDynarray func(ctx *Context) int

func (p charDynarray) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	n := p(ctx)
	return matchCharArray(n, in, ctx)
}

func (p charDynarray) RetType() reflect.Type {

	return tyString
}

func (p charDynarray) SizeOf() int {

	return -1
}

// CharDynarray returns a matching unit that matches `[n(ctx)]char`.
//
func CharDynarray(n func(ctx *Context) int) Ruler {

	return charDynarray(n)
}

// -----------------------------------------------------------------------------
