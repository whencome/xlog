# xlog
һ���򵥵���־���ߣ�֧�ְ��Զ���ʱ���¼����ͬ���ļ���֧�ְ���־�ȼ���¼��־��֧���Զ��������֧��ͬʱ�������ͬ����־�������ͬ�ĵط�

## �ص�

* ֧���Զ�����־���������������ļ��������Զ�����־Ŀ¼������׼�������׼�������
* ����¼��־���ļ���֧�ָ��ݶ����ʱ��������־�������ͬ���ļ������簴Сʱ�����ա����¡������������ͬ����־�ļ�
* ֧����־�ȼ�����Ϊ��debug��info��warn��error��fatal�������ڴ����м�¼��ϸ����־�����ڷ���ֱ���޸���־�ȼ�����ʵ�ֿ�����־����������޸Ĵ���
* ���ֹ���ֱ����golangԭ����log��ͬ��ֱ�ӿ�������ش��룩������flag
* 2021.05.07 ֧�ֽ���ͬ����־�Բ�ͬ�ķ�ʽ�������ͬ�ĵط�

## ʹ��ʾ��

### step1. ��ʼ����־��Ϣ��ȫ��Ĭ�ϣ�

```go
// ������־�������
// LogToFile - ������ļ�
// LogToStdout - �������׼���
// LogToStderr - �������׼�������
xlog.SetLogOutputType(xlog.LogToFile)
// ������־�ȼ�������ʱ���꾡��¼��־�������������޸Ĵ˴��ĵȼ�����
// ��ˣ��˴���ֵ����ŵ������ļ���
xlog.SetLogLevel(xlog.LogLevelDebug)
// ����flag���˴���������golang�е�log�������������ͬ
// ע��˴��İ���xlog������log
xlog.SetLogFlags(xlog.Ldate | xlog.Ltime | xlog.Lmicroseconds | xlog.Llongfile)
// ������־�ļ��洢Ŀ¼�������������Ϊ LogToFile ��Ч
xlog.SetLogDir("/home/logs/test")
// ������־�ļ��и�����
xlog.SetLogRotateType(xlog.RotateByDate)
// ������־�ļ���ǰ׺�������������Ϊ LogToFile ��Ч
xlog.SetLogFilePrefix("test_")
xlog.Init(nil)
```

Ҳ����ͨ������ֱ�ӿ��ٳ�ʼ����־�����磺
```go
cfg := &xlog.Config{
    LogPath : "/home/logs/test",
    LogPrefix : "order_",
    Output : "file",
    LogLevel : "debug",
    Rotate : "date",
    LogStackLevel : "error", // �������Ϊnone��ʾʼ�ղ���¼stack��Ϣ
}
xlog.Init(cfg)
```

### step2. �ڴ����м�¼��־

```go
// ��¼debug��־���ȼ����
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

## �޸���־

* 2021.05.07 ֧�ֽ���ͬ����־�������ͬ�ĵط���ʾ�����£�
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
// ע��logger
for k, cfg := range cfgs {
    xlog.Register(k, cfg)
}

// д��־
xlog.Use("order").Info("order log content")
xlog.Use("api").Debug("api log content")
xlog.Use("curl").Error("curl log content")
```