package def

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
