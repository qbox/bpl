package main

import (
	"flag"
	"fmt"
	"os"

	"qiniupkg.com/text/mockd.v1"
)

var (
	host = flag.String("h", "", "bind address.")
)

// qmockd -h <host> <mock.log>
//
func main() {

	flag.Parse()
	args := flag.Args()

	if *host == "" || len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qmockd -h <host> <mock.log>")
		flag.PrintDefaults()
		return
	}

	mockd.ListenAndServe(*host, args[0])
}
