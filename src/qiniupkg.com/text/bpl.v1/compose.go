package bpl

import (
	"errors"
	"io"
	"io/ioutil"

	"qiniupkg.com/text/bpl.v1/bufio"
)

var (
	// ErrVarNotAssigned is returned when TypeVar.Elem is not assigned.
	ErrVarNotAssigned = errors.New("variable is not assigned")

	// ErrVarAssigned is returned when TypeVar.Elem is already assigned.
	ErrVarAssigned = errors.New("variable is already assigned")

	// ErrNotEOF is returned when current position is not at EOF.
	ErrNotEOF = errors.New("current position is not at EOF")
)

// -----------------------------------------------------------------------------

type nilType int

func (p nilType) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	return nil, nil
}

func (p nilType) SizeOf() int {

	return 0
}

// Nil is a matching unit that matches zero bytes.
//
var Nil Ruler = nilType(0)

// -----------------------------------------------------------------------------

type eof int

func (p eof) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	_, err = in.Peek(1)
	if err == io.EOF {
		return nil, nil
	}
	return nil, ErrNotEOF
}

func (p eof) SizeOf() int {

	return 0
}

// EOF is a matching unit that matches EOF.
//
var EOF Ruler = eof(0)

// -----------------------------------------------------------------------------

type done int

func (p done) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	_, err = in.WriteTo(ioutil.Discard)
	return
}

func (p done) SizeOf() int {

	return -1
}

// Done is a matching unit that seeks current position to EOF.
//
var Done Ruler = done(0)

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
	return ctx.Dom(), nil
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

	ret := ctx.requireVarSlice()
	for _, r := range p.rs {
		v, err = r.Match(in, ctx.NewSub())
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

type dyntype struct {
	r func(ctx *Context) (Ruler, error)
}

func (p *dyntype) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	r, err := p.r(ctx)
	if err != nil {
		return
	}
	if r != nil {
		return r.Match(in, ctx)
	}
	return
}

func (p *dyntype) SizeOf() int {

	return -1
}

// Dyntype returns a dynamic matching unit.
//
func Dyntype(r func(ctx *Context) (Ruler, error)) Ruler {

	return &dyntype{r: r}
}

// -----------------------------------------------------------------------------

type read struct {
	n func(ctx *Context) int
	r Ruler
}

func (p *read) Match(in *bufio.Reader, ctx *Context) (v interface{}, err error) {

	n := p.n(ctx)
	b := make([]byte, n)
	_, err = io.ReadFull(in, b)
	if err != nil {
		return
	}
	in = bufio.NewReaderBuffer(b)
	return p.r.Match(in, ctx)
}

func (p *read) SizeOf() int {

	return -1
}

// Read returns a matching unit that reads n(ctx) bytes and matches R.
//
func Read(n func(ctx *Context) int, r Ruler) Ruler {

	return &read{r: r, n: n}
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
