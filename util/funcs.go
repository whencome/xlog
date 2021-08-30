package util

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/whencome/xlog/def"
)

// NumLogLevel 获取日志等级的对应的数字
func NumLogLevel(l string) int {
	num := def.LevelError
	switch l {
	case def.LogLevelDebug:
		num = def.LevelDebug
	case def.LogLevelInfo:
		num = def.LevelInfo
	case def.LogLevelWarn:
		num = def.LevelWarn
	case def.LogLevelError:
		num = def.LevelError
	case def.LogLevelFatal:
		num = def.LevelFatal
	}
	return num
}

// GetLogRotateTimeFmt 获取日志文件切割时间格式
func GetLogRotateTimeFmt(logRotateType int) string {
	var timeFmt string
	switch logRotateType {
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

// InitLogDir 初始化日志目录
func InitLogDir(path string) (bool, error) {
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
func getLogFilePath(logRotateType int, logDir, logFilePrefix string) string {
	if logRotateType == def.RotateNone {
		return fmt.Sprintf("%s/%s%s.log", logDir, logFilePrefix, "all")
	}
	logRotateTimeFmt := GetLogRotateTimeFmt(logRotateType)
	return fmt.Sprintf("%s/%s%s.log", logDir, logFilePrefix, time.Now().Format(logRotateTimeFmt))
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func Itoa(buf *[]byte, i int, wid int) {
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

// FormatLogPrefix 格式化日志前缀
func FormatLogPrefix(buf *[]byte, logFlags int, t time.Time, level string, file string, line int) {
	// 时间
	if logFlags & (def.Ldate | def.Ltime | def.Lmicroseconds) != 0 {
		if logFlags & def.LUTC != 0 {
			t = t.UTC()
		}
		if logFlags & def.Ldate != 0 {
			year, month, day := t.Date()
			Itoa(buf, year, 4)
			*buf = append(*buf, '/')
			Itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			Itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if logFlags & (def.Ltime | def.Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			Itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			Itoa(buf, min, 2)
			*buf = append(*buf, ':')
			Itoa(buf, sec, 2)
			if logFlags & def.Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				Itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	// 日志等级(一定会输出)
	*buf = append(*buf, '[')
	*buf = append(*buf, strings.ToUpper(level)...)
	*buf = append(*buf, "] "...)
	// 文件
	if logFlags & (def.Lshortfile | def.Llongfile) != 0 {
		if logFlags & def.Lshortfile != 0 {
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
		Itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}

// IsNil 判断给定的值是否为nil
func IsNil(i interface{}) bool {
	ret := i == nil
	// 需要进一步做判断
	if !ret {
		vi := reflect.ValueOf(i)
		kind := reflect.ValueOf(i).Kind()
		if kind == reflect.Slice ||
			kind == reflect.Map ||
			kind == reflect.Chan ||
			kind == reflect.Interface ||
			kind == reflect.Func ||
			kind == reflect.Ptr {
			return vi.IsNil()
		}
	}
	return ret
}