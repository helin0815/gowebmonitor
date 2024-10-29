package log

import (
	"net/url"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultLoggerLevel = zap.NewAtomicLevel()

func SetDefaultLoggerLevel(level zapcore.Level) {
	defaultLoggerLevel.SetLevel(level)
}

func Default(opts ...zap.Option) (*zap.Logger, error) {
	cfg := NewConfig(GetAppName())
	cfg.Level = defaultLoggerLevel
	return cfg.Build(opts...)
}

func NewProduction(name string, options ...zap.Option) (*zap.Logger, error) {
	return NewConfig(name).Build(options...)
}

func sinkRegister(url *url.URL) (zap.Sink, error) {
	if err := os.MkdirAll(filepath.Dir(url.Path), 0755); err != nil && !os.IsExist(err) {
		return nil, err
	}

	return newAsyncWriter(url.Path), nil
}
