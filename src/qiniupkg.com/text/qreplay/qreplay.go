package main

import (
	"flag"
	"fmt"
	"os"

	"qiniupkg.com/text/replay.v1"
)

var (
	filter = flag.String("f", "", "filter condition. eg. -f [REQ]")
	host   = flag.String("s", "", "remote address to dial.")
)

// qreplay -s <host:port> -f <filter> <replay.log>
//
func main() {

	flag.Parse()

	if *host == "" {
		fmt.Fprintln(os.Stderr, "Usage: qreplay -s <host:port> -f <filter> <replay.log>")
		flag.PrintDefaults()
		return
	}

	var in *os.File
	args := flag.Args()
	if len(args) > 0 {
		file := args[0]
		f, err := os.Open(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Open failed:", file)
		}
		defer f.Close()
		in = f
	} else {
		in = os.Stdin
	}

	err := replay.HexRequest(*host, in, *filter)
	if err != nil {
		fmt.Fprintln(os.Stderr, "replay.HexRequest:", err)
		os.Exit(1)
	}
}
