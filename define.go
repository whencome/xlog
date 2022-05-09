package xlog

import (
    "fmt"
    "os"
    "time"

    "github.com/whencome/xlog/def"
    "github.com/whencome/xlog/util"
)

// Config 定义日志配置
type Config struct {
    LogPath       string `json:"log_path" toml:"log_path" yaml:"log_path"`                      // 定义日志根路径
    LogPrefix     string `json:"log_prefix" toml:"log_prefix" yaml:"log_prefix"`                // 日志文件前缀
    Output        string `json:"output" toml:"output" yaml:"output"`                            // 日志输出类型,file,stdout,stderr
    LogLevel      string `json:"log_level" toml:"log_level" yaml:"log_level"`                   // 日志等级，可取值:debug,info,warn,error,fatal
    Rotate        string `json:"rotate" toml:"rotate" yaml:"rotate"`                            // 日志切割类型,可取值：none,year,month,date,hour
    LogStackLevel string `json:"log_stack_level" toml:"log_stack_level" yaml:"log_stack_level"` // 记录调用栈信息的日志等级
    ColorfulPrint bool   `json:"colorful_print" toml:"colorful_print" yaml:"colorful_print"`    // 是否开启彩色打印，仅适用于标准输出，不适用于文件输出
    Switch        string `json:"switch" toml:"switch" yaml:"switch"`                            // 开关，off-关闭，on/empty-开启
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
    Disabled      bool     // 是否禁用
}

// 返回一个默认的日志定义
func DefaultConfig() *Config {
    c := &Config{
        LogPath:       "",
        LogPrefix:     "",
        Output:        "stdout",
        LogLevel:      "debug",
        Rotate:        "date",
        LogStackLevel: "error",
        ColorfulPrint: true,
        Switch:        "on",
    }
    return c
}

// 返回一个默认的日志定义
func defaultLogDefinition() *LogDefinition {
    d := &LogDefinition{}
    d.Dir = LogDir
    d.Level = logLevel
    d.FilePrefix = LogFilePrefix
    d.Flags = logFlags
    d.Output = logOutput
    d.OutputType = logOutputType
    d.RotateType = logRotateType
    d.LogStack = logStack
    d.LogStackLevel = logStackLevel
    d.ColorfulPrint = colorfulPrint
    d.Disabled = false
    return d
}

// 根据配置返回一个日志定义
func newLogDefinition(cfg *Config) *LogDefinition {
    // 没有配置则忽略
    if cfg == nil {
        return defaultLogDefinition()
    }
    d := &LogDefinition{}
    // 设置日志输出类型
    // LogToFile - 输出到文件
    // LogToStdout - 输出到标准输出
    // LogToStderr - 输出到标准错误输出
    switch cfg.Output {
    case "file":
        d.OutputType = def.LogToFile
    case "stderr":
        d.OutputType = def.LogToStderr
    default:
        // 不设置默认全部输出到标准输出设备
        d.OutputType = def.LogToStdout
    }
    // 设置日志等级
    switch cfg.LogLevel {
    case "debug":
        d.Level = def.LevelDebug
    case "info":
        d.Level = def.LevelInfo
    case "warn":
        d.Level = def.LevelWarn
    case "error":
        d.Level = def.LevelError
    case "fatal":
        d.Level = def.LevelFatal
    default:
        d.Level = def.LevelError
    }
    // 设置flag，此处的内容与golang中的log包的相关设置相同
    // 此处暂不支持自定义设置，如果需要设置需要在此方法之外（前）自行设定
    d.Flags = def.Ldate | def.Ltime | def.Lmicroseconds | def.Lshortfile
    // 设置日志文件存储目录，仅当输出类型为 LogToFile 有效
    _, _ = util.InitLogDir(cfg.LogPath)
    d.Dir = cfg.LogPath
    // 设置日志文件切割类型
    switch cfg.Rotate {
    case "none":
        d.RotateType = def.RotateNone
    case "year":
        d.RotateType = def.RotateByYear
    case "month":
        d.RotateType = def.RotateByMonth
    case "date":
        d.RotateType = def.RotateByDate
    case "hour":
        d.RotateType = def.RotateByHour
    default:
        d.RotateType = def.RotateByDate
    }

    // 设置日志文件名前缀，仅当输出类型为 LogToFile 有效
    d.FilePrefix = cfg.LogPrefix

    // 调用栈信息
    d.LogStack = false
    if cfg.LogStackLevel != "none" {
        d.LogStack = true
        d.LogStackLevel = util.NumLogLevel(cfg.LogStackLevel)
    }

    // 日志打印
    d.ColorfulPrint = cfg.ColorfulPrint

    // 日志开关
    if cfg.Switch == "off" {
        d.Disabled = true
    } else {
        d.Disabled = false
    }

    return d
}

// getLogRotateTimeFmt 获取日志文件切割时间格式
func (d *LogDefinition) GetLogRotateTimeFmt() string {
    var timeFmt string
    switch d.RotateType {
    case def.RotateByYear:
        timeFmt = "2006"
    case def.RotateByMonth:
        timeFmt = "200601"
    case def.RotateByDate:
        timeFmt = "20060102"
    case def.RotateByHour:
        timeFmt = "2006010215"
    }
    return timeFmt
}

// getLogFilePath 获取日志文件路径
func (d *LogDefinition) GetLogFilePath() string {
    if d.RotateType == def.RotateNone {
        return fmt.Sprintf("%s/%s%s.log", d.Dir, d.FilePrefix, "all")
    }
    logRotateTimeFmt := util.GetLogRotateTimeFmt(d.RotateType)
    return fmt.Sprintf("%s/%s%s.log", d.Dir, d.FilePrefix, time.Now().Format(logRotateTimeFmt))
}
