package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"qiniupkg.com/text/bpl.ext.v1"
	"qiniupkg.com/text/bpl.v1/bufio"
)

var (
	protocol = flag.String("p", "", "protocol file in BPL syntax.")
	output   = flag.String("o", "", "output log file, default is stderr.")
)

// qbpl -p <protocol>.bpl [-o <output>.log] <file>
//
func main() {

	flag.Parse()
	if *protocol == "" {
		fmt.Fprintln(os.Stderr, "Usage: qbpl -p <protocol>.bpl [-o <output>.log] <file>")
		flag.PrintDefaults()
		return
	}

	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			log.Fatalln("Create log file failed:", err)
		}
		defer f.Close()
		bpl.SetDumper(f)
	}

	ruler, err := bpl.NewFromFile(*protocol)
	if err != nil {
		log.Fatalln("bpl.NewFromFile failed:", err)
	}

	var in *bufio.Reader
	args := flag.Args()
	if len(args) > 0 {
		file := args[0]
		f, err := os.Open(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Open failed:", file)
		}
		defer f.Close()
		in = bufio.NewReader(f)
	} else {
		in = bufio.NewReader(os.Stdin)
	}

	ctx := bpl.NewContext()
	_, err = ruler.SafeMatch(in, ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Match failed:", err)
		return
	}
}
