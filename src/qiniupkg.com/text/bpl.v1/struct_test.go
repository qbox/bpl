package bpl_test

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"

	"qiniupkg.com/text/bpl.v1"
	"qiniupkg.com/text/bpl.v1/bufio"
)

type fixedType struct {
	A int8
	B uint16
	C uint32
	D float32
}

func TestFixedStruct(t *testing.T) {
	v := fixedType{
		A: 1,
		B: 2,
		C: 3,
		D: 3.14,
	}

	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, &v)
	if err != nil {
		t.Fatal("binary.Write failed:", err)
	}
	if b.Len() != 11 {
		t.Fatal("len != 11")
	}

	members := []bpl.NamedType{
		{Name: "a", Type: bpl.Int8},
		{Name: "b", Type: bpl.Uint16},
		{Name: "c", Type: bpl.Uint32},
		{Name: "d", Type: bpl.Float32},
	}
	struc := bpl.Struct(members)
	if struc.SizeOf() != 11 {
		t.Fatal("struct.size != 11")
	}

	in := bufio.NewReaderBuffer(b.Bytes())
	if in.Buffered() != 11 {
		t.Fatal("len != 11")
	}

	ctx := bpl.NewContext()
	ret, err := struc.Match(in, ctx)
	if err != nil {
		t.Fatal("struc.Match failed:", err)
	}
	text, err := json.Marshal(ret)
	if err != nil {
		t.Fatal("json.Marshal failed:", err)
	}
	if string(text) != `{"a":1,"b":2,"c":3,"d":3.14}` {
		t.Fatal("json.Marshal result:", string(text))
	}
}
