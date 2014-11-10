package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	runtime_pprof "runtime/pprof"
	"time"

	"github.com/bakins/net-http-recover"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/justinas/alice"
)

type (
	ZFS struct {
	}
)

func attachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	for _, profile := range runtime_pprof.Profiles() {
		router.Handle(fmt.Sprintf("/debug/pprof/%s", profile.Name()), pprof.Handler(profile.Name()))
	}
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
}

func main() {
	r := mux.NewRouter()
	r.StrictSlash(true)

	attachProfiler(r)

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
