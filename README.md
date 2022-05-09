# xlog
一个简单的日志工具，支持按自动按时间记录到不同的文件，支持按日志等级记录日志，支持自定义输出，支持同时将多个不同的日志输出到不同的地方

## 特点

* 支持自定义日志输出，包括输出到文件（可以自定义日志目录）、标准输出、标准错误输出
* （记录日志到文件）支持根据定义的时间间隔将日志输出到不同的文件，比如按小时、按日、按月、按年输出到不同的日志文件
* 支持日志等级，分为：debug、info、warn、error、fatal，可以在代码中记录详细的日志，后期发布直接修改日志等级即可实现控制日志输出，不用修改代码
* 部分功能直接与golang原生的log相同（直接拷贝的相关代码），比如flag
* 2021.05.07 支持将不同的日志以不同的方式输出到不同的地方
* 2022.05.09 支持快速返回默认配置（不使用配置文件），支持同时注册多个日志对象

## 使用示例

### step1. 初始化日志信息（全局默认）

* 方式1： 使用默认配置（无需增加配置文件，代码简洁），默认篇日志默认为debug模式，所有日志输出到标准输出（控制台）
```go
// 使用默认配置
cfg := xlog.DefaultConfig()
// 如果需要修改日志级别或者输出位置，可以在这里修改cfg的参数
xlog.Init(cfg)
```
如果只是用于调试，将日志输出到控制台，可以直接使用：
```go
// 此方式主要用于开发阶段，与上面两句的效果相同
xlog.InitDefault()
```

* 方式2：单独设置参数（下面的方式只影响默认日志对象，如果需要其它日志对象，需要单独创建并注册）
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
xlog.Init(nil)
```

也可以通过配置直接快速初始化日志对象，如：
```go
cfg := &xlog.Config{
    LogPath : "/home/logs/test",
    LogPrefix : "order_",
    Output : "file",
    LogLevel : "debug",
    Rotate : "date",
    LogStackLevel : "error", // 如果设置为none表示始终不记录stack信息
}
xlog.Init(cfg)
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

## 修改日志

* 2021.05.07 支持将不同的日志输出到不同的地方，示例如下：
```go
cfgs := map[string]*Config{
    "order" : &xlog.Config{
        LogPath : "/home/logs/test",
        LogPrefix : "order_",
        Output : "file",
        LogLevel : "debug",
        Rotate : "date",
        LogStackLevel : "error",
    },
    "curl" : &xlog.Config{
        LogPath : "/home/logs/test",
        LogPrefix : "curl_",
        Output : "file",
        LogLevel : "debug",
        Rotate : "date",
        LogStackLevel : "error",
    },
    "api" : &xlog.Config{
        LogPath : "/home/logs/test",
        LogPrefix : "api_",
        Output : "file",
        LogLevel : "debug",
        Rotate : "date",
        LogStackLevel : "error",
    },
}
// 方式1：注册logger
for k, cfg := range cfgs {
    xlog.Register(k, cfg)
}
// 方式2：直接注册多个
// xlog.RegisterMany(cfgs)

// 写日志
xlog.Use("order").Info("order log content")
xlog.Use("api").Debug("api log content")
xlog.Use("curl").Error("curl log content")
```