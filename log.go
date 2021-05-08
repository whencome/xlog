package xlog

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
)

func Write(level, data string) {
	l := Use("default")
	numLevel := numLogLevel(level)
	if numLevel < l.def.Level {
		return
	}
	l.Output(3, level, data)
	if l.def.LogStack && numLevel >= l.def.LogStackLevel {
		l.Output(3, level, string(debug.Stack()))
	}
}

// Log record a specified level's log
func Log(level string, v ...interface{}) {
	Write(level, fmt.Sprint(v...))
}

// Logf record a specified level's formatted log
func Logf(level string, format string, v ...interface{}) {
	Write(level, fmt.Sprintf(format, v...))
}

// Logf record a specified level's log with a new line
func Logln(level string, v ...interface{}) {
	Write(level, fmt.Sprintln(v...))
}

func Debug(v ...interface{}) {
	Log(LogLevelDebug, v...)
}

func Debugf(format string, v ...interface{}) {
	Logf(LogLevelDebug, format, v...)
}

func Debugln(v ...interface{}) {
	Logln(LogLevelDebug, v...)
}

func Info(v ...interface{}) {
	Log(LogLevelInfo, v...)
}

func Infof(format string, v ...interface{}) {
	Logf(LogLevelInfo, format, v...)
}

func Infoln(v ...interface{}) {
	Logln(LogLevelInfo, v...)
}

func Warn(v ...interface{}) {
	Log(LogLevelWarn, v...)
}

func Warnf(format string, v ...interface{}) {
	Logf(LogLevelWarn, format, v...)
}

func Warnln(v ...interface{}) {
	Logln(LogLevelWarn, v...)
}

func Error(v ...interface{}) {
	Log(LogLevelError, v...)
}

func Errorf(format string, v ...interface{}) {
	Logf(LogLevelError, format, v...)
}

func Errorln(v ...interface{}) {
	Logln(LogLevelError, v...)
}

func Fatal(v ...interface{}) {
	Log(LogLevelFatal, v...)
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	Logf(LogLevelFatal, format, v...)
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	Logln(LogLevelFatal, v...)
	os.Exit(1)
}

func Panic(v ...interface{}) {
	Log(LogLevelFatal, v...)
	panic(fmt.Sprint(v...))
}

func Panicf(format string, v ...interface{}) {
	Logf(LogLevelFatal, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func Panicln(v ...interface{}) {
	Logln(LogLevelFatal, v...)
	panic(fmt.Sprintln(v...))
}

// Raw record origin raw log
func Raw(v ...interface{}) {
	Use("default").OutputRaw(fmt.Sprint(v...))
}

func Rawf(format string, v ...interface{}) {
	Use("default").OutputRaw(fmt.Sprintf(format, v...))
}

func Rawln(v ...interface{}) {
	Use("default").OutputRaw(fmt.Sprintln(v...))
}

func Json(v interface{}) {
	d, e := json.Marshal(v)
	if e != nil {
		return
	}
	d = append(d, '\n')
	Use("default").OutputRawBytes(d)
}
