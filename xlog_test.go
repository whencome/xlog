package xlog

// go test -v xlog_test.go stdlogger.go xlog.go log.go define.go

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/whencome/xlog/def"
)

func TestLog(t *testing.T) {
	// 设置日志输出类型
	// LogToFile - 输出到文件
	// LogToStdout - 输出到标准输出
	// LogToStderr - 输出到标准错误输出
	SetLogOutputType(def.LogToStdout)
	// 设置日志等级，开发时可详尽记录日志，发布线上是修改此处的等级即可
	// 因此，此处的值建议放到配置文件中
	SetLogLevel(def.LogLevelDebug)
	// 设置flag，此处的内容与golang中的log包的相关设置相同
	// 注意此处的包是xlog，不是log
	SetLogFlags(def.Ldate | def.Ltime | def.Lmicroseconds | def.Llongfile)

	Init(nil)

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
			Output : "stdout",
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
	wg.Add(3)
	for i:=0; i<3; i++ {
		go func(k string) {
			defer wg.Done()
			for j:=0; j<100; j++ {
				Use(k).Infof("[%s] Now is %s", k, time.Now().Format("2006-01-02 15:04:05"))
			}
		}(keys[i])
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

func TestKVLogger(t *testing.T) {
	logger := NewKVLogger(os.Stdout)
	logger.Put("query", "select * from table where id > 0 limit 20")
	logger.Put("start", "2021-08-17 10:00:00.023")
	logger.Put("end", "2021-08-17 10:00:00.114")
	logger.Put("cost", "0.91s")
	logger.Write()
}

func TestTimerKVLogger(t *testing.T) {
	logger := NewTimerKVLogger(os.Stdout)
	logger.Put("query", "select * from table where id > 0 limit 20")
	logger.Put("start", "2021-08-17 10:00:00.023")
	logger.Put("end", "2021-08-17 10:00:00.114")
	logger.Put("cost", "0.91s")
	logger.Put("message", "测试一下中文内容")
	logger.Write()
}

func TestTimerKVLogger1(t *testing.T) {
	logger := NewTimerKVLogger(os.Stdout)
	logger.Put("query", "select * from table where id > 0 limit 20")
	logger.Put("start", "2021-08-17 10:00:00.023")
	logger.Put("end", "2021-08-17 10:00:00.114")
	logger.Put("cost", "0.91s")
	logger.Put("message", "测试一下中文内容")
	logger.Put("content", "a string with \", \\, --, # and so on")
	logger.Write()
}

func TestTimerKVLogger2(t *testing.T) {
	logger := NewTimerKVLogger(os.Stdout)
	logger.Put("query", "select * from table where id > 0 limit 20")
	logger.Put("start", "2021-08-17 10:00:00.023")
	logger.Put("end", "2021-08-17 10:00:00.114")
	logger.Put("cost", "0.91s")
	logger.Put("message", "测试一下中文内容")
	logger.Put("content", []interface{}{1,2,3,"what this?", map[string]int{"smith":78, "jack":99}})
	logger.Write()
}

func TestNilWriterTimerKVLogger(t *testing.T) {
	logger := NewTimerKVLogger(MustUse("db"))
	logger.Put("query", "select * from table where id > 0 limit 20")
	logger.Put("start", "2021-08-17 10:00:00.023")
	logger.Put("end", "2021-08-17 10:00:00.114")
	logger.Put("cost", "0.91s")
	logger.Put("message", "测试一下中文内容")
	logger.Write()
}

func TestBufLog(t *testing.T) {
	cfgs := map[string]*Config{
		"order" : &Config{
			LogPath : "/home/logs/test",
			LogPrefix : "buf_order_",
			Output : "stdout",
			LogLevel : "debug",
			Rotate : "date",
			LogStackLevel : "error",
		},
		"curl" : &Config{
			LogPath : "/home/logs/test",
			LogPrefix : "buf_curl_",
			Output : "file",
			LogLevel : "debug",
			Rotate : "date",
			LogStackLevel : "error",
		},
		"api" : &Config{
			LogPath : "/home/logs/test",
			LogPrefix : "buf_api_",
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
	wg.Add(3)
	for i:=0; i<3; i++ {
		go func(k string) {
			bufLog := NewBufLogger(Use(k))
			bufLog.SetBufferSize(1024)
			defer wg.Done()
			defer bufLog.Close()
			for j:=0; j<100; j++ {
				bufLog.Infof("[%s] /buflog/ Now is %s", k, time.Now().Format("2006-01-02 15:04:05"))
			}
		}(keys[i])
	}
	wg.Wait()
	fmt.Println("buflog test finished")
}

func TestStackBufLog(t *testing.T) {
	// 设置日志输出类型
	// LogToFile - 输出到文件
	// LogToStdout - 输出到标准输出
	// LogToStderr - 输出到标准错误输出
	SetLogOutputType(def.LogToStdout)
	// 设置日志等级，开发时可详尽记录日志，发布线上是修改此处的等级即可
	// 因此，此处的值建议放到配置文件中
	SetLogLevel(def.LogLevelDebug)
	// 设置flag，此处的内容与golang中的log包的相关设置相同
	// 注意此处的包是xlog，不是log
	SetLogFlags(def.Ldate | def.Ltime | def.Lmicroseconds | def.Llongfile)

	Init(nil)

	bufLog := NewStackBufLogger(defaultLogger)

	// 测试日志输出
	bufLog.Info("info log")
	bufLog.Infof("now is %s", time.Now().Format("2006-01-02 15:04:05.0000"))
	bufLog.Debug("debug log")
	bufLog.Warn("warn log")
	bufLog.Error("error log")
	bufLog.Close()
}

func TestCloseLogger(t *testing.T) {
	// 设置日志输出类型
	// LogToFile - 输出到文件
	// LogToStdout - 输出到标准输出
	// LogToStderr - 输出到标准错误输出
	SetLogOutputType(def.LogToStdout)
	// 设置日志等级，开发时可详尽记录日志，发布线上是修改此处的等级即可
	// 因此，此处的值建议放到配置文件中
	SetLogLevel(def.LogLevelDebug)
	// 设置flag，此处的内容与golang中的log包的相关设置相同
	// 注意此处的包是xlog，不是log
	SetLogFlags(def.Ldate | def.Ltime | def.Lmicroseconds | def.Llongfile)

	Init(nil)

	// 测试日志输出
	Info("log before closed")
	Use("default").Close()
	fmt.Printf("start to test closed logger")
	for i := 0; i < 100; i++ {
		Infof("log after closed of [%d]", i)
	}
	fmt.Printf("test finished")
}

// 测试更新配置
func TestSwitchCfg(t *testing.T) {
	cfg := &Config{
		LogPath : "/home/logs/test",
		LogPrefix : "buf_api_",
		Output : "stdout",
		LogLevel : "debug",
		Rotate : "date",
		LogStackLevel : "error",
		ColorfulPrint:true,
		Switch:"on",
}
	logCat := "api"
	switches := map[int]string{
		0:"on",
		1:"off",
	}
	i := 0
	ticker := time.NewTicker(time.Second * 3)
	for {
		<-ticker.C
		idx := i % 2
		sw := switches[idx]
		cfg.Switch = sw
		fmt.Printf("[%d] cfg: %+v\n", i, cfg)
		Clear()
		Register(logCat, cfg)
		l := Use(logCat)
		l.Debugf("TestSwitchCfg --> debug log")
		l.Infof("TestSwitchCfg --> info log")
		l.Warnf("TestSwitchCfg --> warn log")
		l.Errorf("TestSwitchCfg --> error log")
		i++
		if i == 10 {
			break
		}
	}
}