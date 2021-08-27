package xlog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/whencome/xlog/def"
	"github.com/whencome/xlog/util"
)

// StdLogger a standard logger
type StdLogger struct {
	mu         sync.Mutex
	OutputType int
	Out        io.Writer      // 日志输出对象
	LogFile    string         // 目标日志文件
	ValidMark  string         // 设置有效标记，不匹配的时候就重新初始化
	def        *LogDefinition // 日志定义
	buf        []byte
}

// NewStdLogger create a new StdLogger, and return its address
func NewStdLogger(c *Config) *StdLogger {
	def := newLogDefinition(c)
	stdLogger := &StdLogger{
		def: def,
		mu:  sync.Mutex{},
		buf: make([]byte, 2048),
	}
	stdLogger.initOut()
	return stdLogger
}

// 更新配置
func (l *StdLogger) refresh(c *Config) {
	l.def = newLogDefinition(c)
	l.initOut()
}

// initOut 初始化输出对象
func (l *StdLogger) initOut() {
	// 关闭之前的输出对象，以支持动态重置
	if l.Out != nil {
		if l.def.OutputType != l.OutputType && l.OutputType == def.LogToFile {
			oldWriter := l.Out
			// 关闭之前的文件
			go func() {
				x, ok := oldWriter.(io.Closer)
				if ok {
					x.Close()
				}
			}()
		}
	}
	// 执行初始化
	switch l.def.OutputType {
	case def.LogToStdout:
		l.OutputType = def.LogToStdout
		l.Out = os.Stdout
	case def.LogToStderr:
		l.OutputType = def.LogToStderr
		l.Out = os.Stderr
	case def.LogToFile:
		// 设置日志文件
		l.LogFile = l.def.GetLogFilePath()
		l.ValidMark = l.calCurrentMark()
		writer, err := os.OpenFile(l.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			// 如果文件无法写入，则将日志输出到标准输出
			l.OutputType = def.LogToStdout
			l.Out = os.Stdout
		} else {
			l.OutputType = def.LogToFile
			l.Out = writer
		}
	}
}

// calCurrentMark 计算当前时间有效标记
func (l *StdLogger) calCurrentMark() string {
	if l.def.RotateType == def.RotateNone {
		return ""
	}
	return time.Now().Format(l.def.GetLogRotateTimeFmt())
}

