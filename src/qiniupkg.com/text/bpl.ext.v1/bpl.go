package bpl

import (
	"errors"
	"io"
	"io/ioutil"
	"os"

	"qiniupkg.com/text/bpl.v1"
	"qiniupkg.com/text/bpl.v1/bufio"
	"qiniupkg.com/text/tpl.v1/interpreter"
	"qiniupkg.com/x/log.v7"
)

// -----------------------------------------------------------------------------

var (
	// Dumper is used for dumping log informations.
	Dumper = log.New(os.Stderr, "", log.Ldefault)
)

// SetDumper sets the dumper instance for dumping log informations.
//
func SetDumper(w io.Writer, flags ...int) {

	flag := log.Ldefault
	if len(flags) > 0 {
		flag = flags[0]
	}
	Dumper = log.New(w, "", flag)
}

// -----------------------------------------------------------------------------

type dump int

func (p dump) Match(in *bufio.Reader, ctx *bpl.Context) (v interface{}, err error) {

	dom := ctx.Dom()
	if dom != nil {
		Dumper.Info(dom)
	}
	return
}

func (p dump) SizeOf() int {

	return 0
}

// -----------------------------------------------------------------------------

// A Ruler is a matching unit.
//
type Ruler struct {
	bpl.Ruler
}

// SafeMatch matches input stream `in`, and returns matching result.
//
func (p Ruler) SafeMatch(in *bufio.Reader, ctx *bpl.Context) (v interface{}, err error) {

	defer func() {
		if e := recover(); e != nil {
			switch v := e.(type) {
			case string:
				err = errors.New(v)
			case error:
				err = v
			default:
				panic(e)
			}
		}
	}()

	return p.Ruler.Match(in, ctx)
}

// MatchStream matches input stream `r`, and returns matching result.
//
func (p Ruler) MatchStream(r io.Reader) (v interface{}, err error) {

	in := bufio.NewReader(r)
	ctx := bpl.NewContext()
	return p.SafeMatch(in, ctx)
}

// MatchBuffer matches input buffer `b`, and returns matching result.
//
func (p Ruler) MatchBuffer(b []byte) (v interface{}, err error) {

	in := bufio.NewReaderBuffer(b)
	ctx := bpl.NewContext()
	return p.SafeMatch(in, ctx)
}

// -----------------------------------------------------------------------------

// New compiles bpl source code and returns the corresponding matching unit.
//
func New(code []byte, fname string) (r Ruler, err error) {

	p := NewCompiler()
	engine, err := interpreter.New(p, interpreter.InsertSemis)
	if err != nil {
		panic(err)
	}

	err = engine.MatchExactly(code, fname)
	if err != nil {
		return
	}

	return p.Ret()
}

// NewFromString compiles bpl source code and returns the corresponding matching unit.
//
func NewFromString(code string, fname string) (r Ruler, err error) {

	return New([]byte(code), fname)
}

// NewFromFile compiles bpl source file and returns the corresponding matching unit.
//
func NewFromFile(fname string) (r Ruler, err error) {

	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return
	}
	return New(b, fname)
}

// -----------------------------------------------------------------------------
