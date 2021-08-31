package gomodel

import (
	"github.com/whencome/xlog"
	"github.com/whencome/xlog/logger"
	"io"
	"regexp"
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

func (l *Logger) getSQLCommand(q string) string {
	q = strings.TrimSpace(q)
	p, err := regexp.Compile(`\s`)
	if err != nil {
		return strings.ToUpper(q[:strings.Index(q, " ")])
	}
	s := p.Split(q, -1)
	if len(s) > 0 {
		return strings.ToUpper(s[0])
	}
	return strings.ToUpper(q[:strings.Index(q, " ")])
}

func (l *Logger) SetCommand(q string) {
	q = strings.TrimSpace(q)
	l.l.Put("command", l.getSQLCommand(q))
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
