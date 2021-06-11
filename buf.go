package xlog

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// BufLogger buffer log pieces and write all in once, make sure relevant logs show together.
// note that this will make log in a rough part of time
type BufLogger struct {
	mu     sync.Mutex
	buf    []byte
	def    *LogDefinition // 日志定义
	logger *StdLogger
}

// NewBufLogger create a buflogger of default logger with a specified cap
func NewBufLogger(size int) *BufLogger {
	return Use("default").NewBufLogger(size)
}

// DefaultBufLogger create a buflogger of default logger with a fixed cap
func DefaultBufLogger() *BufLogger {
	return Use("default").DefaultBufLogger()
}

func (l *StdLogger) NewBufLogger(size int) *BufLogger {
	if size < 0 {
		size = 2048
	}
	if size > 1024 * 100 {
		size = 1024 * 100
	}
	return &BufLogger{
		mu:     sync.Mutex{},
		buf:    make([]byte, size),
		def:    l.def,
		logger: l,
	}
}

func (l *StdLogger) DefaultBufLogger() *BufLogger {
	return &BufLogger{
		mu:     sync.Mutex{},
		buf:    make([]byte, 2048),
		def:    l.def,
		logger: l,
	}
}

// Output write log to stdout / file
func (l *BufLogger) Output(calldepth int, level, s string) error {
	now := time.Now()
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.def.Flags&(Lshortfile|Llongfile) != 0 {
		// Release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	// colorful print begin
	if l.def.ColorfulPrint && l.def.OutputType != LogToFile {
		switch level {
		case LogLevelInfo:
			l.buf = append(l.buf, "\x1b[34m"...)
		case LogLevelWarn:
			l.buf = append(l.buf, "\x1b[33m"...)
		case LogLevelError:
			l.buf = append(l.buf, "\x1b[31m"...)
		case LogLevelFatal:
			l.buf = append(l.buf, "\x1b[35m"...)
		}

	}
	// log prefix
	formatLogPrefix(&l.buf, now, level, file, line)
	// log content
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	// colorful print end
	if l.def.ColorfulPrint && l.def.OutputType != LogToFile {
		l.buf = append(l.buf, "\x1b[0m"...)
	}
	return nil
}

func (l *BufLogger) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.OutputRawBytes(l.buf)
	// 清空缓存
	l.buf = l.buf[:0]
	return nil
}

func (l *BufLogger) Close() {
	l.Flush()
	l.buf = nil
	l.logger = nil
}

func (l *BufLogger) Write(level, data string) {
	numLevel := numLogLevel(level)
	if numLevel < l.def.Level {
		return
	}
	if l.buf == nil || l.logger == nil {
		return
	}
	l.Output(3, level, data)
	if l.def.LogStack && numLevel >= l.def.LogStackLevel {
		l.Output(3, level, string(debug.Stack()))
	}
}

func (l *BufLogger) Log(level string, v ...interface{}) {
	l.Write(level, fmt.Sprint(v...))
}

func (l *BufLogger) Logf(level string, format string, v ...interface{}) {
	l.Write(level, fmt.Sprintf(format, v...))
}

func (l *BufLogger) Logln(level string, v ...interface{}) {
	l.Write(level, fmt.Sprintln(v...))
}

func (l *BufLogger) Debug(v ...interface{}) {
	l.Log(LogLevelDebug, v...)
}

func (l *BufLogger) Debugf(format string, v ...interface{}) {
	l.Logf(LogLevelDebug, format, v...)
}

func (l *BufLogger) Debugln(v ...interface{}) {
	l.Logln(LogLevelDebug, v...)
}

func (l *BufLogger) Info(v ...interface{}) {
	l.Log(LogLevelInfo, v...)
}

func (l *BufLogger) Infof(format string, v ...interface{}) {
	l.Logf(LogLevelInfo, format, v...)
}

func (l *BufLogger) Infoln(v ...interface{}) {
	l.Logln(LogLevelInfo, v...)
}

func (l *BufLogger) Warn(v ...interface{}) {
	l.Log(LogLevelWarn, v...)
}

func (l *BufLogger) Warnf(format string, v ...interface{}) {
	l.Logf(LogLevelWarn, format, v...)
}

func (l *BufLogger) Warnln(v ...interface{}) {
	l.Logln(LogLevelWarn, v...)
}

func (l *BufLogger) Error(v ...interface{}) {
	l.Log(LogLevelError, v...)
}

func (l *BufLogger) Errorf(format string, v ...interface{}) {
	l.Logf(LogLevelError, format, v...)
}

func (l *BufLogger) Errorln(v ...interface{}) {
	l.Logln(LogLevelError, v...)
}
