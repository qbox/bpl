package bpl

import (
	"io"
	"reflect"

	"qiniupkg.com/text/bpl.v1/bufio"
)

// -----------------------------------------------------------------------------

type repeat0 struct {
	r Ruler
}

func (p *repeat0) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	_, err = in.Peek(1)
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return
	}
	return repeat(p.r, in, ctx)
}

func repeat(R Ruler, in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	v1, err := R.Match(in, ctx)
	if err != nil {
		return
	}

	t := reflect.TypeOf(v1)
	ret := reflect.MakeSlice(reflect.SliceOf(t), 1, 4)
	ret = reflect.Append(ret, reflect.ValueOf(v1))
	for {
		v1, err = R.Match(in, ctx)
		if err != nil {
			if err == io.EOF {
				return ret.Interface(), nil
			}
			return
		}
		ret = reflect.Append(ret, reflect.ValueOf(v1))
	}
}

func (p *repeat0) SizeOf() int {

	return -1
}

// Repeat0 returns a matching unit that matches R*
//
func Repeat0(R Ruler) Ruler {

	return &repeat0{r: R}
}

// -----------------------------------------------------------------------------

type repeat1 struct {
	r Ruler
}

func (p *repeat1) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	_, err = in.Peek(1)
	if err != nil {
		return
	}
	return repeat(p.r, in, ctx)
}

func (p *repeat1) SizeOf() int {

	return -1
}

// Repeat1 returns a matching unit that matches R+
//
func Repeat1(R Ruler) Ruler {

	return &repeat1{r: R}
}

// -----------------------------------------------------------------------------

type repeat01 struct {
	r Ruler
}

func (p *repeat01) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	_, err = in.Peek(1)
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return
	}
	return p.r.Match(in, ctx)
}

func (p *repeat01) SizeOf() int {

	return -1
}

// Repeat01 returns a matching unit that matches R?
//
func Repeat01(R Ruler) Ruler {

	return nil
}

// -----------------------------------------------------------------------------
