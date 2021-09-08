package xlog

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/whencome/xlog/def"
	"github.com/whencome/xlog/util"
)

func levelLog(level, data string) {
	l := Use("default")
	if l.def.Disabled {
		return
	}
	numLevel := util.NumLogLevel(level)
	if numLevel < l.def.Level {
		return
	}
	_ = l.Output(3, level, data)
	if l.def.LogStack && numLevel >= l.def.LogStackLevel {
		_ = l.Output(3, level, string(debug.Stack()))
	}
}

// Log record a specified level's log
func Log(level string, v ...interface{}) {
	levelLog(level, fmt.Sprint(v...))
}

// Logf record a specified level's formatted log
func Logf(level string, format string, v ...interface{}) {
	levelLog(level, fmt.Sprintf(format, v...))
}

// Logf record a specified level's log with a new line
func Logln(level string, v ...interface{}) {
	levelLog(level, fmt.Sprintln(v...))
}

func Debug(v ...interface{}) {
	Log(def.LogLevelDebug, v...)
}

func Debugf(format string, v ...interface{}) {
	Logf(def.LogLevelDebug, format, v...)
}

func Debugln(v ...interface{}) {
	Logln(def.LogLevelDebug, v...)
}

func Info(v ...interface{}) {
	Log(def.LogLevelInfo, v...)
}

func Infof(format string, v ...interface{}) {
	Logf(def.LogLevelInfo, format, v...)
}

func Infoln(v ...interface{}) {
	Logln(def.LogLevelInfo, v...)
}

func Warn(v ...interface{}) {
	Log(def.LogLevelWarn, v...)
}

func Warnf(format string, v ...interface{}) {
	Logf(def.LogLevelWarn, format, v...)
}

func Warnln(v ...interface{}) {
	Logln(def.LogLevelWarn, v...)
}

func Error(v ...interface{}) {
	Log(def.LogLevelError, v...)
}

func Errorf(format string, v ...interface{}) {
	Logf(def.LogLevelError, format, v...)
}

func Errorln(v ...interface{}) {
	Logln(def.LogLevelError, v...)
}

func Fatal(v ...interface{}) {
	Log(def.LogLevelFatal, v...)
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	Logf(def.LogLevelFatal, format, v...)
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	Logln(def.LogLevelFatal, v...)
	os.Exit(1)
}

func Panic(v ...interface{}) {
	Log(def.LogLevelFatal, v...)
	panic(fmt.Sprint(v...))
}

func Panicf(format string, v ...interface{}) {
	Logf(def.LogLevelFatal, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func Panicln(v ...interface{}) {
	Logln(def.LogLevelFatal, v...)
	panic(fmt.Sprintln(v...))
}

// Raw record origin raw log
func Raw(v ...interface{}) {
	_ = Use("default").WriteString(fmt.Sprint(v...))
}

func Rawf(format string, v ...interface{}) {
	_ = Use("default").WriteString(fmt.Sprintf(format, v...))
}

func Rawln(v ...interface{}) {
	_ = Use("default").WriteString(fmt.Sprintln(v...))
}

func Json(v interface{}) {
	d, e := json.Marshal(v)
	if e != nil {
		return
	}
	d = append(d, '\n')
	_, _ = Use("default").Write(d)
}
