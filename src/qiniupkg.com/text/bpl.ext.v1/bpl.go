package bpl

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"qiniupkg.com/text/bpl.v1"
	"qiniupkg.com/text/bpl.v1/bufio"
	"qiniupkg.com/text/tpl.v1/interpreter"
	"qiniupkg.com/x/log.v7"
)

// -----------------------------------------------------------------------------

var (
	// Ldefault is default flag for `Dumper`.
	Ldefault = log.Llevel | log.LstdFlags

	// Dumper is used for dumping log informations.
	Dumper = log.New(os.Stdout, "", Ldefault)

	// SetCaseType controls to set `_type` into matching result or not.
	SetCaseType = true
)

// SetDumper sets the dumper instance for dumping log informations.
//
func SetDumper(w io.Writer, flags ...int) {

	flag := Ldefault
	if len(flags) > 0 {
		flag = flags[0]
	}
	Dumper = log.New(w, "", flag)
}

func writePrefix(b *bytes.Buffer, lvl int) {

	for i := 0; i < lvl; i++ {
		b.WriteString("  ")
	}
}

// DumpDom dumps a dom tree.
//
func DumpDom(b *bytes.Buffer, dom interface{}, lvl int) {

	if dom == nil {
		b.WriteString("<nil>")
		return
	}
	switch v := dom.(type) {
	case []interface{}:
		b.WriteByte('[')
		for _, item := range v {
			b.WriteByte('\n')
			writePrefix(b, lvl+1)
			DumpDom(b, item, lvl+1)
			b.WriteByte(',')
		}
		b.WriteByte('\n')
		writePrefix(b, lvl)
		b.WriteByte(']')
	case map[string]interface{}:
		b.WriteByte('{')
		keys := make([]string, 0, len(v))
		for key := range v {
			if strings.HasPrefix(key, "_") {
				continue
			}
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			item := v[key]
			b.WriteByte('\n')
			writePrefix(b, lvl+1)
			b.WriteString(key)
			b.WriteString(": ")
			DumpDom(b, item, lvl+1)
		}
		b.WriteByte('\n')
		writePrefix(b, lvl)
		b.WriteByte('}')
	case []byte:
		b.WriteByte('\n')
		d := hex.Dumper(b)
		d.Write(v)
		d.Close()
	default:
		ret, _ := json.Marshal(dom)
		b.Write(ret)
	}
}

// -----------------------------------------------------------------------------

type dump int

func (p dump) Match(in *bufio.Reader, ctx *bpl.Context) (v interface{}, err error) {

	var b bytes.Buffer
	if prefix, ok := ctx.Globals["DUMP_PREFIX"]; ok {
		b.WriteString(prefix.(string))
	}
	b.WriteByte('\n')
	DumpDom(&b, ctx.Dom(), 0)
	Dumper.Info(b.String())
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

	p := newCompiler()
	engine, err := interpreter.New(p, interpreter.InsertSemis)
	if err != nil {
		return
	}

	p.ipt = engine
	err = engine.MatchExactly(code, fname)
	if err != nil {
		return
	}

	if DumpCode != 0 {
		p.code.Dump()
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

// NewContext returns a new matching Context.
//
func NewContext() *bpl.Context {

	return bpl.NewContext()
}

// -----------------------------------------------------------------------------

// SetDumpCode sets dump code mode:
//	"1" - dump code with rem instruction.
//	"2" - dump code without rem instruction.
//  else - don't dump code.
//
func SetDumpCode(dumpCode string) {

	switch dumpCode {
	case "true", "1":
		DumpCode = 1
	case "2":
		DumpCode = 2
	default:
		DumpCode = 0
	}
}

// -----------------------------------------------------------------------------
