package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultScheme = "llog"

	AppName        = "TEST_APP_NAME"
	LogBase        = "LOG_BASE"
	LogBaseDefault = "/data/log/"
)

func init() {
	if err := zap.RegisterSink(defaultScheme, sinkRegister); err != nil {
		log.Fatalln(err)
	}
}

type Config struct {
	zap.Config

	loggerName string
}

func NewConfig(name string) *Config {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		"stdout",
		fmt.Sprintf("%s://%s/%s.log", defaultScheme, GetLogOutputPath(), name),
	}
	cfg.ErrorOutputPaths = []string{
		"stderr",
		fmt.Sprintf("%s://%s/%s_error.log", defaultScheme, GetLogOutputPath(), name),
	}
	cfg.Encoding = "console"
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return &Config{
		Config: cfg,

		loggerName: name,
	}
}

func (c *Config) Build(opts ...zap.Option) (*zap.Logger, error) {
	opts = append(opts, zap.Fields(zap.String("logger", c.loggerName)))
	c.buildCustomEncoder()
	return c.Config.Build(opts...)
}

func (c *Config) buildCustomEncoder() {
	c.EncoderConfig.ConsoleSeparator = " " // 日志规范要求一个空格
	c.EncoderConfig.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + l.CapitalString() + "]")
	}
	c.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[v1] [" + t.Format("2006-01-02 15:04:05.000") + "]")
	}
	c.EncoderConfig.EncodeCaller = func(call zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		trPath := call.TrimmedPath()
		enc.AppendString("[main] " + trPath)
	}
}

func GetLogOutputPath() string {
	logBase := os.Getenv(LogBase)
	if logBase == "" {
		logBase = LogBaseDefault
	}

	if _, err := os.Stat(logBase); os.IsNotExist(err) {
		pwd, _ := os.Getwd()
		logBase = filepath.Join(pwd, "logs")
	}

	return filepath.Join(logBase, GetAppName())
}

func GetAppName() string {
	appName := os.Getenv(AppName)
	if appName == "" {
		appName = "default"
	}
	return appName
}
