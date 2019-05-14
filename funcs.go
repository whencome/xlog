package xlog

import (
	"fmt"
	"time"
	"strings"
)

// getLogCutTimeFmt 获取日志文件切割时间格式
func getLogCutTimeFmt() string {
	var timeFmt string
	switch logCutType {
	case CutByYear:
		timeFmt = "2006"
	case CutByMonth:
		timeFmt = "200601"
	case CutByDate:
		timeFmt = "20060102"
	case CutByHour:
		timeFmt = "2006010203"
	}
	return timeFmt
}

// getLogFilePath 获取日志文件路径
func getLogFilePath() string {
	if logCutType == CutNone {
		return fmt.Sprintf("%s/%s%s.log", LogDir, LogFilePrefix, "all")
	}
	logCutTimeFmt := getLogCutTimeFmt()
	return fmt.Sprintf("%s/%s%s.log", LogDir, LogFilePrefix, time.Now().Format(logCutTimeFmt))
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
	if logFlags & (Ldate|Ltime|Lmicroseconds) != 0 {
		if logFlags & LUTC != 0 {
			t = t.UTC()
		}
		if logFlags & Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if logFlags & (Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if logFlags & Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond() / 1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	// 日志等级(一定会输出)
	*buf = append(*buf, '[')
	*buf = append(*buf, strings.ToUpper(level)...)
	*buf = append(*buf, "] "...)
	// 文件
	if logFlags & (Lshortfile|Llongfile) != 0 {
		if logFlags & Lshortfile != 0 {
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

