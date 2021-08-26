package xlog

import (
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type BufLogger struct {
	mu      sync.Mutex
	buf     []byte
	maxSize int
	def     *LogDefinition // 日志定义
	logger  *StdLogger
}

func NewBufLogger(size int) *BufLogger {
	return Use("default").NewBufLogger(size)
}

func DefaultBufLogger() *BufLogger {
	return Use("default").DefaultBufLogger()
}

func (l *StdLogger) NewBufLogger(size int) *BufLogger {
	return &BufLogger{
		mu:      sync.Mutex{},
		buf:     make([]byte, 0),
		maxSize: size,
		def:     l.def,
		logger:  l,
	}
}

func (l *StdLogger) DefaultBufLogger() *BufLogger {
	return &BufLogger{
		mu:      sync.Mutex{},
		buf:     make([]byte, 0),
		maxSize: 2048,
		def:     l.def,
		logger:  l,
	}
}

// Output write log to stdout / file
func (l *BufLogger) Output(calldepth int, level, s string) error {
	now := time.Now()
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	// l.buf = l.buf[:0]
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
	if len(l.buf) >= l.maxSize {
		_ = l.Flush()
	}
	return nil
}

// OutputRaw write raw log to stdout / file
func (l *BufLogger) OutputRaw(s string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf = append(l.buf, s...)
	if len(l.buf) >= l.maxSize {
		_ = l.Flush()
	}
	return nil
}

func (l *BufLogger) OutputRawBytes(b []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf = append(l.buf, b...)
	if len(l.buf) >= l.maxSize {
		_ = l.Flush()
	}
	return nil
}

func (l *BufLogger) Flush() error {
	if len(l.buf) > 0 {
		_ = l.logger.OutputRawBytes(l.buf)
		// 清空缓存
		l.buf = l.buf[:0]
	}
	return nil
}

func (l *BufLogger) Close() {
	_ = l.Flush()
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
	_ = l.Output(3, level, data)
	if l.def.LogStack && numLevel >= l.def.LogStackLevel {
		_ = l.Output(3, level, string(debug.Stack()))
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

// Raw record origin raw log
func (l *BufLogger) Raw(v ...interface{}) {
	_ = l.OutputRaw(fmt.Sprint(v...))
}

func (l *BufLogger) Rawf(format string, v ...interface{}) {
	_ = l.OutputRaw(fmt.Sprintf(format, v...))
}

func (l *BufLogger) Rawln(v ...interface{}) {
	_ = l.OutputRaw(fmt.Sprintln(v...))
}

func (l *BufLogger) Json(v interface{}) {
	d, e := json.Marshal(v)
	if e != nil {
		return
	}
	d = append(d, '\n')
	_ = l.OutputRawBytes(d)
}
