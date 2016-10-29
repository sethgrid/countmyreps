package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

// mwPanic wraps the outer router so all panics are caught
func mwPanic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logEvent(r, "panic", fmt.Sprintf("%v %s", rec, debug.Stack()))
			}
		}()
		h.ServeHTTP(w, r)
	})
}

// ranOnce is a hack used when initializing the logger and mwLog
var ranOnce bool

// mwLog sets up the logger and caputues logging before the request returns
func mwLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logData := logDataGet(r)
		logData["request_time"] = start.Unix()
		logData["request_id"] = fmt.Sprintf("%08x", rand.Int63n(1e9))
		logData["event"] = "request"
		logData["remote_addr"] = r.RemoteAddr
		logData["method"] = r.Method
		logData["url"] = r.URL.String()
		logData["content_length"] = r.ContentLength

		// init the logger's response writer used to caputure the status code
		// pull from a pool, set the writer, initialize / reset the response code to a sensible default, reset that this response writer has been used
		// for the logging middleware (based on noodle's logger middleware)
		// could put the ranOnce in the init, but I want to make copy-pasta easier if I use mwLog again (before turning it into a real package)
		if !ranOnce {
			ranOnce = true
			writers.New = func() interface{} {
				return &logWriter{}
			}
		}
		lw := writers.Get().(*logWriter)
		lw.ResponseWriter = w
		lw.code = http.StatusOK
		lw.headerWritten = false
		defer writers.Put(lw)

		h.ServeHTTP(lw, r)

		logData["code"] = lw.Code()
		logData["tts_ns"] = time.Since(start).Nanoseconds() / 1e6 // time to serve in nano seconds

		log.Println(logAsString(logData))
	})
}

// everything below is for the logger mw (from noodle)
// the purpose is to allow us to capture the response code that will be issued to the client

// logWriter mimics http.ResponseWriter functionality while storing
// HTTP status code for later logging
type logWriter struct {
	code          int
	headerWritten bool
	http.ResponseWriter
}

func (l *logWriter) WriteHeader(code int) {
	l.headerWritten = false
	if !l.headerWritten {
		l.ResponseWriter.WriteHeader(code)
		l.code = code
		l.headerWritten = true
	}
}

func (l *logWriter) Write(buf []byte) (int, error) {
	l.headerWritten = true
	return l.ResponseWriter.Write(buf)
}

func (l *logWriter) Code() int {
	return l.code
}

// provide other typical ResponseWriter methods
func (l *logWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return l.ResponseWriter.(http.Hijacker).Hijack()
}

func (l *logWriter) CloseNotify() <-chan bool {
	return l.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (l *logWriter) Flush() {
	l.ResponseWriter.(http.Flusher).Flush()
}

var writers sync.Pool
