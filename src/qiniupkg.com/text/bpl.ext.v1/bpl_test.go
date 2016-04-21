package bpl

import (
	"encoding/json"
	"testing"

	"qiniupkg.com/text/bpl.v1/binary"
	"qiniupkg.com/text/bpl.v1/bufio"
)

const codeBasic = `

sub1 = int8 uint16

subType = cstring

doc = [sub1 uint32 float32 cstring subType float64]
`

type subType struct {
	Foo string
}

type fooType struct {
	A int8
	B uint16
	C uint32
	D float32
	E string
	F subType
	G float64
}

func TestBasic(t *testing.T) {

	foo := &fooType{
		A: 1, B: 2, C: 3, D: 3.14, E: "Hello", F: subType{Foo: "foo"}, G: 7.52,
		// 1 + 2 + 4 + 4 + 6 + 4 + 8 = 29
	}
	b, err := binary.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	if len(b) != 29 {
		t.Fatal("len(b) != 29, len:", len(b), "data:", string(b))
	}

	r, err := NewFromString(codeBasic, "")
	if err != nil {
		t.Fatal("New failed:", err)
	}
	in := bufio.NewReaderBuffer(b)
	v, err := r.Match(in, nil)
	if err != nil {
		t.Fatal("Match failed:", err, "len:", len(b))
	}
	ret, err := json.Marshal(v)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(ret) != `[null,3,3.14,"Hello","foo",7.52]` {
		t.Fatal("ret:", string(ret))
	}
}
