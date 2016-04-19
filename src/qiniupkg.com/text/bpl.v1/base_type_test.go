package bpl_test

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"testing"

	"qiniupkg.com/text/bpl.v1"
)

func TestBaseType(t *testing.T) {

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, 123)
	in := bufio.NewReader(bytes.NewReader(b))

	ctx := bpl.NewContext()
	v, err := bpl.NamedBaseType("foo", bpl.Int64).Match(in, ctx)
	if err != nil {
		t.Fatal("NamedBaseType.Match failed:", err)
	}
	if v != int64(123) {
		t.Fatal("v != 123")
	}
	if v, ok := ctx.Var("foo"); !ok || v != int64(123) {
		t.Fatal("v != 123")
	}
}
