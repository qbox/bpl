package bpl

import (
	"encoding/json"
	"testing"

	"qiniupkg.com/text/bpl.v1/binary"
	"qiniupkg.com/text/bpl.v1/bufio"
)

// -----------------------------------------------------------------------------

const codeArray = `

sub1 = int8 uint16

subType = {
    array [2]cstring
}

doc = [sub1 uint32 float32 cstring subType float64]
`

type subType2 struct {
	Foo string
	Bar string
}

type fooType2 struct {
	A int8
	B uint16
	C uint32
	D float32
	E string
	F subType2
	G float64
}

func TestArray(t *testing.T) {

	foo := &fooType2{
		A: 1, B: 2, C: 3, D: 3.14, E: "Hello", F: subType2{Foo: "foo", Bar: "bar"}, G: 7.52,
		// 1 + 2 + 4 + 4 + 6 + 4 + 8 = 29
	}
	b, err := binary.Marshal(&foo)
	if err != nil {
		t.Fatal("binary.Marshal failed:", err)
	}
	if len(b) != 33 {
		t.Fatal("len(b) != 29, len:", len(b), "data:", string(b))
	}

	r, err := NewFromString(codeArray, "")
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
	if string(ret) != `[null,3,3.14,"Hello",{"array":["foo","bar"]},7.52]` {
		t.Fatal("ret:", string(ret))
	}
}
