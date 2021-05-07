package xlog

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// getLogRotateTimeFmt 获取日志文件切割时间格式
func getLogRotateTimeFmt() string {
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

// initLogDir 初始化日志目录
func initLogDir(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		// 尝试创建目录
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, err
}

// getLogFilePath 获取日志文件路径
func getLogFilePath() string {
	if logRotateType == RotateNone {
		return fmt.Sprintf("%s/%s%s.log", LogDir, LogFilePrefix, "all")
	}
	logRotateTimeFmt := getLogRotateTimeFmt()
	return fmt.Sprintf("%s/%s%s.log", LogDir, LogFilePrefix, time.Now().Format(logRotateTimeFmt))
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// formatLogPrefix 格式化日志前缀
func formatLogPrefix(buf *[]byte, t time.Time, level string, file string, line int) {
	// 时间
	if logFlags&(Ldate|Ltime|Lmicroseconds) != 0 {
		if logFlags&LUTC != 0 {
			t = t.UTC()
		}
		if logFlags&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if logFlags&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if logFlags&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	// 日志等级(一定会输出)
	*buf = append(*buf, '[')
	*buf = append(*buf, strings.ToUpper(level)...)
	*buf = append(*buf, "] "...)
	// 文件
	if logFlags&(Lshortfile|Llongfile) != 0 {
		if logFlags&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}