// Output write log to stdout / file
func (l *StdLogger) Output(calldepth int, level, s string) error {
	now := time.Now()
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.def.Flags & (def.Lshortfile | def.Llongfile) != 0 {
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
	l.buf = l.buf[:0]
	// colorful print begin
	if l.def.ColorfulPrint && l.def.OutputType != def.LogToFile {
		switch level {
		case def.LogLevelInfo:
			l.buf = append(l.buf, "\x1b[34m"...)
		case def.LogLevelWarn:
			l.buf = append(l.buf, "\x1b[33m"...)
		case def.LogLevelError:
			l.buf = append(l.buf, "\x1b[31m"...)
		case def.LogLevelFatal:
			l.buf = append(l.buf, "\x1b[35m"...)
		}

	}
	// log prefix
	util.FormatLogPrefix(&l.buf, logFlags, now, level, file, line)
	// log content
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	// colorful print end
	if l.def.ColorfulPrint && l.def.OutputType != def.LogToFile {
		l.buf = append(l.buf, "\x1b[0m"...)
	}
	// 输出到文件
	return l.flush()
}

func (l *StdLogger) WriteString(s string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf = l.buf[:0]
	l.buf = append(l.buf, s...)
	return l.flush()
}

func (l *StdLogger) Write(b []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := len(b)
	l.buf = l.buf[:0]
	l.buf = append(l.buf, b...)
	return n, l.flush()
}

// Flush 用于将缓存中的日志内容吸入文件或者输出到标准输出设备
func (l *StdLogger) flush() error {
	// 计算mark，用以确认输出文件
	if l.def.OutputType == def.LogToFile && l.def.RotateType != def.RotateNone {
		curMark := l.calCurrentMark()
		if curMark != l.ValidMark {
			oldWriter := l.Out
			l.initOut()
			// 关闭之前的文件
			go func() {
				x, ok := oldWriter.(io.Closer)
				if ok {
					x.Close()
				}
			}()
		}
	}
	// 输出日志
	_, err := l.Out.Write(l.buf)
	return err
}

func (l *StdLogger) levelLog(level, data string) {
	numLevel := util.NumLogLevel(level)
	if numLevel < l.def.Level {
		return
	}
	l.Output(3, level, data)
	if l.def.LogStack && numLevel >= l.def.LogStackLevel {
		l.Output(3, level, string(debug.Stack()))
	}
}

func (l *StdLogger) Log(level string, v ...interface{}) {
	l.levelLog(level, fmt.Sprint(v...))
}

func (l *StdLogger) Logf(level string, format string, v ...interface{}) {
	l.levelLog(level, fmt.Sprintf(format, v...))
}

func (l *StdLogger) Logln(level string, v ...interface{}) {
	l.levelLog(level, fmt.Sprintln(v...))
}

func (l *StdLogger) Debug(v ...interface{}) {
	l.Log(def.LogLevelDebug, v...)
}

func (l *StdLogger) Debugf(format string, v ...interface{}) {
	l.Logf(def.LogLevelDebug, format, v...)
}

func (l *StdLogger) Debugln(v ...interface{}) {
	l.Logln(def.LogLevelDebug, v...)
}

func (l *StdLogger) Info(v ...interface{}) {
	l.Log(def.LogLevelInfo, v...)
}

func (l *StdLogger) Infof(format string, v ...interface{}) {
	l.Logf(def.LogLevelInfo, format, v...)
}

func (l *StdLogger) Infoln(v ...interface{}) {
	l.Logln(def.LogLevelInfo, v...)
}

func (l *StdLogger) Warn(v ...interface{}) {
	l.Log(def.LogLevelWarn, v...)
}

func (l *StdLogger) Warnf(format string, v ...interface{}) {
	l.Logf(def.LogLevelWarn, format, v...)
}

func (l *StdLogger) Warnln(v ...interface{}) {
	l.Logln(def.LogLevelWarn, v...)
}

func (l *StdLogger) Error(v ...interface{}) {
	l.Log(def.LogLevelError, v...)
}

func (l *StdLogger) Errorf(format string, v ...interface{}) {
	l.Logf(def.LogLevelError, format, v...)
}

func (l *StdLogger) Errorln(v ...interface{}) {
	l.Logln(def.LogLevelError, v...)
}

func (l *StdLogger) Fatal(v ...interface{}) {
	l.Log(def.LogLevelFatal, v...)
	os.Exit(1)
}

func (l *StdLogger) Fatalf(format string, v ...interface{}) {
	l.Logf(def.LogLevelFatal, format, v...)
	os.Exit(1)
}

func (l *StdLogger) Fatalln(v ...interface{}) {
	l.Logln(def.LogLevelFatal, v...)
	os.Exit(1)
}

func (l *StdLogger) Panic(v ...interface{}) {
	l.Log(def.LogLevelFatal, v...)
	panic(fmt.Sprint(v...))
}

func (l *StdLogger) Panicf(format string, v ...interface{}) {
	l.Logf(def.LogLevelFatal, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (l *StdLogger) Panicln(v ...interface{}) {
	l.Logln(def.LogLevelFatal, v...)
	panic(fmt.Sprintln(v...))
}

// Raw record origin raw log
func (l *StdLogger) Raw(v ...interface{}) {
	_ = l.WriteString(fmt.Sprint(v...))
}

func (l *StdLogger) Rawf(format string, v ...interface{}) {
	_ = l.WriteString(fmt.Sprintf(format, v...))
}

func (l *StdLogger) Rawln(v ...interface{}) {
	_ = l.WriteString(fmt.Sprintln(v...))
}

func (l *StdLogger) Json(v interface{}) {
	d, e := json.Marshal(v)
	if e != nil {
		return
	}
	d = append(d, '\n')
	_, _ = l.Write(d)
}
