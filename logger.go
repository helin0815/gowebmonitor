package lsego

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/status"

	"github.com/helin0815/gowebmonitor/pkg/log"
)

var ErrorMsgExtractor = func() ErrorMessage {
	return &defaultHTTPResponseBody{}
}

func DefaultLogger(next http.Handler) http.Handler {
	logger, err := log.NewProduction("access")
	if err != nil {
		log.Fatalf("logger init failed: %s", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestShouldSkip(r) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		// TODO use buffer pool to improve performance
		bodyBuffer := bytes.NewBuffer([]byte(""))
		if r.ContentLength > 0 {
			io.Copy(bodyBuffer, r.Body)
		}
		reqBodyLength := bodyBuffer.Len()
		var reqBodyStr, respBodyStr string
		if reqBodyLength <= 256 {
			reqBodyStr = bodyBuffer.String()
		}
		r.Body = io.NopCloser(bodyBuffer)
		rw := newMetricRw(w)
		next.ServeHTTP(rw, r)
		if rw.Buffer.Len() <= 512 {
			respBodyStr = rw.Buffer.String()
		}

		printfF := logger.Debug
		if rw.Code < 400 {
			printfF = logger.Info
		} else if rw.Code >= 400 {
			printfF = logger.Warn
		} else if rw.Code >= 500 {
			printfF = logger.Error
		}

		traceId := trace.SpanContextFromContext(r.Context()).TraceID().String()
		printfF(fmt.Sprintf("[%s] %s", traceId, extractErrorMsg(r, rw)),
			zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.String("query", r.URL.RawQuery),
			zap.Int("code", rw.Code), zap.Int64("timeuse", time.Since(start).Milliseconds()),
			zap.Int("received", reqBodyLength), zap.Int("written", rw.Written),
			zap.String("reqBody", reqBodyStr), zap.String("respBody", respBodyStr),
		)
	})
}

func extractErrorMsg(r *http.Request, rw *metricRw) string {
	if rw.Code < http.StatusBadRequest || rw.Buffer.Len() == 0 {
		return "-"
	}

	if r.Header.Get("Content-Type") != "application/json" {
		return rw.Buffer.String()
	}

	var ss ErrorMessage
	if isGrpc(r) {
		ss = &status.Status{}
	} else {
		ss = ErrorMsgExtractor()
	}

	if err := json.Unmarshal(rw.Buffer.Bytes(), &ss); err != nil {
		log.Printf("Error decoding error response: %v", err)
	}

	return ss.GetMessage()
}

func isGrpc(r *http.Request) bool {
	if r.ProtoAtLeast(2, 0) && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
		return true
	}
	return false
}

type ErrorMessage interface {
	GetMessage() string
}

type defaultHTTPResponseBody struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (rb *defaultHTTPResponseBody) GetMessage() string {
	return fmt.Sprintf("%d: %s", rb.Code, rb.Msg)
}
