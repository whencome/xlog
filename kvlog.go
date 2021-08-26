package xlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func unescapeUnicode(raw []byte) ([]byte, error) {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(string(raw)), `\\u`, `\u`, -1))
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

func getKey(v string) string {
	if !strings.Contains(v, "\"") {
		return v
	}
	return strings.ReplaceAll(v, "\"", "\\\"")
}

func getVal(v interface{}) string {
	b, e := json.Marshal(v)
	if e != nil {
		return ""
	}
	nb, e := unescapeUnicode(b)
	if e != nil {
		return ""
	}
	return string(nb)
}

type KVData struct {
	pairs map[string]interface{}
	keys  []string
}

func NewKVData() *KVData {
	return &KVData{
		pairs: make(map[string]interface{}),
		keys:  make([]string, 0),
	}
}

func (d *KVData) Put(k string, v interface{}) {
	// 检查k是否存在
	if _, ok := d.pairs[k]; !ok {
		d.keys = append(d.keys, k)
	}
	d.pairs[k] = v
}

func (d *KVData) GetLines() string {
	if len(d.keys) == 0 {
		return ""
	}
	buf := bytes.Buffer{}
	for i, k := range d.keys {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(k)
		buf.WriteString(":")
		buf.WriteString(fmt.Sprintf("%s", d.pairs[k]))
	}
	return buf.String()
}

func (d *KVData) GetRaw() string {
	if len(d.keys) == 0 {
		return ""
	}
	buf := bytes.Buffer{}
	for i, k := range d.keys {
		if i > 0 {
			buf.WriteString(";")
		}
		buf.WriteString(k)
		buf.WriteString(":")
		buf.WriteString(fmt.Sprintf("%s", d.pairs[k]))
	}
	return buf.String()
}

func (d *KVData) GetJson() string {
	if len(d.keys) == 0 {
		return ""
	}
	buf := bytes.Buffer{}
	buf.WriteString("{")
	for i, k := range d.keys {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString("\"")
		buf.WriteString(getKey(k))
		buf.WriteString("\"")
		buf.WriteString(":")
		buf.WriteString(getVal(d.pairs[k]))
	}
	buf.WriteString("}")
	return buf.String()
}

type KVLogger struct {
	data       *KVData
	recordTime bool      // 是否记录耗时
	startTime  time.Time // 开始时间
	endTime    time.Time // 结束时间
}

func NewKVLogger() *KVLogger {
	return &KVLogger{
		data:       NewKVData(),
		recordTime: false,
	}
}

func NewTimerKVLogger() *KVLogger {
	l := &KVLogger{
		data:       NewKVData(),
		recordTime: true,
	}
	l.startTime = time.Now()
	return l
}

func (l *KVLogger) Put(k string, v interface{}) {
	l.data.Put(k, v)
}

func (l *KVLogger) fill()  {
	if !l.recordTime {
		return
	}
	l.endTime = time.Now()
	// 添加时间信息
	l.Put("@start_time", l.startTime.Format("2006-01-02 15:04:05.000"))
	l.Put("@end_time", l.endTime.Format("2006-01-02 15:04:05.000"))
	l.Put("@time_cost", fmt.Sprintf("%.3f ms", float64(l.endTime.UnixNano()-l.startTime.UnixNano())/1e6))
}

func (l *KVLogger) OutputLines(logger *StdLogger) {
	l.fill()
	data := l.data.GetLines()
	if len(data) == 0 {
		return
	}
	logger.Rawln(data)
	l.Reset()
}

func (l *KVLogger) OutputRaw(logger *StdLogger) {
	l.fill()
	data := l.data.GetRaw()
	if len(data) == 0 {
		return
	}
	logger.Rawln(data)
	l.Reset()
}

func (l *KVLogger) OutputJson(logger *StdLogger) {
	l.fill()
	data := l.data.GetJson()
	if len(data) == 0 {
		return
	}
	logger.Rawln(data)
	l.Reset()
}

func (l *KVLogger) Reset() {
	l.data = NewKVData()
	if l.recordTime {
		l.startTime = time.Now()
	}
	return
}
