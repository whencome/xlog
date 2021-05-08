package xlog

import (
	"io"
	"os"
	"runtime"
	"sync"
	"time"
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
		if l.def.OutputType != l.OutputType && l.OutputType == LogToFile {
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
	case LogToStdout:
		l.OutputType = LogToStdout
		l.Out = os.Stdout
	case LogToStderr:
		l.OutputType = LogToStderr
		l.Out = os.Stderr
	case LogToFile:
		// 设置日志文件
		l.LogFile = l.def.GetLogFilePath()
		l.ValidMark = l.calCurrentMark()
		writer, err := os.OpenFile(l.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			// 如果文件无法写入，则将日志输出到标准输出
			l.OutputType = LogToStdout
			l.Out = os.Stdout
		} else {
			l.OutputType = LogToFile
			l.Out = writer
		}
	}
}

// calCurrentMark 计算当前时间有效标记
func (l *StdLogger) calCurrentMark() string {
	if l.def.RotateType == RotateNone {
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
	if l.def.Flags & (Lshortfile|Llongfile) != 0 {
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
	// 输出到文件
	return l.flush()
}

// OutputRaw write raw log to stdout / file
func (l *StdLogger) OutputRaw(s string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf = l.buf[:0]
	l.buf = append(l.buf, s...)
	return l.flush()
}

func (l *StdLogger) OutputRawBytes(b []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf = l.buf[:0]
	l.buf = append(l.buf, b...)
	return l.flush()
}

// Flush 用于将缓存中的日志内容吸入文件或者输出到标准输出设备
func (l *StdLogger) flush() error {
	// 计算mark，用以确认输出文件
	if l.def.OutputType == LogToFile && l.def.RotateType != RotateNone {
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
