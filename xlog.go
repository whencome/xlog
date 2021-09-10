/**
 * 日志对象，用于分类、按日期存储日志.
 */
package xlog

import (
	"io"
	"os"
	"sync"

	"github.com/whencome/xlog/def"
	"github.com/whencome/xlog/logger"
	"github.com/whencome/xlog/util"
)

// 定义日志对象

// 定义日志存储目录，默认存储在当前目录下的logs目录
var LogDir = "./logs"

// 定义日志文件名前缀
var LogFilePrefix = "log"

// 定义日志输出类型
var logOutputType = def.LogToStdout

// 定义日志输出目标
var logOutput = os.Stderr

// 定义日志切割类型
var logRotateType = def.RotateByDate

// 设置日志记录级别
var logLevel = def.LevelWarn

// 日志格式标签
var logFlags = def.LstdFlags

// 是否开启彩色打印
var colorfulPrint = true

// 设置记录调用栈的开关
// 默认在error以及以上的级别记录调用栈，如果需要关闭调用栈，调用DisableLogStack()方法
var logStack = true
var logStackLevel = def.LevelError

// 默认日志对象
var defaultLogger *StdLogger = NewStdLogger(nil)

// 定义日志映射列表
var loggerMaps sync.Map

// SetLogFilePrefix 设置日志文件前缀
func SetLogFilePrefix(prefix string) {
	LogFilePrefix = prefix
}

// SetLogDir 设置日志存储目录
func SetLogDir(path string) {
	_, _ = util.InitLogDir(path)
	LogDir = path
}

// SetLogLevel 设置日志等级
func SetLogLevel(level string) {
	numLevel := util.NumLogLevel(level)
	if numLevel > def.LevelFatal {
		numLevel = def.LevelError
	}
	if numLevel < def.LevelDebug {
		numLevel = def.LevelDebug
	}
	logLevel = numLevel
}

// SetLogOutputType 设置日志输出类型
func SetLogOutputType(out int) {
	if out != def.LogToStdout && out != def.LogToStderr && out != def.LogToFile {
		logOutputType = def.LogToStdout
	}
	logOutputType = out
}

// SetLogFlags sets the output flags for the logger.
func SetLogFlags(flag int) {
	logFlags = flag
}

// SetLogRotateType set the way to cut log files
func SetLogRotateType(t int) {
	if t < def.RotateNone || t > def.RotateByHour {
		t = def.RotateByDate
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

// DisableColorfulPrint 禁止彩色日志打印
func DisableColorfulPrint() {
	colorfulPrint = false
}

// EnableColorfulPrint 开启彩色日志打印
func EnableColorfulPrint() {
	colorfulPrint = true
}

// Init 初始化日志设置
func Init(cfg *Config) {
	Register("default", cfg)
}

// 注册一个日志对象
func Register(k string, cfg *Config) {
	var stdLogger *StdLogger
	// 检查logger是否已经存在
	l, ok := loggerMaps.Load(k)
	if ok && l != nil {
		stdLogger = l.(*StdLogger)
		stdLogger.refresh(cfg)
		return
	}
	// 创建一个新的logger
	stdLogger = NewStdLogger(cfg)
	loggerMaps.Store(k, stdLogger)
}

// 清除全部日志设置
func Clear() {
	loggerMaps.Range(func(key, value interface{}) bool {
		k, ok := key.(string)
		if ok {
			l := MustUse(k)
			if l != nil {
				_ = l.Close()
			}
		}
		loggerMaps.Delete(key)
		return true
	})
}

// 选择需要使用的日志对象
func Use(k string) *StdLogger {
	l, ok := loggerMaps.Load(k)
	if !ok {
		return defaultLogger
	}
	sl, ok := l.(*StdLogger)
	if !ok {
		return defaultLogger
	}
	return sl
}

// 强制使用指定的日志对象
func MustUse(k string) *StdLogger {
	l, ok := loggerMaps.Load(k)
	if !ok {
		return nil
	}
	sl, ok := l.(*StdLogger)
	if !ok {
		return nil
	}
	return sl
}

// NewKVLogger 创建一个普通的KVLogger
func NewKVLogger(w io.Writer) *logger.KVLogger {
	return logger.NewKVLogger(w)
}

// NewTimerKVLogger 创建一个自带记时间的KVLogger
func NewTimerKVLogger(w io.Writer) *logger.KVLogger {
	return logger.NewTimerKVLogger(w)
}

// NewKVLogger 创建一个普通的KVLogger
func NewBufLogger(w io.Writer) *logger.BufLogger {
	return logger.NewBufLogger(w)
}

// NewTimerKVLogger 创建一个自带记时间的KVLogger
func NewStackBufLogger(w io.Writer) *logger.BufLogger {
	return logger.NewStackBufLogger(w)
}
