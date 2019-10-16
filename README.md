# xlog
一个简单的日志工具，支持按自动按时间记录到不同的文件，支持按日志等级记录日志，支持自定义输出

## 特点

* 支持自定义日志输出，包括输出到文件（可以自定义日志目录）、标准输出、标准错误输出
* （记录日志到文件）支持根据定义的时间间隔将日志输出到不同的文件，比如按小时、按日、按月、按年输出到不同的日志文件
* 支持日志等级，分为：debug、info、warn、error、fatal，可以在代码中记录详细的日志，后期发布直接修改日志等级即可实现控制日志输出，不用修改代码
* 部分功能直接与golang原生的log相同（直接拷贝的相关代码），比如flag

## 使用示例

### step1. 初始化日志信息（全局）

```go
// 设置日志输出类型
// LogToFile - 输出到文件
// LogToStdout - 输出到标准输出
// LogToStderr - 输出到标准错误输出
xlog.SetLogOutputType(xlog.LogToFile)
// 设置日志等级，开发时可详尽记录日志，发布线上是修改此处的等级即可
// 因此，此处的值建议放到配置文件中
xlog.SetLogLevel(xlog.LogLevelDebug)
// 设置flag，此处的内容与golang中的log包的相关设置相同
// 注意此处的包是xlog，不是log
xlog.SetLogFlags(xlog.Ldate | xlog.Ltime | xlog.Lmicroseconds | xlog.Llongfile)
// 设置日志文件存储目录，仅当输出类型为 LogToFile 有效
xlog.SetLogDir("/home/logs/test")
// 设置日志文件切割类型
xlog.SetLogRotateType(xlog.RotateByDate)
// 设置日志文件名前缀，仅当输出类型为 LogToFile 有效
xlog.SetLogFilePrefix("test_")
```

### step2. 在代码中记录日志

```go
// 记录debug日志，等级最低
xlog.Debug("debug log info here")
xlog.Debugln("debug log info here")
xlog.Debugf("debug log info here: %s", "test")
// more log functions
//  info
xlog.Info("info log info here")
xlog.Infoln("info log info here")
xlog.Infof("info log info here: %s", "test")
// warn
xlog.Warn("warn log info here")
xlog.Warnln("warn log info here")
xlog.Warnf("warn log info here: %s", "test")
// error
xlog.Error("error log info here")
xlog.Errorln("error log info here")
xlog.Errorf("error log info here: %s", "test")
// fatal
xlog.Fatal("fatal log info here")
xlog.Fatalln("fatal log info here")
xlog.Fatalf("fatal log info here: %s", "test")
// panic
xlog.Panic("panic log info here")
xlog.Panicln("panic log info here")
xlog.Panicf("panic log info here: %s", "test")
```