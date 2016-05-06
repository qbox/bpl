package bufio

import (
	"testing"
)

// ---------------------------------------------------

func TestReader(t *testing.T) {

	text := []byte("Hello, bpl")
	br := NewReaderBuffer(text)
	if br.Buffered() != len(text) {
		t.Fatal("br.Buffered() != len(text)")
	}
	if c, err := br.ReadByte(); err != nil || c != 'H' {
		t.Fatal("ReadByte:", c, err)
	}
	if b, err := br.Peek(2); err != nil || b[0] != 'e' || b[1] != 'l' {
		t.Fatal("Peek:", b, err)
	}
	if s, err := br.ReadString(','); err != nil || s != "ello," {
		t.Fatal("ReadString:", s, err)
	}
	if n, err := br.Discard(4); err != nil || n != 4 {
		t.Fatal("Discard:", n, err)
	}
	if br.Buffered() != 0 {
		t.Fatal("br.Buffered() != 0")
	}
}

// ---------------------------------------------------
