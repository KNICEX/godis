package logx

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
)

type logger struct {
	logFile   *os.File
	l         *log.Logger
	CallDepth int
}

func NewLogger() Logger {
	return logger{
		l:         log.New(os.Stdout, "", log.LstdFlags),
		CallDepth: 2,
	}
}

func (l logger) output(level LogLevel, callDepth int, msg string) {
	_, file, line, ok := runtime.Caller(callDepth)
	content := ""
	if !ok {
		content = fmt.Sprintf("[%s] %s", levelFlags[level], msg)
	} else {
		content = fmt.Sprintf("[%s] [%s:%d] %s", levelFlags[level], file, line, msg)
	}
	_ = l.l.Output(0, content)
}

func (l logger) WithCtx(ctx context.Context) Logger {
	//TODO implement me
	panic("implement me")
}

func (l logger) Debug(v ...any) {
	l.output(Debug, l.CallDepth, fmt.Sprintln(v...))
}

func (l logger) Debugf(format string, v ...any) {
	l.output(Debug, l.CallDepth, fmt.Sprintf(format, v...))
}

func (l logger) Info(v ...any) {
	l.output(Info, l.CallDepth, fmt.Sprintln(v...))
}

func (l logger) Infof(format string, v ...any) {
	l.output(Info, l.CallDepth, fmt.Sprintf(format, v...))
}

func (l logger) Warn(v ...any) {
	l.output(Warn, l.CallDepth, fmt.Sprintln(v...))
}

func (l logger) Warnf(format string, v ...any) {
	l.output(Warn, l.CallDepth, fmt.Sprintf(format, v...))
}

func (l logger) Error(v ...any) {
	l.output(Error, l.CallDepth, fmt.Sprintln(v...))
}

func (l logger) Errorf(format string, v ...any) {
	l.output(Error, l.CallDepth, fmt.Sprintf(format, v...))
}

func (l logger) Fatal(v ...any) {
	l.output(Fatal, l.CallDepth, fmt.Sprintln(v...))
	os.Exit(1)
}

func (l logger) Fatalf(format string, v ...any) {
	l.output(Fatal, l.CallDepth, fmt.Sprintf(format, v...))
	os.Exit(1)
}
