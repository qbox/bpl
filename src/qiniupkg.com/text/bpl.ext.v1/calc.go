package bpl

import (
	"reflect"
	"strconv"

	"qlang.io/exec.v2"
	"qlang.io/qlang.spec.v1"
)

// -----------------------------------------------------------------------------

func castInt(a interface{}) (int, bool) {

	switch a1 := a.(type) {
	case int:
		return a1, true
	case int32:
		return int(a1), true
	case int64:
		return int(a1), true
	case int16:
		return int(a1), true
	case int8:
		return int(a1), true
	case uint:
		return int(a1), true
	case uint32:
		return int(a1), true
	case uint64:
		return int(a1), true
	case uint16:
		return int(a1), true
	case uint8:
		return int(a1), true
	}
	return 0, false
}

func toInt(a interface{}, msg string) int {

	if v, ok := castInt(a); ok {
		return v
	}
	panic(msg)
}

func toBool(a interface{}, msg string) bool {

	if v, ok := a.(bool); ok {
		return v
	}
	if v, ok := castInt(a); ok {
		return v != 0
	}
	panic(msg)
}

// CallFn generates a function call instruction. It is required by tpl.Interpreter engine.
//
func (p *Compiler) CallFn(fn interface{}) {

	p.code.Block(exec.Call(fn))
}

func eq(a, b interface{}) bool {

	if a1, ok := castInt(a); ok {
		switch b1 := b.(type) {
		case int:
			return a1 == b1
		}
	}
	panicUnsupportedOp2("==", a, b)
	return false
}

func and(a, b bool) bool {

	return a && b
}

func or(a, b bool) bool {

	return a || b
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

	return p.popConstInt()
}

func (p *Compiler) popConstInt() int {

	if v, ok := p.gstk.Pop(); ok {
		if val, ok := v.(int); ok {
			return val
		}
	}
	panic("no int")
}

func (p *Compiler) arity(arity int) {

	p.gstk.Push(arity)
}

func (p *Compiler) call() {

	arity := p.popArity()
	p.code.Block(exec.CallFn(arity))
}

func (p *Compiler) ref(name string) {

	var instr exec.Instr
	if v, ok := p.consts[name]; ok {
		instr = exec.Push(v)
	} else {
		instr = exec.Ref(name)
	}
	p.code.Block(instr)
}

func (p *Compiler) mref(name string) {

	p.code.Block(exec.MemberRef(name))
}

func (p *Compiler) pushi(v int) {

	p.code.Block(exec.Push(v))
}

func (p *Compiler) pushs(lit string) {

	v, err := strconv.Unquote(lit)
	if err != nil {
		panic("invalid string `" + lit + "`: " + err.Error())
	}
	p.code.Block(exec.Push(v))
}

func (p *Compiler) cpushi(v int) {

	p.gstk.Push(v)
}

func (p *Compiler) fnConst(name string) {

	p.consts[name] = p.popConstInt()
}

// -----------------------------------------------------------------------------

func (p *Compiler) fnMap() {

	arity := p.popArity()
	p.code.Block(exec.Call(qlang.MapFrom, arity*2))
}

func (p *Compiler) fnSlice() {

	arity := p.popArity()
	p.code.Block(exec.SliceFrom(arity))
}

// -----------------------------------------------------------------------------
