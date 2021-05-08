package xlog

import (
	"fmt"
	"os"
	"time"
)

// Config 定义日志配置
type Config struct {
	LogPath       string `json:"log_path"`        // 定义日志根路径
	LogPrefix     string `json:"log_prefix"`      // 日志文件前缀
	Output        string `json:"output"`          // 日志输出类型,file,stdout,stderr
	LogLevel      string `json:"log_level"`       // 日志等级，可取值:debug,info,warn,error,fatal
	Rotate        string `json:"rotate"`          // 日志切割类型,可取值：none,year,month,date,hour
	LogStackLevel string `json:"log_stack_level"` // 记录调用栈信息的日志等级
	ColorfulPrint bool   `json:"colorful_print"`  // 是否开启彩色打印，仅适用于标准输出，不适用于文件输出
}

// LogDefinition 日志定义，由Config转换后得到
type LogDefinition struct {
	Dir           string   // 定义日志存储目录，默认存储在当前目录下的logs目录
	FilePrefix    string   // 定义日志文件名前缀
	OutputType    int      // 定义日志输出类型
	Output        *os.File // 定义日志输出目标
	RotateType    int      // 定义日志切割类型
	Level         int      // 设置日志记录级别
	Flags         int      // 日志格式标签
	LogStack      bool     // 是否记录日志调用栈信息
	LogStackLevel int      // 记录调用栈的日志等级
	ColorfulPrint bool     // 是否开启彩色打印，仅适用于标准输出，不适用于文件输出
}

// 返回一个默认的日志定义
func defaultLogDefinition() *LogDefinition {
	def := &LogDefinition{}
	def.Dir = LogDir
	def.Level = logLevel
	def.FilePrefix = LogFilePrefix
	def.Flags = logFlags
	def.Output = logOutput
	def.OutputType = logOutputType
	def.RotateType = logRotateType
	def.LogStack = logStack
	def.LogStackLevel = logStackLevel
	def.ColorfulPrint = colorfulPrint
	return def
}

// 根据配置返回一个日志定义
func newLogDefinition(cfg *Config) *LogDefinition {
	// 没有配置则忽略
	if cfg == nil {
		return defaultLogDefinition()
	}
	def := &LogDefinition{}
	// 设置日志输出类型
	// LogToFile - 输出到文件
	// LogToStdout - 输出到标准输出
	// LogToStderr - 输出到标准错误输出
	switch cfg.Output {
	case "file":
		def.OutputType = LogToFile
	case "stderr":
		def.OutputType = LogToStderr
	default:
		// 不设置默认全部输出到标准输出设备
		def.OutputType = LogToStdout
	}
	// 设置日志等级
	switch cfg.LogLevel {
	case "debug":
		def.Level = LevelDebug
	case "info":
		def.Level = LevelInfo
	case "warn":
		def.Level = LevelWarn
	case "error":
		def.Level = LevelError
	case "fatal":
		def.Level = LevelFatal
	default:
		def.Level = LevelError
	}
	// 设置flag，此处的内容与golang中的log包的相关设置相同
	// 此处暂不支持自定义设置，如果需要设置需要在此方法之外（前）自行设定
	def.Flags = Ldate | Ltime | Lmicroseconds | Lshortfile
	// 设置日志文件存储目录，仅当输出类型为 LogToFile 有效
	initLogDir(cfg.LogPath)
	def.Dir = cfg.LogPath
	// 设置日志文件切割类型
	switch cfg.Rotate {
	case "none":
		def.RotateType = RotateNone
	case "year":
		def.RotateType = RotateByYear
	case "month":
		def.RotateType = RotateByMonth
	case "date":
		def.RotateType = RotateByDate
	case "hour":
		def.RotateType = RotateByHour
	default:
		def.RotateType = RotateByDate
	}

	// 设置日志文件名前缀，仅当输出类型为 LogToFile 有效
	def.FilePrefix = cfg.LogPrefix

	// 调用栈信息
	def.LogStack = false
	if cfg.LogStackLevel != "none" {
		def.LogStack = true
		def.LogStackLevel = numLogLevel(cfg.LogStackLevel)
	}

	// 日志打印
	def.ColorfulPrint = cfg.ColorfulPrint

	return def
}

// getLogRotateTimeFmt 获取日志文件切割时间格式
func (def *LogDefinition) GetLogRotateTimeFmt() string {
	var timeFmt string
	switch logRotateType {
	case RotateByYear:
		timeFmt = "2006"
	case RotateByMonth:
		timeFmt = "200601"
	case RotateByDate:
		timeFmt = "20060102"
	case RotateByHour:
		timeFmt = "2006010215"
	}
	return timeFmt
}

// getLogFilePath 获取日志文件路径
func (def *LogDefinition) GetLogFilePath() string {
	if def.RotateType == RotateNone {
		return fmt.Sprintf("%s/%s%s.log", def.Dir, def.FilePrefix, "all")
	}
	logRotateTimeFmt := getLogRotateTimeFmt()
	return fmt.Sprintf("%s/%s%s.log", def.Dir, def.FilePrefix, time.Now().Format(logRotateTimeFmt))
}
