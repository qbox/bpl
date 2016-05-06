package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"

	"qiniupkg.com/text/bpl.ext.v1"
	"qiniupkg.com/text/bpl.v1/bufio"
	"qiniupkg.com/x/log.v7"
)

// -----------------------------------------------------------------------------

// A Env is the environment of a callback.
//
type Env struct {
	Src       *net.TCPConn
	Dest      *net.TCPConn
	Direction string
}

// A ReverseProxier is a reverse proxier server.
//
type ReverseProxier struct {
	Addr       string
	Backend    string
	OnResponse func(io.Reader, *Env) (err error)
	OnRequest  func(io.Reader, *Env) (err error)
	Listened   chan bool
}

// ListenAndServe listens on `Addr` and serves to proxy requests to `Backend`.
//
func (p *ReverseProxier) ListenAndServe() (err error) {

	addr := p.Addr
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("ListenAndServe(tcprproxyd) %s failed: %v\n", addr, err)
		return
	}
	if p.Listened != nil {
		p.Listened <- true
	}
	err = p.Serve(l)
	if err != nil {
		log.Fatalf("ListenAndServe(tcprproxyd) %s failed: %v\n", addr, err)
	}
	return
}

func onNil(r io.Reader, env *Env) (err error) {

	_, err = io.Copy(ioutil.Discard, r)
	return
}

// Serve serves to proxy requests to `Backend`.
//
func (p *ReverseProxier) Serve(l net.Listener) (err error) {

	defer l.Close()

	backend, err := net.ResolveTCPAddr("tcp", p.Backend)
	if err != nil {
		return
	}

	onResponse := p.OnResponse
	if onResponse == nil {
		onResponse = onNil
	}

	onRequest := p.OnRequest
	if onRequest == nil {
		onRequest = onNil
	}

	for {
		c1, err1 := l.Accept()
		if err1 != nil {
			return err1
		}
		c := c1.(*net.TCPConn)
		go func() {
			c2, err2 := net.DialTCP("tcp", nil, backend)
			if err2 != nil {
				log.Error("tcprproxy: dial backend failed -", p.Backend, "error:", err2)
				c.Close()
				return
			}

			go func() {
				r2 := io.TeeReader(c2, c)
				onResponse(r2, &Env{Src: c2, Dest: c, Direction: "RESP"})
				c.CloseWrite()
				c2.CloseRead()
			}()

			r := io.TeeReader(c, c2)
			err2 = onRequest(r, &Env{Src: c, Dest: c2, Direction: "REQ"})
			if err2 != nil {
				log.Info("tcprproxy (request):", err2)
			}
			c.CloseRead()
			c2.CloseWrite()
		}()
	}
}

// -----------------------------------------------------------------------------

var (
	host     = flag.String("h", "", "listen host (listenIp:port).")
	backend  = flag.String("b", "", "backend host (backendIp:port).")
	protocol = flag.String("p", "", "protocol file in BPL syntax, default is guessed by <port>.")
	output   = flag.String("o", "", "output log file, default is stderr.")
)

var (
	baseDir string // $HOME/.qbpl/formats/
)

func fileExists(file string) bool {

	_, err := os.Stat(file)
	return err == nil
}

func guessProtocol(host string) string {

	index := strings.Index(host, ":")
	if index >= 0 {
		proto := baseDir + host[index+1:] + ".bpl"
		if fileExists(proto) {
			return proto
		}
	}
	return ""
}

// qbplproxy -h <listenIp:port> -b <backendIp:port> [-p <protocol>.bpl -o <output>.log]
//
func main() {

	flag.Parse()
	if *host == "" || *backend == "" {
		fmt.Fprintln(os.Stderr, "Usage: qbplproxy -h <listenIp:port> -b <backendIp:port> [-p <protocol>.bpl -o <output>.log]")
		flag.PrintDefaults()
		return
	}

	baseDir = os.Getenv("HOME") + "/.qbpl/formats/"
	if *protocol == "" {
		*protocol = guessProtocol(*host)
		if *protocol == "" {
			*protocol = guessProtocol(*backend)
			if *protocol == "" {
				log.Fatalln("I can't know protocol by listening port, please use -p <protocol>.")
			}
		}
	} else {
		if path.Ext(*protocol) == "" {
			*protocol = baseDir + *protocol + ".bpl"
		}
	}

	onBpl := onNil
	if *protocol != "nil" {
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
		onBpl = func(r io.Reader, env *Env) (err error) {
			in := bufio.NewReader(r)
			ctx := bpl.NewContext()
			ctx.Globals["DUMP_PREFIX"] = "[" + env.Direction + "]"
			_, err = ruler.SafeMatch(in, ctx)
			if err != nil {
				log.Error("Match failed:", err)
			}
			in.WriteTo(ioutil.Discard)
			return
		}
	}

	rp := &ReverseProxier{
		Addr:       *host,
		Backend:    *backend,
		OnRequest:  onBpl,
		OnResponse: onBpl,
	}
	rp.ListenAndServe()
}

// -----------------------------------------------------------------------------
