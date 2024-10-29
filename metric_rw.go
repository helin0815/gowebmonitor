package lsego

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
)

// metricRw is a simple wrapper to intercept set data on a
// ResponseWriter.
type metricRw struct {
	http.ResponseWriter
	http.Hijacker

	Code    int
	Written int
	Buffer  *bytes.Buffer
}

func newMetricRw(responseWriter http.ResponseWriter) *metricRw {
	return &metricRw{ResponseWriter: responseWriter, Buffer: bytes.NewBufferString(""), Code: http.StatusOK}
}

func (w *metricRw) Write(p []byte) (int, error) {
	w.Buffer.Write(p)
	w.Written += len(p)
	return w.ResponseWriter.Write(p)
}

func (w *metricRw) WriteHeader(statusCode int) {
	w.Code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// CloseNotify implements the http.CloseNotifier interface.
func (w *metricRw) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Flush implements the http.Flusher interface.
func (w *metricRw) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *metricRw) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *metricRw) Pusher() (pusher http.Pusher) {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}
