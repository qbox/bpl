package bpl

import (
	"fmt"

	"qiniupkg.com/text/bpl.v1/bufio"
)

// -----------------------------------------------------------------------------

// A Context represents the matching context of bpl.
//
type Context struct {
	dom    interface{}
	Parent *Context
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
	return &Context{Parent: p}
}

func (p *Context) requireVarSlice() []interface{} {

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
func (p *Context) Dom() interface{} {

	return p.dom
}

func domOf(p *Context) interface{} {

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
