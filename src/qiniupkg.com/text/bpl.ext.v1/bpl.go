package bpl

import (
	"io"
	"io/ioutil"

	"qiniupkg.com/text/bpl.v1"
	"qiniupkg.com/text/bpl.v1/bufio"
	"qiniupkg.com/text/tpl.v1/interpreter"
)

// -----------------------------------------------------------------------------

// A Ruler is a matching unit.
//
type Ruler struct {
	R bpl.Ruler
}

// MatchStream matches input stream `r`, and returns matching result.
//
func (p Ruler) MatchStream(r io.Reader) (v interface{}, err error) {

	in := bufio.NewReader(r)
	ctx := bpl.NewContext()
	return p.R.Match(in, ctx)
}

// MatchBuffer matches input buffer `b`, and returns matching result.
//
func (p Ruler) MatchBuffer(b []byte) (v interface{}, err error) {

	in := bufio.NewReaderBuffer(b)
	ctx := bpl.NewContext()
	return p.R.Match(in, ctx)
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
