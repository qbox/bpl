package bpl

import (
	"errors"

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

	ret := ctx.requireVarSlice()
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
