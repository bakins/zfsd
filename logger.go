package main

// From https://github.com/mistifyio/mistify-agent/blob/88133d9c14c32185a84dc64f144820b3e02decc5/rpc/logger.go

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/rpc"
	rpc_json "github.com/gorilla/rpc/json"
)

type (
	logger struct {
		writer  io.Writer
		handler http.Handler
	}

	logHandler struct {
		http.ResponseWriter
		status int
		size   int
	}

	logEntry struct {
		Method    string    `json:"method"`
		Duration  float64   `json:"duration"`
		URL       string    `json:"url"`
		Time      time.Time `json:"time"`
		Status    int       `json:"status"`
		Size      int       `json:"size"`
		UserAgent string    `json:"user_agent"`
		Client    string    `json:"client"`
	}

	Codec struct {
		*rpc_json.Codec
	}
)

func NewCodec() *Codec {
	c := rpc_json.NewCodec()
	return &Codec{c}
}

func (c *Codec) NewRequest(r *http.Request) rpc.CodecRequest {
	cr := c.Codec.NewRequest(r)

	if m, err := cr.Method(); err == nil {
		context.Set(r, "_rpc_method_", m)
	}

	return cr
}

func NewLogger(w io.Writer, h http.Handler) *logger {
	return &logger{
		writer:  w,
		handler: h,
	}
}

func (l *logHandler) Write(b []byte) (int, error) {
	if l.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		l.status = http.StatusOK
	}
	size, err := l.ResponseWriter.Write(b)
	l.size += size
	return size, err
}

func (l *logHandler) WriteHeader(s int) {
	l.ResponseWriter.WriteHeader(s)
	l.status = s
}

func (l *logHandler) Header() http.Header {
	return l.ResponseWriter.Header()
}

func (l *logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := logHandler{
		ResponseWriter: w,
	}
	now := time.Now()
	l.handler.ServeHTTP(&h, r)

	method := ""
	m := context.Get(r, "_rpc_method_")
	if m != nil {
		method = m.(string)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		host = r.RemoteAddr
	}

	entry := logEntry{
		Method:    method,
		Duration:  time.Since(now).Seconds(),
		URL:       r.URL.RequestURI(),
		Time:      now,
		Status:    h.status,
		Size:      h.size,
		UserAgent: r.UserAgent(),
		Client:    host,
	}
	b := &bytes.Buffer{}
	e := json.NewEncoder(b)
	err = e.Encode(&entry)
	if err != nil {
		return
	}
	b.WriteTo(l.writer)
}
