/**
 * 日志对象，用于分类、按日期存储日志.
 */
package xlog

import (
	"os"
)

// 定义日志等级
const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// 定义输出类型
const (
	LogToStdout = iota // 输出到标准输出
	LogToStderr        // 输出到标准错误输出
	LogToFile          // 输出到文件
)

// 定义日志切割类型
const (
	RotateNone = iota
	RotateByYear
	RotateByMonth
	RotateByDate
	RotateByHour
)

// flags
const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

// 日志级别字符串
const LogLevelDebug = "debug"
const LogLevelInfo = "info"
const LogLevelWarn = "warn"
const LogLevelError = "error"
const LogLevelFatal = "fatal"

// 定义日志对象

// 定义日志存储目录，默认存储在当前目录下的logs目录
var LogDir = "./logs"

// 定义日志文件名前缀
var LogFilePrefix = "log"

// 定义日志输出类型
var logOutputType = LogToStdout

// 定义日志输出目标
var logOutput = os.Stderr

// 定义日志切割类型
var logRotateType = RotateByDate

// 设置日志记录级别
var logLevel = LevelWarn

// 日志格式标签
var logFlags = LstdFlags

// 设置记录调用栈的开关
// 默认在error以及以上的级别记录调用栈，如果需要关闭调用栈，调用DisableLogStack()方法
var logStack = true
var logStackLevel = LevelError

// 默认日志对象
var defaultLogger *StdLogger = NewStdLogger(nil)

// numLogLevel 获取日志等级的对应的数字
func numLogLevel(l string) int {
	num := LevelError
	switch l {
	case LogLevelDebug:
		num = LevelDebug
	case LogLevelInfo:
		num = LevelInfo
	case LogLevelWarn:
		num = LevelWarn
	case LogLevelError:
		num = LevelError
	case LogLevelFatal:
		num = LevelFatal
	}
	return num
}

// SetLogFilePrefix 设置日志文件前缀
func SetLogFilePrefix(prefix string) {
	LogFilePrefix = prefix
}

// SetLogDir 设置日志存储目录
func SetLogDir(path string) {
	initLogDir(path)
	LogDir = path
}

// SetLogLevel 设置日志等级
func SetLogLevel(level string) {
	numLevel := numLogLevel(level)
	if numLevel > LevelFatal {
		numLevel = LevelError
	}
	if numLevel < LevelDebug {
		numLevel = LevelDebug
	}
	logLevel = numLevel
}

// SetLogOutputType 设置日志输出类型
func SetLogOutputType(out int) {
	if out != LogToStdout && out != LogToStderr && out != LogToFile {
		logOutputType = LogToStdout
	}
	logOutputType = out
}

// SetLogFlags sets the output flags for the logger.
func SetLogFlags(flag int) {
	logFlags = flag
}

// SetLogRotateType set the way to cut log files
func SetLogRotateType(t int) {
	if t < RotateNone || t > RotateByHour {
		t = RotateByDate
	}
	logRotateType = t
}

// DisableLogStack 禁止记录调用栈信息
func DisableLogStack() {
	logStack = false
}

// EnableLogStack 开启记录调用栈信息
func EnableLogStack() {
	logStack = true
}

// Init 初始化日志设置
func Init(cfg *Config) {
	Register("default", cfg)
}
