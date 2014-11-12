package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "expvar"
	_ "net/http/pprof"

	"flag"

	"github.com/bakins/net-http-recover"
	"github.com/bakins/zfsd"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/justinas/alice"
)

func main() {
	var address string
	var perms uint
	flag.StringVar(&address, "address", "/tmp/zfsd.sock", "TCP Address or unix socket to listen")
	flag.UintVar(&perms, "perms", 0700, "permissions for unix socket")
	flag.Parse()

	var l net.Listener

	if strings.ContainsAny(address, "/") {
		addr, err := net.ResolveUnixAddr("unix", address)
		if err != nil {
			log.Fatal(err)
		}
		u, err := net.ListenUnix("unix", addr)
		if err != nil {
			log.Fatal(err)
		}
		defer u.Close()

		err = os.Chmod(address, os.FileMode(perms))
		if err != nil {
			log.Fatal(err)
		}
		//http://stackoverflow.com/questions/16681944/how-to-reliably-unlink-a-unix-domain-socket-in-go-programming-language
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
		go func(c chan os.Signal) {
			// Wait for a SIGINT or SIGKILL:
			sig := <-c
			log.Printf("Caught signal %s: shutting down.", sig)
			// Stop listening (and unlink the socket if unix type):
			u.Close()
			// And we're done:
			os.Exit(0)
		}(sigc)
		l = u
	} else {
		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			log.Fatal(err)
		}
		t, err := net.ListenTCP("tcp", addr)
		if err != nil {
			log.Fatal(err)
		}
		l = t
	}

	r := mux.NewRouter()
	r.StrictSlash(true)

	// default mux will have the profiler handlers
	r.PathPrefix("/debug/").Handler(http.DefaultServeMux)

	chain := alice.New(
		handlers.CompressHandler,
		func(h http.Handler) http.Handler {
			return NewLogger(os.Stdout, h)
		},
		func(h http.Handler) http.Handler {
			return recovery.Handler(os.Stderr, h, true)
		},
	)

	rpcs := rpc.NewServer()
	rpcs.RegisterCodec(NewCodec(), "application/json")
	rpcs.RegisterService(&zfsd.ZFS{}, "")

	// TODO: unix socket
	s := &http.Server{
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	r.Handle("/_zfs_", chain.Then(rpcs))
	log.Fatal(s.Serve(l))

}
