package bpl

import (
	"errors"
	"io"
	"reflect"

	"qiniupkg.com/text/bpl.v1/bufio"
)

var (
	// ErrVarNotAssigned is returned when TypeVar.Elem is not assigned.
	ErrVarNotAssigned = errors.New("variable is not assigned")

	// ErrVarAssigned is returned when TypeVar.Elem is already assigned.
	ErrVarAssigned = errors.New("variable is already assigned")
)

// -----------------------------------------------------------------------------

type and struct {
	rs []Ruler
}

func (p *and) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	for _, r := range p.rs {
		_, err = r.Match(in, ctx)
		if err != nil {
			return
		}
	}
	return
}

func (p *and) SizeOf() int {

	return -1
}

// And returns a matching unit that matches R1 R2 ... RN
//
func And(rs ...Ruler) Ruler {

	return &and{rs: rs}
}

// -----------------------------------------------------------------------------

type seq struct {
	rs []Ruler
}

func (p *seq) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	if ctx == nil {
		ctx = NewContext()
	}

	ret := ctx.RequireVarSlice()
	for _, r := range p.rs {
		v, err = r.Match(in, NewSubContext(ctx))
		if err != nil {
			return
		}
		ret = append(ret, v)
	}
	ctx.dom = ret
	return ret, nil
}

func (p *seq) SizeOf() int {

	return -1
}

// Seq returns a matching unit that matches R1 R2 ... RN and returns matching result.
//
func Seq(rs ...Ruler) Ruler {

	return &seq{rs: rs}
}

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

// A TypeVar is typeinfo of a `Struct` member.
//
type TypeVar struct {
	Name string
	Elem Ruler
}

// Assign assigns TypeVar.Elem.
//
func (p *TypeVar) Assign(r Ruler) error {

	if p.Elem != nil {
		return ErrVarAssigned
	}
	p.Elem = r
	return nil
}

// Match is required by a matching unit. see Ruler interface.
//
func (p *TypeVar) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	r := p.Elem
	if r == nil {
		return 0, ErrVarNotAssigned
	}
	return r.Match(in, ctx)
}

// SizeOf is required by a matching unit. see Ruler interface.
//
func (p *TypeVar) SizeOf() int {

	return p.Elem.SizeOf()
}

// -----------------------------------------------------------------------------
