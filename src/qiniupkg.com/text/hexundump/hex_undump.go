package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func hexUndump1(w *bytes.Buffer, text string) { // bd c2 c1 24 93 55 2a 4d

	b, err := hex.DecodeString(strings.Replace(text, " ", "", -1))
	if err != nil {
		panic(err)
	}
	w.Write(b)
}

func hexUndump(w *bytes.Buffer, text string) {

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "  ", 4)
		max := 3
		if len(parts) < max {
			max = len(parts)
		}
		for i := 1; i < max; i++ {
			hexUndump1(w, parts[i])
		}
	}
}

// Usage: hexundump <hexdump-file> <binary-file>
//
func main() {

	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: hexundump <hexdump-file> <binary-file>\n\n")
		return
	}
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	w := bytes.NewBuffer(nil)
	hexUndump(w, string(b))
	err = ioutil.WriteFile(os.Args[2], w.Bytes(), 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
