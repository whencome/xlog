package xlog

// Config 定义日志配置
type Config struct {
	LogPath   string `json:"log_path"`   // 定义日志根路径
	LogPrefix string `json:"log_prefix"` // 日志文件前缀
	Output    string `json:"output"`     // 日志输出类型,file,stdout,stderr
	LogLevel  string `json:"log_level"`  // 日志等级，可取值:debug,info,warn,error,fatal
	Rotate    string `json:"rotate"`     // 日志切割类型,可取值：none,year,month,date,hour
}