package bpl

import (
	"io/ioutil"

	"qiniupkg.com/text/bpl.v1"
	"qiniupkg.com/text/tpl.v1/interpreter"
)

// -----------------------------------------------------------------------------

// New compiles bpl source code and returns the corresponding matching unit.
//
func New(code []byte, fname string) (r bpl.Ruler, err error) {

	p := NewCompiler()
	engine, err := interpreter.New(p, interpreter.InsertSemis)
	if err != nil {
		panic(err)
	}

	err = engine.MatchExactly(code, fname)
	if err != nil {
		panic(err)
	}

	return p.Ret()
}

// NewFromFile compiles bpl source file and returns the corresponding matching unit.
//
func NewFromFile(fname string) (r bpl.Ruler, err error) {

	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return
	}
	return New(b, fname)
}

// -----------------------------------------------------------------------------
