package xlog

import (
	"sync"
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


func TestMultiLog(t *testing.T) {
	cfgs := map[string]*Config{
		"order" : &Config{
			LogPath : "/home/logs/test",
			LogPrefix : "order_",
			Output : "file",
			LogLevel : "debug",
			Rotate : "date",
			LogStackLevel : "error",
		},
		"curl" : &Config{
			LogPath : "/home/logs/test",
			LogPrefix : "curl_",
			Output : "stdout",
			LogLevel : "debug",
			Rotate : "date",
			LogStackLevel : "error",
		},
		"api" : &Config{
			LogPath : "/home/logs/test",
			LogPrefix : "api_",
			Output : "file",
			LogLevel : "debug",
			Rotate : "date",
			LogStackLevel : "error",
		},
	}
	// 注册logger
	for k, cfg := range cfgs {
		Infof("register logger [%s]", k)
		Register(k, cfg)
	}
	keys := []string{"order", "curl", "api"}
	// 写日志
	wg := sync.WaitGroup{}
	for i:=0; i<3; i++ {
		wg.Add(1)
		key := keys[i]
		go func(k string) {
			defer wg.Done()
			for j:=0; j<100000; j++ {
				Use(k).Infof("[%s] Now is %s", k, time.Now().Format("2006-01-02 15:04:05"))
			}
		}(key)
	}
	wg.Wait()
	Info("test finished")
}

func TestLogRaw(t *testing.T) {
	Raw(1,2,3,4)
	Raw(5,6,7)
	Rawln(8,9,10)
	Rawf("now is: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}

func TestLogJson(t *testing.T) {
	data := map[string]interface{}{
		"path" : "/home/logs/test",
		"prefix" : "curl_",
		"output" : "stdout",
		"level" : "debug",
		"rotate" : "date",
		"stack_level" : "error",
	}
	Json(data)
}


