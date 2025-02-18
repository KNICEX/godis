package logx

import "context"

type LogLevel int

const (
	Debug = iota
	Info
	Warn
	Error
	Fatal
)

var levelFlags = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

type Logger interface {
	WithCtx(ctx context.Context) Logger
	Debug(v ...any)
	Debugf(format string, v ...any)
	Info(v ...any)
	Infof(format string, v ...any)
	Warn(v ...any)
	Warnf(format string, v ...any)
	Error(v ...any)
	Errorf(format string, v ...any)
	Fatal(v ...any)
	Fatalf(format string, v ...any)
}

var defaultLogger Logger = NewLogger()

func L() Logger {
	return defaultLogger
}
