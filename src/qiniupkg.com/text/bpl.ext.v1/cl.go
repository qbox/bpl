package bpl

import (
	"errors"
	"fmt"

	"qiniupkg.com/text/bpl.ext.v1/bson"
	"qiniupkg.com/text/bpl.v1"
	"qiniupkg.com/text/tpl.v1/interpreter.util"
)

const grammar = `

expr = +factor/And

doc = +((IDENT '=' expr ';')/assign)

factor =
    IDENT/ident |
    '*' factor/repeat0 |
    '+' factor/repeat1 |
    '?' factor/repeat01 |
    '(' expr ')'
`

var (
	// ErrNoDoc is returned when `doc` is undefined.
	ErrNoDoc = errors.New("no doc")
)

// -----------------------------------------------------------------------------

// A Compiler compiles bpl source code to matching units.
//
type Compiler struct {
	stk    []bpl.Ruler
	rulers map[string]bpl.Ruler
	vars   map[string]*bpl.TypeVar
}

// NewCompiler returns a bpl compiler.
//
func NewCompiler() (p *Compiler) {

	rulers := make(map[string]bpl.Ruler)
	vars := make(map[string]*bpl.TypeVar)
	return &Compiler{rulers: rulers, vars: vars}
}

// Ret returns compiling result.
//
func (p *Compiler) Ret() (r bpl.Ruler, err error) {

	root, ok := p.rulers["doc"]
	if !ok {
		if v, ok := p.vars["doc"]; ok {
			root = v.Elem
		} else {
			return nil, ErrNoDoc
		}
	}
	for name, v := range p.vars {
		if v.Elem == nil {
			err = fmt.Errorf("variable `%s` is not assigned", name)
			return
		}
	}
	return root, nil
}

// Grammar returns the qlang compiler's grammar. It is required by tpl.Interpreter engine.
//
func (p *Compiler) Grammar() string {

	return grammar
}

// Fntable returns the qlang compiler's function table. It is required by tpl.Interpreter engine.
//
func (p *Compiler) Fntable() map[string]interface{} {

	return fntable
}

// Stack returns nil (no stack). It is required by tpl.Interpreter engine.
//
func (p *Compiler) Stack() interpreter.Stack {

	return nil
}

// -----------------------------------------------------------------------------

func clone(rs []bpl.Ruler) []bpl.Ruler {

	dest := make([]bpl.Ruler, len(rs))
	for i, r := range rs {
		dest[i] = r
	}
	return dest
}

func (p *Compiler) and(m int) {

	if m == 1 {
		return
	}
	stk := p.stk
	n := len(stk)
	stk[n-m] = bpl.And(clone(stk[n-m:])...)
	p.stk = stk[:n-m+1]
}

func (p *Compiler) ident(name string) {

	r, ok := p.rulers[name]
	if ok {
		r = &bpl.NamedType{Name: name, Type: r}
	} else if r, ok = p.vars[name]; !ok {
		if r, ok = builtins[name]; ok {
			p.rulers[name] = r
		} else {
			v := &bpl.TypeVar{Name: name}
			p.vars[name] = v
			r = v
		}
	}
	p.stk = append(p.stk, r)
}

func (p *Compiler) assign(name string) {

	a := p.stk[0]
	if v, ok := p.vars[name]; ok {
		if err := v.Assign(a); err != nil {
			panic(err)
		}
	} else if _, ok := p.rulers[name]; ok {
		panic("ruler already exists: " + name)
	} else {
		p.rulers[name] = a
	}
	p.stk = p.stk[:0]
}

func (p *Compiler) repeat0() {

	stk := p.stk
	i := len(stk) - 1
	stk[i] = bpl.Repeat0(stk[i])
}

func (p *Compiler) repeat1() {

	stk := p.stk
	i := len(stk) - 1
	stk[i] = bpl.Repeat1(stk[i])
}

func (p *Compiler) repeat01() {

	stk := p.stk
	i := len(stk) - 1
	stk[i] = bpl.Repeat01(stk[i])
}

// -----------------------------------------------------------------------------

var fntable = map[string]interface{}{
	"$And":      (*Compiler).and,
	"$ident":    (*Compiler).ident,
	"$assign":   (*Compiler).assign,
	"$repeat0":  (*Compiler).repeat0,
	"$repeat1":  (*Compiler).repeat1,
	"$repeat01": (*Compiler).repeat01,
}

var builtins = map[string]bpl.Ruler{
	"int8":    bpl.Int8,
	"int16":   bpl.Int16,
	"int32":   bpl.Int32,
	"int64":   bpl.Int64,
	"uint8":   bpl.Uint8,
	"uint16":  bpl.Uint16,
	"uint32":  bpl.Uint32,
	"uint64":  bpl.Uint64,
	"float32": bpl.Float32,
	"float64": bpl.Float64,
	"cstring": bpl.CString,
	"bson":    bson.Type,
}

// -----------------------------------------------------------------------------
