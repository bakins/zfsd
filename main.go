package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bakins/net-http-recover"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/justinas/alice"

	_ "expvar"
	_ "net/http/pprof"
)

type (
	ZFS struct {
	}
)

func main() {
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
	rpcs.RegisterService(&ZFS{}, "")

	// TODO: unix socket
	s := &http.Server{
		Addr:           ":9373",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	r.Handle("/_rpc_", chain.Then(rpcs))
	log.Fatal(s.ListenAndServe())

}
