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

term2 = term1 *('+' term1/add | '-' term1/sub | '|' term1/bitor | '^' term1/xor)

term3 = term2 *('<' term2/lt | '>' term2/gt | "==" term2/eq | "<=" term2/le | ">=" term2/ge | "!=" term2/ne)

term4 = term3 *("&&" term3/iand)

iexpr = term4 *("||" term4/ior)

index = '['/istart iexpr ']'/iend

casebody = (INT/casei ':' expr/source) %= ';'/ARITY ?(';' "default" ':' expr)/ARITY

caseexpr = "case"/istart! iexpr/source '{'/iend casebody ?';' '}' /case

exprblock = true/istart! iexpr (@'{' | "do")/iend expr

ifexpr = "if" exprblock *("elif" exprblock)/ARITY ?("else"! expr)/ARITY /if

readexpr = "read" exprblock /read

evalexpr = "eval" exprblock /eval

doexpr = "do"/istart! iexpr /iend /do

letexpr = "let"! IDENT/var '='/istart! iexpr /iend /let

assertexpr = ("assert"/istart! iexpr /iend) /assert

gblexpr = "global"! IDENT/var '='/istart! iexpr /iend /global

lzwexpr = "lzw"/istart! iexpr /iend ',' /istart! iexpr /iend ',' /istart! iexpr /iend exprblock /lzw

dynexpr = caseexpr | readexpr | evalexpr | assertexpr | ifexpr | letexpr | doexpr | gblexpr | lzwexpr

type =
	IDENT/ident |
	(index IDENT/ident)/array |
	('*'! IDENT/ident)/array0 |
	('?'! IDENT/ident)/array01 |
	('+'! IDENT/ident)/array1

member = ((IDENT type)/member | dynexpr)/xline

cmember = (IDENT/ident ?(index/array | '*'/array0 | '?'/array01 | '+'/array1) IDENT/member | dynexpr)/xline

retexpr = "return"/istart! iexpr /iend

cstruct = cmember %= ';'/ARITY ?';' ?retexpr/ARITY /struct

struct = member %= ';'/ARITY ?';' ?retexpr/ARITY /struct

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
	'(' iexpr %= ','/ARITY ?"..."/ARITY ?',' ')'/call |
	'.' IDENT/mref |
	'[' ?iexpr/ARITY ?':'/ARITY ?iexpr/ARITY ']'/index

ifactor =
	INT/pushi |
	STRING/pushs |
	(IDENT/ref | '('! iexpr ')' | '[' iexpr %= ','/ARITY ?',' ']'/slice) *atom |
	"sizeof"! '(' IDENT/sizeof ')' |
	'{'! (iexpr ':' iexpr) %= ','/ARITY ?',' '}'/map |
	'^' ifactor/bitnot |
	'-' ifactor/neg |
	'+' ifactor

cexpr = INT/cpushi

const = (IDENT '=' cexpr ';')/const

doc = +(
	(IDENT '=' expr/xline ';')/assign |
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

// Stack returns nil (no stack). It is required by tpl.Interpreter engine.
//
func (p *Compiler) Stack() interpreter.Stack {

	return nil
}

// -----------------------------------------------------------------------------

var exports = map[string]interface{}{
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

	"$iand": and,
	"$ior":  or,

	"$sizeof": (*Compiler).sizeof,
	"$map":    (*Compiler).fnMap,
	"$slice":  (*Compiler).fnSlice,
	"$index":  (*Compiler).index,
	"$ARITY":  (*Compiler).arity,
	"$call":   (*Compiler).call,
	"$ref":    (*Compiler).ref,
	"$mref":   (*Compiler).mref,
	"$pushi":  (*Compiler).pushi,
	"$pushs":  (*Compiler).pushs,
	"$cpushi": (*Compiler).cpushi,
	"$let":    (*Compiler).fnLet,
	"$global": (*Compiler).fnGlobal,
	"$eval":   (*Compiler).fnEval,
	"$do":     (*Compiler).fnDo,
	"$if":     (*Compiler).fnIf,
	"$read":   (*Compiler).fnRead,
	"$lzw":    (*Compiler).fnLzw,
	"$case":   (*Compiler).fnCase,
	"$assert": (*Compiler).fnAssert,
	"$const":  (*Compiler).fnConst,
	"$casei":  (*Compiler).casei,
	"$source": (*Compiler).source,
	"$member": (*Compiler).member,
	"$struct": (*Compiler).gostruct,
	"$xline":  (*Compiler).xline,
}

var builtins = map[string]bpl.Ruler{
	"int8":     bpl.Int8,
	"int16":    bpl.Int16,
	"int32":    bpl.Int32,
	"int64":    bpl.Int64,
	"uint8":    bpl.Uint8,
	"byte":     bpl.Uint8,
	"char":     bpl.Char,
	"uint16":   bpl.Uint16,
	"uint24":   bpl.Uint24,
	"uint32":   bpl.Uint32,
	"uint64":   bpl.Uint64,
	"uint16be": bpl.Uintbe(2),
	"uint24be": bpl.Uintbe(3),
	"uint32be": bpl.Uintbe(4),
	"uint64be": bpl.Uintbe(8),
	"uint16le": bpl.Uint16,
	"uint24le": bpl.Uint24,
	"uint32le": bpl.Uint32,
	"uint64le": bpl.Uint64,
	"float32":  bpl.Float32,
	"float64":  bpl.Float64,
	"cstring":  bpl.CString,
	"nil":      bpl.Nil,
	"eof":      bpl.EOF,
	"done":     bpl.Done,
	"bson":     bson.Type,
	"dump":     dump(0),
}

// -----------------------------------------------------------------------------
