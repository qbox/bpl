package hex

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"strings"
)

func undump1(w *bytes.Buffer, text string) { // bd c2 c1 24 93 55 2a 4d

	b, err := hex.DecodeString(strings.Replace(text, " ", "", -1))
	if err != nil {
		panic(err)
	}
	w.Write(b)
}

// Undump reverts `hexdump -C binary` result back to a binary data.
//
func Undump(w *bytes.Buffer, text string) {

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "  ", 4)
		max := 3
		if len(parts) < max {
			max = len(parts)
		}
		addr := parts[0]
		if len(addr) < 8 {
			continue
		}
		_, err := strconv.ParseInt(addr, 16, 64)
		if err != nil {
			continue
		}
		for i := 1; i < max; i++ {
			undump1(w, parts[i])
		}
	}
}

// Reader returns a reader that reverts `hexdump -C binary` result back to a binary data.
//
func Reader(text string) *bytes.Reader {

	var w bytes.Buffer
	Undump(&w, text)
	return bytes.NewReader(w.Bytes())
}
