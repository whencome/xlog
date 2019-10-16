package xlog

import (
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	// 设置日志输出类型
	// LogToFile - 输出到文件
	// LogToStdout - 输出到标准输出
	// LogToStderr - 输出到标准错误输出
	SetLogOutputType(LogToStdout)
	// 设置日志等级，开发时可详尽记录日志，发布线上是修改此处的等级即可
	// 因此，此处的值建议放到配置文件中
	SetLogLevel(LogLevelDebug)
	// 设置flag，此处的内容与golang中的log包的相关设置相同
	// 注意此处的包是xlog，不是log
	SetLogFlags(Ldate | Ltime | Lmicroseconds | Llongfile)

	// 测试日志输出
	Info("info log")
	Infof("now is %s", time.Now().Format("2006-01-02 15:04:05.0000"))
	Debug("debug log")
	Warn("warn log")
	Error("error log")
}
