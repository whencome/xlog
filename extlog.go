package xlog

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
)

// 定义日志映射列表
var loggerMaps sync.Map

// 注册一个日志对象
func Register(k string, cfg *Config) {
	var stdLogger *StdLogger
	// 检查logger是否已经存在
	logger, ok := loggerMaps.Load(k)
	if ok && logger != nil {
		stdLogger = logger.(*StdLogger)
		stdLogger.refresh(cfg)
		return
	}
	// 创建一个新的logger
	stdLogger = NewStdLogger(cfg)
	loggerMaps.Store(k, stdLogger)
}

// 选择需要使用的日志对象
func Use(k string) *StdLogger {
	logger, ok := loggerMaps.Load(k)
	if !ok {
		return stdLogger
	}
	l, ok := logger.(*StdLogger)
	if !ok {
		return stdLogger
	}
	return l
}

func (l *StdLogger) Write(level, data string) {
	numLevel := numLogLevel(level)
	if numLevel < logLevel {
		return
	}
	l.Output(3, level, data)
	if l.def.LogStack && numLevel >= l.def.LogStackLevel {
		l.Output(3, level, string(debug.Stack()))
	}
}

func (l *StdLogger) Log(level string, v ...interface{}) {
	l.Write(level, fmt.Sprint(v...))
}

func (l *StdLogger) Logf(level string, format string, v ...interface{}) {
	l.Write(level, fmt.Sprintf(format, v...))
}

func (l *StdLogger) Logln(level string, v ...interface{}) {
	l.Write(level, fmt.Sprintln(v...))
}

func (l *StdLogger) Debug(v ...interface{}) {
	l.Log(LogLevelDebug, v...)
}

func (l *StdLogger) Debugf(format string, v ...interface{}) {
	l.Logf(LogLevelDebug, format, v...)
}

func (l *StdLogger) Debugln(v ...interface{}) {
	l.Logln(LogLevelDebug, v...)
}

func (l *StdLogger) Info(v ...interface{}) {
	l.Log(LogLevelInfo, v...)
}

func (l *StdLogger) Infof(format string, v ...interface{}) {
	l.Logf(LogLevelInfo, format, v...)
}

func (l *StdLogger) Infoln(v ...interface{}) {
	l.Logln(LogLevelInfo, v...)
}

func (l *StdLogger) Warn(v ...interface{}) {
	l.Log(LogLevelWarn, v...)
}

func (l *StdLogger) Warnf(format string, v ...interface{}) {
	l.Logf(LogLevelWarn, format, v...)
}

func (l *StdLogger) Warnln(v ...interface{}) {
	l.Logln(LogLevelWarn, v...)
}

func (l *StdLogger) Error(v ...interface{}) {
	l.Log(LogLevelError, v...)
}

func (l *StdLogger) Errorf(format string, v ...interface{}) {
	l.Logf(LogLevelError, format, v...)
}

func (l *StdLogger) Errorln(v ...interface{}) {
	l.Logln(LogLevelError, v...)
}

func (l *StdLogger) Fatal(v ...interface{}) {
	l.Log(LogLevelFatal, v...)
	os.Exit(1)
}

func (l *StdLogger) Fatalf(format string, v ...interface{}) {
	l.Logf(LogLevelFatal, format, v...)
	os.Exit(1)
}

func (l *StdLogger) Fatalln(v ...interface{}) {
	l.Logln(LogLevelFatal, v...)
	os.Exit(1)
}

func (l *StdLogger) Panic(v ...interface{}) {
	l.Log(LogLevelFatal, v...)
	panic(fmt.Sprint(v...))
}

func (l *StdLogger) Panicf(format string, v ...interface{}) {
	l.Logf(LogLevelFatal, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (l *StdLogger) Panicln(v ...interface{}) {
	l.Logln(LogLevelFatal, v...)
	panic(fmt.Sprintln(v...))
}

// Raw record origin raw log
func (l *StdLogger) Raw(v ...interface{}) {
	l.OutputRaw(fmt.Sprint(v...))
}

func (l *StdLogger) Rawf(format string, v ...interface{}) {
	l.OutputRaw(fmt.Sprintf(format, v...))
}

func (l *StdLogger) Rawln(v ...interface{}) {
	l.OutputRaw(fmt.Sprintln(v...))
}

func (l *StdLogger) Json(v interface{}) {
	d, e := json.Marshal(v)
	if e != nil {
		return
	}
	d = append(d, '\n')
	l.OutputRawBytes(d)
}

