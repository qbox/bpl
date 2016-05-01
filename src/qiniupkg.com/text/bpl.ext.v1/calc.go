package bpl

import (
	"reflect"
	"strconv"

	"qlang.io/exec.v2"
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

func neg(a interface{}) interface{} {

	switch a1 := a.(type) {
	case int:
		return -a1
	}
	return panicUnsupportedOp1("-", a)
}

func bitnot(a interface{}) interface{} {

	switch a1 := a.(type) {
	case int:
		return ^a1
	}
	return panicUnsupportedOp1("^", a)
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

func equ(a, b interface{}) bool {

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 == b1
		}
	case string:
		switch b1 := b.(type) {
		case string:
			return a1 == b1
		}
	}
	panicUnsupportedOp2("==", a, b)
	return false
}

func ne(a, b interface{}) bool {

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 != b1
		}
	case string:
		switch b1 := b.(type) {
		case string:
			return a1 != b1
		}
	}
	panicUnsupportedOp2("!=", a, b)
	return false
}

func le(a, b interface{}) bool {

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 <= b1
		}
	case string:
		switch b1 := b.(type) {
		case string:
			return a1 <= b1
		}
	}
	panicUnsupportedOp2("<=", a, b)
	return false
}

func lt(a, b interface{}) bool {

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 < b1
		}
	case string:
		switch b1 := b.(type) {
		case string:
			return a1 < b1
		}
	}
	panicUnsupportedOp2("<", a, b)
	return false
}

func ge(a, b interface{}) bool {

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 >= b1
		}
	case string:
		switch b1 := b.(type) {
		case string:
			return a1 >= b1
		}
	}
	panicUnsupportedOp2(">=", a, b)
	return false
}

func gt(a, b interface{}) bool {

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 > b1
		}
	case string:
		switch b1 := b.(type) {
		case string:
			return a1 > b1
		}
	}
	panicUnsupportedOp2(">", a, b)
	return false
}

func and(a, b bool) bool {

	return a && b
}

func or(a, b bool) bool {

	return a || b
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

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
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

func lshr(a, b interface{}) interface{} { // a << b

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 << uint(b1)
		}
	}
	return panicUnsupportedOp2("<<", a, b)
}

func rshr(a, b interface{}) interface{} { // a >> b

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 >> uint(b1)
		}
	}
	return panicUnsupportedOp2(">>", a, b)
}

func bitand(a, b interface{}) interface{} { // a & b

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 & b1
		}
	}
	return panicUnsupportedOp2("&", a, b)
}

func bitor(a, b interface{}) interface{} { // a | b

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 | b1
		}
	}
	return panicUnsupportedOp2("|", a, b)
}

func xor(a, b interface{}) interface{} { // a ^ b

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 ^ b1
		}
	}
	return panicUnsupportedOp2("^", a, b)
}

func andnot(a, b interface{}) interface{} { // a &^ b

	switch a1 := a.(type) {
	case int:
		switch b1 := b.(type) {
		case int:
			return a1 &^ b1
		}
	}
	return panicUnsupportedOp2("&^", a, b)
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
