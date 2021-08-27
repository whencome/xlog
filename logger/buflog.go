package logger

import (
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/whencome/xlog/def"
	"github.com/whencome/xlog/util"
)

type BufLogger struct {
	writer     io.Writer
	mu         sync.Mutex
	buf        []byte
	bufSize    int // 缓存大小
	logStack   bool
	stackLevel int
}

func NewBufLogger(w io.Writer) *BufLogger {
	return &BufLogger{
		writer:     w,
		mu:         sync.Mutex{},
		buf:        make([]byte, 0),
		bufSize:    10240, // 10k
		logStack:   false,
		stackLevel: def.LevelError,
	}
}

func NewStackBufLogger(w io.Writer) *BufLogger {
	return &BufLogger{
		writer:     w,
		mu:         sync.Mutex{},
		buf:        make([]byte, 0),
		bufSize:    10240, // 10k
		logStack:   true,
		stackLevel: def.LevelError,
	}
}

// 设置buffer缓冲区大小
func (l *BufLogger) SetBufferSize(n int) {
	l.bufSize = n
}

// Output write log to stdout / file
func (l *BufLogger) Output(calldepth int, level, s string) error {
	now := time.Now()
	var ok bool
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	// log prefix
	util.FormatLogPrefix(&l.buf, def.Ldate|def.Ltime|def.Lshortfile, now, level, file, line)
	// log content
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	return nil
}

func (l *BufLogger) levelLog(level, data string) {
	numLevel := util.NumLogLevel(level)
	_ = l.Output(3, level, data)
	if l.logStack && numLevel >= l.stackLevel {
		_ = l.Output(3, level, string(debug.Stack()))
	}
	// 如果超过缓冲区大小，则强制写入
	if len(l.buf) >= l.bufSize {
		l.Write()
	}
}

func (l *BufLogger) Log(level string, v ...interface{}) {
	l.levelLog(level, fmt.Sprint(v...))
}

func (l *BufLogger) Logf(level string, format string, v ...interface{}) {
	l.levelLog(level, fmt.Sprintf(format, v...))
}

func (l *BufLogger) Logln(level string, v ...interface{}) {
	l.levelLog(level, fmt.Sprintln(v...))
}

func (l *BufLogger) Debug(v ...interface{}) {
	l.Log(def.LogLevelDebug, v...)
}

func (l *BufLogger) Debugf(format string, v ...interface{}) {
	l.Logf(def.LogLevelDebug, format, v...)
}

func (l *BufLogger) Debugln(v ...interface{}) {
	l.Logln(def.LogLevelDebug, v...)
}

func (l *BufLogger) Info(v ...interface{}) {
	l.Log(def.LogLevelInfo, v...)
}

func (l *BufLogger) Infof(format string, v ...interface{}) {
	l.Logf(def.LogLevelInfo, format, v...)
}

func (l *BufLogger) Infoln(v ...interface{}) {
	l.Logln(def.LogLevelInfo, v...)
}

func (l *BufLogger) Warn(v ...interface{}) {
	l.Log(def.LogLevelWarn, v...)
}

func (l *BufLogger) Warnf(format string, v ...interface{}) {
	l.Logf(def.LogLevelWarn, format, v...)
}

func (l *BufLogger) Warnln(v ...interface{}) {
	l.Logln(def.LogLevelWarn, v...)
}

func (l *BufLogger) Error(v ...interface{}) {
	l.Log(def.LogLevelError, v...)
}

func (l *BufLogger) Errorf(format string, v ...interface{}) {
	l.Logf(def.LogLevelError, format, v...)
}

func (l *BufLogger) Errorln(v ...interface{}) {
	l.Logln(def.LogLevelError, v...)
}

func (l *BufLogger) Write() (int, error) {
	if len(l.buf) == 0 || l.writer == nil {
		fmt.Printf("buffer or writer empty, buf size = %d\n", len(l.buf))
		return 0, nil
	}
	if len(l.buf) == 0 || l.buf[len(l.buf)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	n, err := l.writer.Write(l.buf)
	l.Reset()
	return n, err
}

func (l *BufLogger) Reset() {
	l.buf = make([]byte, 0)
	return
}

func (l *BufLogger) Close() {
	_, _ = l.Write()
	l.buf = nil
}
