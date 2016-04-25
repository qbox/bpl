package bpl

import (
	"reflect"

	"qlang.io/exec.v2"
)

// -----------------------------------------------------------------------------

// CallFn generates a function call instruction. It is required by tpl.Interpreter engine.
//
func (p *Compiler) CallFn(fn interface{}) {

	p.code.Block(exec.Call(fn))
}

func neg(a interface{}) interface{} {

	switch a1 := a.(type) {
	case int:
		return -a1
	}
	return panicUnsupportedOp1("-", a)
}

func mul(a, b interface{}) interface{} {

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 * b1
		}
	}
	return panicUnsupportedOp2("*", a, b)
}

func quo(a, b interface{}) interface{} {

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 / b1
		}
	}
	return panicUnsupportedOp2("/", a, b)
}

func mod(a, b interface{}) interface{} {

	if a1, ok := a.(int); ok {
		if b1, ok := b.(int); ok {
			return a1 % b1
		}
	}
	return panicUnsupportedOp2("%", a, b)
}

func add(a, b interface{}) interface{} {

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 + b1
		}
	}
	return panicUnsupportedOp2("+", a, b)
}

func sub(a, b interface{}) interface{} {

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 - b1
		}
	}
	return panicUnsupportedOp2("-", a, b)
}

func panicUnsupportedOp1(op string, a interface{}) interface{} {

	ta := typeString(a)
	panic("unsupported operator: " + op + ta)
}

func panicUnsupportedOp2(op string, a, b interface{}) interface{} {

	ta := typeString(a)
	tb := typeString(b)
	panic("unsupported operator: " + ta + op + tb)
}

func typeString(a interface{}) string {

	if a == nil {
		return "nil"
	}
	return reflect.TypeOf(a).String()
}

// -----------------------------------------------------------------------------

func (p *Compiler) popArity() int {

	if v, ok := p.gstk.Pop(); ok {
		if arity, ok := v.(int); ok {
			return arity
		}
	}
	panic("no arity")
}

func (p *Compiler) arity(arity int) {

	p.gstk.Push(arity)
}

func (p *Compiler) call() {

	arity := p.popArity()
	p.code.Block(exec.CallFn(arity))
}

func (p *Compiler) ref(name string) {

	p.code.Block(exec.Ref(name))
}

func (p *Compiler) mref(name string) {

	p.code.Block(exec.MemberRef(name))
}

func (p *Compiler) pushi(v int) {

	p.code.Block(exec.Push(v))
}

// -----------------------------------------------------------------------------
