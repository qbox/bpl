package bpl

import (
	"errors"
	"fmt"

	"qiniupkg.com/text/bpl.ext.v1/bson"
	"qiniupkg.com/text/bpl.v1"
	"qiniupkg.com/text/tpl.v1/interpreter.util"
	"qlang.io/exec.v2"
)

const grammar = `

expr = +factor/And

doc = +((IDENT '=' expr ';')/assign)

iterm = ifactor *('*' ifactor/mul | '/' ifactor/quo | '%' ifactor/mod)

iexpr = iterm *('+' iterm/add | '-' iterm/sub)

index = '['/istart iexpr ']'/iend

ctype = IDENT/ident ?(index/array | '*'/repeat0 | '?'/repeat01 | '+'/repeat1)

type =
	IDENT/ident |
	(index IDENT/ident)/array |
	('*'! IDENT/ident)/array0 |
	('?'! IDENT/ident)/array01 |
	('+'! IDENT/ident)/array1

casebody = (INT/casei ':' expr) %= ';'/ARITY ?(';' "default" ':' expr)/ARITY

caseexpr = "case"/istart! iexpr '{'/iend casebody ?';' '}' /case

readexpr = "read"/istart! iexpr "do"/iend expr /read

dynexpr = caseexpr | readexpr

cstruct = (ctype IDENT/var) %= ';'/ARITY ?(';' dynexpr)/ARITY /cstruct

struct = (IDENT/var type) %= ';'/ARITY ?(';' dynexpr)/ARITY /struct

factor =
	IDENT/ident |
	'{' ('/' "C" ';' cstruct | struct) ?';' '}' |
	'*' factor/repeat0 |
	'+' factor/repeat1 |
	'?' factor/repeat01 |
	'(' expr ')' |
	'[' +factor/Seq ']' |
	dynexpr

atom = 
	'(' iexpr %= ','/ARITY ')'/call |
	'.' IDENT/mref

ifactor =
	INT/pushi |
	(IDENT/ref | '('! iexpr ')') *atom |
	"sizeof"! '(' IDENT/sizeof ')' |
	'-' ifactor/neg |
	'+' ifactor
`

var (
	// ErrNoDoc is returned when `doc` is undefined.
	ErrNoDoc = errors.New("no doc")
)

// -----------------------------------------------------------------------------

// A Compiler compiles bpl source code to matching units.
//
type Compiler struct {
	*executor
	stk      []interface{}
	rulers   map[string]bpl.Ruler
	vars     map[string]*bpl.TypeVar
	gstk     exec.Stack
	idxStart int
}

// NewCompiler returns a bpl compiler.
//
func NewCompiler() (p *Compiler) {

	rulers := make(map[string]bpl.Ruler)
	vars := make(map[string]*bpl.TypeVar)
	e := newExecutor()
	return &Compiler{rulers: rulers, vars: vars, executor: e}
}

// Ret returns compiling result.
//
func (p *Compiler) Ret() (r Ruler, err error) {

	root, ok := p.rulers["doc"]
	if !ok {
		if v, ok := p.vars["doc"]; ok {
			root = v.Elem
		} else {
			return Ruler{Ruler: nil}, ErrNoDoc
		}
	}
	for name, v := range p.vars {
		if v.Elem == nil {
			err = fmt.Errorf("variable `%s` is not assigned", name)
			return
		}
	}
	return Ruler{Ruler: root}, nil
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

func clone(rs []interface{}) []bpl.Ruler {

	dest := make([]bpl.Ruler, len(rs))
	for i, r := range rs {
		dest[i] = r.(bpl.Ruler)
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

func (p *Compiler) seq(m int) {

	stk := p.stk
	n := len(stk)
	stk[n-m] = bpl.Seq(clone(stk[n-m:])...)
	p.stk = stk[:n-m+1]
}

func (p *Compiler) variable(name string) {

	p.stk = append(p.stk, name)
}

func (p *Compiler) ruleOf(name string) (r bpl.Ruler, ok bool) {

	r, ok = p.rulers[name]
	if !ok {
		if r, ok = p.vars[name]; !ok {
			if r, ok = builtins[name]; ok {
				p.rulers[name] = r
			}
		}
	}
	return
}

func (p *Compiler) sizeof(name string) {

	r, ok := p.ruleOf(name)
	if !ok {
		panic(fmt.Errorf("sizeof error: type `%v` not found", name))
	}
	n := r.SizeOf()
	if n < 0 {
		panic(fmt.Errorf("sizeof error: type `%v` isn't a fixed size type", name))
	}
	p.code.Block(exec.Push(n))
}

func (p *Compiler) ident(name string) {

	r, ok := p.ruleOf(name)
	if !ok {
		v := &bpl.TypeVar{Name: name}
		p.vars[name] = v
		r = v
	}
	p.stk = append(p.stk, r)
}

func (p *Compiler) assign(name string) {

	a := p.stk[0].(bpl.Ruler)
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
	stk[i] = bpl.Repeat0(stk[i].(bpl.Ruler))
}

func (p *Compiler) repeat1() {

	stk := p.stk
	i := len(stk) - 1
	stk[i] = bpl.Repeat1(stk[i].(bpl.Ruler))
}

func (p *Compiler) repeat01() {

	stk := p.stk
	i := len(stk) - 1
	stk[i] = bpl.Repeat01(stk[i].(bpl.Ruler))
}

// -----------------------------------------------------------------------------

var fntable = map[string]interface{}{
	"$And":      (*Compiler).and,
	"$Seq":      (*Compiler).seq,
	"$istart":   (*Compiler).istart,
	"$iend":     (*Compiler).iend,
	"$array":    (*Compiler).array,
	"$array1":   (*Compiler).array1,
	"$array0":   (*Compiler).array0,
	"$array01":  (*Compiler).repeat01,
	"$var":      (*Compiler).variable,
	"$ident":    (*Compiler).ident,
	"$assign":   (*Compiler).assign,
	"$repeat0":  (*Compiler).repeat0,
	"$repeat1":  (*Compiler).repeat1,
	"$repeat01": (*Compiler).repeat01,

	"$mul":     mul,
	"$quo":     quo,
	"$mod":     mod,
	"$neg":     neg,
	"$add":     add,
	"$sub":     sub,
	"$sizeof":  (*Compiler).sizeof,
	"$ARITY":   (*Compiler).arity,
	"$call":    (*Compiler).call,
	"$ref":     (*Compiler).ref,
	"$mref":    (*Compiler).mref,
	"$pushi":   (*Compiler).pushi,
	"$read":    (*Compiler).fnRead,
	"$case":    (*Compiler).fnCase,
	"$casei":   (*Compiler).casei,
	"$cstruct": (*Compiler).cstruct,
	"$struct":  (*Compiler).gostruct,
}

var builtins = map[string]bpl.Ruler{
	"int8":    bpl.Int8,
	"int16":   bpl.Int16,
	"int32":   bpl.Int32,
	"int64":   bpl.Int64,
	"uint8":   bpl.Uint8,
	"byte":    bpl.Uint8,
	"uint16":  bpl.Uint16,
	"uint32":  bpl.Uint32,
	"uint64":  bpl.Uint64,
	"float32": bpl.Float32,
	"float64": bpl.Float64,
	"cstring": bpl.CString,
	"nil":     bpl.Nil,
	"eof":     bpl.EOF,
	"done":    bpl.Done,
	"bson":    bson.Type,
	"dump":    dump(0),
}

// -----------------------------------------------------------------------------
