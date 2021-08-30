package gomodel

import (
	"github.com/whencome/xlog"
	"github.com/whencome/xlog/logger"
	"io"
	"strings"
)

type Logger struct {
	l *logger.KVLogger
}

func NewLogger() *Logger {
	return &Logger{
		l:xlog.NewTimerKVLogger(xlog.MustUse("db")),
	}
}

func CustomLogger(w io.Writer) *Logger {
	return &Logger{
		l:xlog.NewTimerKVLogger(w),
	}
}

func (l *Logger) SetCommand(q string) {
	q = strings.TrimSpace(q)
	command := q[:strings.Index(q, " ")]
	l.l.Put("command", command)
	l.l.Put("sql", q)
}

func (l *Logger) Fail(msg interface{}) {
	l.l.Put("result", "failed")
	l.l.Put("message", msg)
	_, _ = l.l.Write()
}

func (l *Logger) Success()  {
	l.l.Put("result", "success")
	l.l.Put("message", "ok")
	_, _ = l.l.Write()
}

func (l *Logger) Close() {
	l.l.Close()
	l.l = nil
}
