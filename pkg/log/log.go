package log

import (
	"context"
	"log"
	"sync"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	once sync.Once

	sugaredLogger *zap.SugaredLogger
)

func SugaredLogger() *zap.SugaredLogger {
	once.Do(func() {
		l, err := Default(zap.AddCallerSkip(1))
		if err != nil {
			log.Panicln(err)
		}

		sugaredLogger = l.Sugar()
	})

	return sugaredLogger
}

func Debugf(format string, args ...interface{}) {
	SugaredLogger().Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	SugaredLogger().Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	SugaredLogger().Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	SugaredLogger().Errorf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	SugaredLogger().Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	SugaredLogger().Fatalf(format, args...)
}

func Printf(format string, v ...interface{}) {
	SugaredLogger().Infof(format, v...)
}

type Logger interface {
	Debugf(format string, args ...interface{})

	Infof(format string, args ...interface{})

	Warnf(format string, args ...interface{})

	Errorf(format string, args ...interface{})

	Panicf(format string, args ...interface{})

	Fatalf(format string, args ...interface{})
}

type wrapLogger struct {
	*zap.SugaredLogger

	traceId string
}

// WithContext 使用本函数返回的 Logger 会自动带上 traceId
func WithContext(ctx context.Context) Logger {
	traceId := trace.SpanContextFromContext(ctx).TraceID().String()
	if traceId == "" {
		traceId = "-"
	}
	return &wrapLogger{traceId: traceId, SugaredLogger: SugaredLogger()}
}

func (wl *wrapLogger) Debugf(format string, args ...interface{}) {
	wl.SugaredLogger.Debugf("[%s] "+format, append([]any{wl.traceId}, args...)...)
}

func (wl *wrapLogger) Infof(format string, args ...interface{}) {
	wl.SugaredLogger.Infof("[%s] "+format, append([]any{wl.traceId}, args...)...)
}

func (wl *wrapLogger) Warnf(format string, args ...interface{}) {
	wl.SugaredLogger.Warnf("[%s] "+format, append([]any{wl.traceId}, args...)...)
}

func (wl *wrapLogger) Errorf(format string, args ...interface{}) {
	wl.SugaredLogger.Errorf("[%s] "+format, append([]any{wl.traceId}, args...)...)
}

func (wl *wrapLogger) Panicf(format string, args ...interface{}) {
	wl.SugaredLogger.Panicf("[%s] "+format, append([]any{wl.traceId}, args...)...)
}

func (wl *wrapLogger) Fatalf(format string, args ...interface{}) {
	wl.SugaredLogger.Fatalf("[%s] "+format, append([]any{wl.traceId}, args...)...)
}
