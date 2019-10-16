package xlog

import (
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

// StdLogger a standard logger
type StdLogger struct {
	mu         sync.Mutex
	OutputType int
	Out        io.Writer // 日志输出对象
	LogFile    string    // 目标日志文件
	ValidMark  string    // 设置有效标记，不匹配的时候就重新初始化
	buf        []byte
}

// NewStdLogger create a new StdLogger, and return its address
func NewStdLogger() *StdLogger {
	stdLogger := &StdLogger{
		mu:  sync.Mutex{},
		buf: make([]byte, 2048),
	}
	stdLogger.initOut()
	return stdLogger
}

// initOut 初始化输出对象
func (sl *StdLogger) initOut() {
	// 关闭之前的输出对象，以支持动态重置
	if sl.Out != nil {
		if logOutputType != sl.OutputType && sl.OutputType == LogToFile {
			oldWriter := sl.Out
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
	switch logOutputType {
	case LogToStdout:
		sl.OutputType = LogToStdout
		sl.Out = os.Stdout
	case LogToStderr:
		sl.OutputType = LogToStderr
		sl.Out = os.Stderr
	case LogToFile:
		// 设置日志文件
		sl.LogFile = getLogFilePath()
		sl.ValidMark = sl.calCurrentMark()
		writer, err := os.OpenFile(sl.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("open log file [%s] failed : %s", sl.LogFile, err)
			// 如果文件无法写入，则将日志输出到标准输出
			sl.OutputType = LogToStdout
			sl.Out = os.Stdout
		} else {
			sl.OutputType = LogToFile
			sl.Out = writer
		}
	}
}

// calCurrentMark 计算当前时间有效标记
func (sl *StdLogger) calCurrentMark() string {
	if logRotateType == RotateNone {
		return ""
	}
	return time.Now().Format(getLogRotateTimeFmt())
}

// Output write log to stdout / file
func (sl *StdLogger) Output(calldepth int, level, s string) error {
	now := time.Now() // get this early.
	var file string
	var line int
	sl.mu.Lock()
	defer sl.mu.Unlock()
	if logFlags&(Lshortfile|Llongfile) != 0 {
		// Release lock while getting caller info - it's expensive.
		sl.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		sl.mu.Lock()
	}
	sl.buf = sl.buf[:0]
	formatLogPrefix(&sl.buf, now, level, file, line)
	sl.buf = append(sl.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		sl.buf = append(sl.buf, '\n')
	}
	// 输出到文件
	return sl.flush()
}

// Flush 用于将缓存中的日志内容吸入文件或者输出到标准输出设备
func (sl *StdLogger) flush() error {
	// 计算mark，用以确认输出文件
	if logOutputType == LogToFile && logRotateType != RotateNone {
		curMark := sl.calCurrentMark()
		if curMark != sl.ValidMark {
			oldWriter := sl.Out
			sl.initOut()
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
	_, err := sl.Out.Write(sl.buf)
	return err
}
