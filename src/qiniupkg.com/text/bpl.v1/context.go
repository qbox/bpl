package bpl

import (
	"bytes"

	"qiniupkg.com/text/bpl.v1/bufio"
)

// -----------------------------------------------------------------------------

// A Context represents the matching context of bpl.
//
type Context struct {
	vars map[string]interface{}

	// Capture is optional when matching.
	capt *bytes.Buffer
}

// NewContext returns a new Context.
//
func NewContext() *Context {

	vars := make(map[string]interface{})
	return &Context{vars: vars}
}

// CaptureIf captures the matching text if needed.
//
func (p *Context) CaptureIf(b []byte) {

	if p != nil && p.capt != nil {
		p.capt.Write(b)
	}
}

// SetVar sets a new variable to matching context.
//
func (p *Context) SetVar(name string, v interface{}) {

	p.vars[name] = v
}

// Var gets a variable from matching context.
//
func (p *Context) Var(name string) (v interface{}, ok bool) {

	v, ok = p.vars[name]
	return
}

// Vars returns all variables in matching context.
//
func (p *Context) Vars() map[string]interface{} {

	return p.vars
}

// -----------------------------------------------------------------------------

// A Ruler interface is required to a matching unit.
//
type Ruler interface {
	// Match matches input stream `in`, and returns matching result.
	Match(in *bufio.Reader, ctx *Context) (v interface{}, err error)

	// SizeOf returns expected length of result. If length is variadic, it returns -1.
	SizeOf() int
}

// -----------------------------------------------------------------------------
