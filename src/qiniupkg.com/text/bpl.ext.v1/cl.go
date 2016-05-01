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

term1 = ifactor *(
	'*' ifactor/mul | '/' ifactor/quo | '%' ifactor/mod |
	"<<" ifactor/lshr | ">>" ifactor/rshr | '&' ifactor/bitand | "&^" ifactor/andnot)

term2 = term1 *('+' term1/add | '-' term1/sub)

term3 = term2 *('<' term2/lt | '>' term2/gt | "==" term2/eq | "<=" term2/le | ">=" term2/ge | "!=" term2/ne)

term4 = term3 *("&&" term3/and)

iexpr = term4 *("||" term4/or)

index = '['/istart iexpr ']'/iend

ctype = IDENT/ident ?(index/array | '*'/array0 | '?'/array01 | '+'/array1)

type =
	IDENT/ident |
	(index IDENT/ident)/array |
	('*'! IDENT/ident)/array0 |
	('?'! IDENT/ident)/array01 |
	('+'! IDENT/ident)/array1

casebody = (INT/casei ':' expr/source) %= ';'/ARITY ?(';' "default" ':' expr)/ARITY

caseexpr = "case"/istart! iexpr/source '{'/iend casebody ?';' '}' /case

exprblock = true/istart! iexpr (@'{' | "do")/iend expr

ifexpr = "if" exprblock *("elif" exprblock)/ARITY ?("else"! expr)/ARITY /if

readexpr = "read" exprblock /read

evalexpr = "eval" exprblock /eval

assertexpr = ("assert"/istart! iexpr /iend) /assert

letexpr = "let"! IDENT/var '='/istart! iexpr /iend /let

lzwexpr = "lzw"/istart! iexpr /iend ',' /istart! iexpr /iend ',' /istart! iexpr /iend exprblock /lzw

dynexpr = (caseexpr | readexpr | evalexpr | assertexpr | ifexpr | letexpr | lzwexpr)/xline

retexpr = ?';' "return"/istart! iexpr /iend

cstruct = (ctype IDENT/var) %= ';'/ARITY ?';' dynexpr %= ';'/ARITY ?retexpr/ARITY /cstruct

struct = (IDENT/var type) %= ';'/ARITY ?';' dynexpr %= ';'/ARITY ?retexpr/ARITY /struct

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
	STRING/pushs |
	(IDENT/ref | '('! iexpr ')') *atom |
	"sizeof"! '(' IDENT/sizeof ')' |
	'^' ifactor/bitnot |
	'-' ifactor/neg |
	'+' ifactor

cexpr = INT/cpushi

const = (IDENT '=' cexpr ';')/const

doc = +(
	(IDENT '=' expr ';')/assign |
	"const" '(' *const ')' ';')
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
	consts   map[string]interface{}
	gstk     exec.Stack
	ipt      interpreter.Engine
	idxStart int
}

func newCompiler() (p *Compiler) {

	rulers := make(map[string]bpl.Ruler)
	vars := make(map[string]*bpl.TypeVar)
	consts := make(map[string]interface{})
	e := newExecutor()
	return &Compiler{rulers: rulers, vars: vars, consts: consts, executor: e}
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
	"$lt":      lt,
	"$gt":      gt,
	"$eq":      equ,
	"$le":      le,
	"$ge":      ge,
	"$ne":      ne,
	"$and":     and,
	"$or":      or,
	"$lshr":    lshr,
	"$rshr":    rshr,
	"$bitand":  bitand,
	"$bitnot":  bitnot,
	"$andnot":  andnot,
	"$sizeof":  (*Compiler).sizeof,
	"$ARITY":   (*Compiler).arity,
	"$call":    (*Compiler).call,
	"$ref":     (*Compiler).ref,
	"$mref":    (*Compiler).mref,
	"$pushi":   (*Compiler).pushi,
	"$pushs":   (*Compiler).pushs,
	"$cpushi":  (*Compiler).cpushi,
	"$let":     (*Compiler).fnLet,
	"$eval":    (*Compiler).fnEval,
	"$if":      (*Compiler).fnIf,
	"$read":    (*Compiler).fnRead,
	"$lzw":     (*Compiler).fnLzw,
	"$case":    (*Compiler).fnCase,
	"$assert":  (*Compiler).fnAssert,
	"$const":   (*Compiler).fnConst,
	"$casei":   (*Compiler).casei,
	"$source":  (*Compiler).source,
	"$cstruct": (*Compiler).cstruct,
	"$struct":  (*Compiler).gostruct,
	"$xline":   (*Compiler).xline,
}

var builtins = map[string]bpl.Ruler{
	"int8":    bpl.Int8,
	"int16":   bpl.Int16,
	"int32":   bpl.Int32,
	"int64":   bpl.Int64,
	"uint8":   bpl.Uint8,
	"byte":    bpl.Uint8,
	"char":    bpl.Char,
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
