package bpl

import (
	"io"
	"reflect"

	"qiniupkg.com/text/bpl.v1/bufio"
)

// -----------------------------------------------------------------------------

var (
	typeIntf = reflect.TypeOf((*interface{})(nil)).Elem()
)

func typeOf(v interface{}) reflect.Type {

	if v != nil {
		return reflect.TypeOf(v)
	}
	return typeIntf
}

func matchArray1(R Ruler, in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	v1, err := R.Match(in, ctx)
	if err != nil {
		return
	}

	t := typeOf(v1)
	ret := reflect.MakeSlice(reflect.SliceOf(t), 0, 4)
	ret = reflect.Append(ret, reflect.ValueOf(v1))
	for {
		_, err = in.Peek(1)
		if err != nil {
			if err == io.EOF {
				return ret.Interface(), nil
			}
			return
		}
		v1, err = R.Match(in, ctx)
		if err != nil {
			return
		}
		ret = reflect.Append(ret, reflect.ValueOf(v1))
	}
}

func matchArray(R Ruler, n int, in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	if n == 0 {
		return
	}

	v1, err := R.Match(in, ctx)
	if err != nil {
		return
	}

	t := typeOf(v1)
	ret := reflect.MakeSlice(reflect.SliceOf(t), 0, n)
	ret = reflect.Append(ret, reflect.ValueOf(v1))
	for i := 1; i < n; i++ {
		v1, err = R.Match(in, ctx)
		if err != nil {
			return
		}
		ret = reflect.Append(ret, reflect.ValueOf(v1))
	}
	return ret.Interface(), nil
}

// -----------------------------------------------------------------------------

type array1 struct {
	r Ruler
}

func (p *array1) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	_, err = in.Peek(1)
	if err != nil {
		return
	}
	return matchArray1(p.r, in, ctx)
}

func (p *array1) SizeOf() int {

	return -1
}

// Array1 returns a matching unit that matches R+
//
func Array1(R Ruler) Ruler {

	return &array1{r: R}
}

// -----------------------------------------------------------------------------

type array0 struct {
	r Ruler
}

func (p *array0) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	_, err = in.Peek(1)
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return
	}
	return matchArray1(p.r, in, ctx)
}

func (p *array0) SizeOf() int {

	return -1
}

// Array0 returns a matching unit that matches R*
//
func Array0(R Ruler) Ruler {

	return &array0{r: R}
}

// Array01 returns a matching unit that matches R?
//
func Array01(R Ruler) Ruler {

	return Repeat01(R)
}

// -----------------------------------------------------------------------------

type array struct {
	r Ruler
	n int
}

func (p *array) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	n := p.n
	return matchArray(p.r, n, in, ctx)
}

func (p *array) SizeOf() int {

	size := p.r.SizeOf()
	if size < 0 {
		return -1
	}
	return p.n * size
}

// Array returns a matching unit that matches R n times.
//
func Array(r Ruler, n int) Ruler {

	return &array{r: r, n: n}
}

// -----------------------------------------------------------------------------

type dynarray struct {
	r Ruler
	n func(ctx *Context) int
}

func (p *dynarray) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	n := p.n(ctx)
	return matchArray(p.r, n, in, ctx)
}

func (p *dynarray) SizeOf() int {

	return -1
}

// Dynarray returns a matching unit that matches R n(ctx) times.
//
func Dynarray(r Ruler, n func(ctx *Context) int) Ruler {

	return &dynarray{r: r, n: n}
}

// -----------------------------------------------------------------------------
