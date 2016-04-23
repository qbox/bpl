package bpl

import (
	"bytes"
	"fmt"

	"qiniupkg.com/text/bpl.v1/bufio"
)

// -----------------------------------------------------------------------------

// A Context represents the matching context of bpl.
//
type Context struct {
	dom  interface{}
	capt *bytes.Buffer
}

// NewContext returns a new Context.
//
func NewContext() *Context {

	return &Context{}
}

// NewSubContext returns a new sub Context.
//
func NewSubContext(p *Context) *Context {

	if p == nil {
		return nil
	}
	return &Context{capt: p.capt}
}

// WithCapture returns the matching context with capture.
//
func (p *Context) WithCapture() *Context {

	p.capt = new(bytes.Buffer)
	return p
}

// CaptureIf captures the matching text if needed.
//
func (p *Context) CaptureIf(b []byte) {

	if p != nil && p.capt != nil {
		p.capt.Write(b)
	}
}

// RequireVarSlice verifies and returns matching result as []interface{}.
//
func (p *Context) RequireVarSlice() []interface{} {

	var vars []interface{}
	if p.dom == nil {
		vars = make([]interface{}, 0, 4)
	} else if domv, ok := p.dom.([]interface{}); ok {
		vars = domv
	} else {
		panic("dom type isn't []interface{}")
	}
	return vars
}

// SetVar sets a new variable to matching context.
//
func (p *Context) SetVar(name string, v interface{}) {

	var vars map[string]interface{}
	if p.dom == nil {
		vars = make(map[string]interface{})
		p.dom = vars
	} else if domv, ok := p.dom.(map[string]interface{}); ok {
		if _, ok = domv[name]; ok {
			panic(fmt.Errorf("variable `%s` exists in dom", name))
		}
		vars = domv
	} else {
		panic("dom type isn't map[string]interface{}")
	}
	vars[name] = v
}

// Var gets a variable from matching context.
//
func (p *Context) Var(name string) (v interface{}, ok bool) {

	vars, ok := p.dom.(map[string]interface{})
	if ok {
		v, ok = vars[name]
	} else {
		panic("dom type isn't map[string]interface{}")
	}
	return
}

// SetDom set matching result of matching result.
//
func (p *Context) SetDom(v interface{}) {

	if p.dom == nil {
		p.dom = v
	} else {
		panic("dom was assigned already")
	}
}

// Dom returns matching result.
//
func Dom(p *Context) interface{} {

	if p == nil {
		return nil
	}
	return p.dom
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
