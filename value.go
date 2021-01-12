package gomodel

import (
    "fmt"
    "strconv"
    "strings"
    "unicode/utf8"

    "github.com/axgle/mahonia"
)

// 定义数据表特殊字符替换映射表
var sqlSpecialCharMaps = []map[string]string {
    {"old":`\`, "new":`\\`},
    {"old":`'`, "new":`\'`},
    {"old":`"`, "new":`\"`},
}

// EscapeSqlValue 转义数据库中的特殊字符，暂时只处理常见内容
func EscapeSqlValue(str string) string {
    // 先检查是否是utf8，不是则先转换
    if !utf8.ValidString(str) {
        utf8Encoder := mahonia.NewEncoder("UTF-8")
        str = utf8Encoder.ConvertString(str)
    }
    for _, repl := range sqlSpecialCharMaps {
        str = strings.ReplaceAll(str, repl["old"], repl["new"])
    }
    return str
}

// Value 定义一个通用的Value结构体，用于统一处理类型转换
type Value struct {
    Data interface{}
}

// NewValue Create a new *Value
func NewValue(val interface{}) *Value {
    return &Value{
        Data: val,
    }
}

// String Get a string value of val
func (val *Value) String() string {
    var strVal = ""
    switch val.Data.(type) {
    case int, int8, int16, int32, int64:
        n := val.Int64()
        strVal = strconv.FormatInt(n, 10)
    case uint, uint8, uint16, uint32, uint64:
        n := val.Uint64()
        strVal = strconv.FormatUint(n, 10)
    case float32:
        strVal = strconv.FormatFloat(float64(val.Data.(float32)), 'f', -1, 64)
    case float64:
        strVal = strconv.FormatFloat(val.Data.(float64), 'f', -1, 64)
    case string:
        strVal = val.Data.(string)
    case []byte:
        strVal = string(val.Data.([]byte))
    case []rune:
        strVal = string(val.Data.([]rune))
    case bool:
        strVal = strconv.FormatBool(val.Data.(bool))
    default:
        if val.Data == nil {
            strVal = ""
        } else {
            strVal = fmt.Sprint(val.Data)
        }
    }
    return strVal
}

// SQLValue 获取插入数据库需要的值
func (val *Value) SQLValue() string {
    var strVal = ""
    switch val.Data.(type) {
    case int, int8, int16, int32, int64:
        n := val.Int64()
        strVal = strconv.FormatInt(n, 10)
    case uint, uint8, uint16, uint32, uint64:
        n := val.Uint64()
        strVal = strconv.FormatUint(n, 10)
    case float32:
        strVal = strconv.FormatFloat(float64(val.Data.(float32)), 'f', -1, 64)
    case float64:
        strVal = strconv.FormatFloat(val.Data.(float64), 'f', -1, 64)
    case string:
        strVal = val.Data.(string)
        strVal = fmt.Sprintf("'%s'", EscapeSqlValue(strVal))
    case []byte:
        strVal = string(val.Data.([]byte))
        strVal = fmt.Sprintf("'%s'", EscapeSqlValue(strVal))
    case []rune:
        strVal = string(val.Data.([]rune))
        strVal = fmt.Sprintf("'%s'", EscapeSqlValue(strVal))
    case bool:
        strVal = "0"
        if val.Data.(bool) {
            strVal = "1"
        }
    default:
        strVal = fmt.Sprint(val.Data)
        strVal = fmt.Sprintf("'%s'", EscapeSqlValue(strVal))
    }
    // 返回结果
    return strVal
}

// Int64 get int64 value
func (val *Value) Int64() int64 {
    switch val.Data.(type) {
    case int:
        return int64(val.Data.(int))
    case int8:
        return int64(val.Data.(int8))
    case int16:
        return int64(val.Data.(int16))
    case int32:
        return int64(val.Data.(int32))
    case int64:
        return val.Data.(int64)
    case uint:
        return int64(val.Data.(uint))
    case uint8:
        return int64(val.Data.(uint8))
    case uint16:
        return int64(val.Data.(uint16))
    case uint32:
        return int64(val.Data.(uint32))
    case uint64:
        return int64(val.Data.(uint64))
    case float32:
        return int64(val.Data.(float32))
    case float64:
        return int64(val.Data.(float64))
    case string:
        n, err := strconv.ParseInt(string(val.Data.(string)), 10, 64)
        if err != nil {
            return 0
        }
        return n
    case []byte:
        n, err := strconv.ParseInt(string(val.Data.([]byte)), 10, 64)
        if err != nil {
            return 0
        }
        return n
    case []rune:
        n, err := strconv.ParseInt(string(val.Data.([]rune)), 10, 64)
        if err != nil {
            return 0
        }
        return n
    case bool:
        intVal := int64(0)
        if val.Data.(bool) {
            intVal = 1
        }
        return intVal
    default:
        return 0
    }
    return 0
}

// Uint64 get uint64 value
func (val *Value) Uint64() uint64 {
    switch val.Data.(type) {
    case int:
        return uint64(val.Data.(int))
    case int8:
        return uint64(val.Data.(int8))
    case int16:
        return uint64(val.Data.(int16))
    case int32:
        return uint64(val.Data.(int32))
    case int64:
        return uint64(val.Data.(int64))
    case uint:
        return uint64(val.Data.(uint))
    case uint8:
        return uint64(val.Data.(uint8))
    case uint16:
        return uint64(val.Data.(uint16))
    case uint32:
        return uint64(val.Data.(uint32))
    case uint64:
        return val.Data.(uint64)
    case float32:
        return uint64(val.Data.(float32))
    case float64:
        return uint64(val.Data.(float64))
    case string:
        n, err := strconv.ParseUint(string(val.Data.(string)), 10, 64)
        if err != nil {
            return 0
        }
        return n
    case []byte:
        n, err := strconv.ParseUint(string(val.Data.([]byte)), 10, 64)
        if err != nil {
            return 0
        }
        return n
    case []rune:
        n, err := strconv.ParseUint(string(val.Data.([]rune)), 10, 64)
        if err != nil {
            return 0
        }
        return n
    case bool:
        intVal := uint64(0)
        if val.Data.(bool) {
            intVal = 1
        }
        return intVal
    default:
        return 0
    }
    return 0
}

// Float64 get float64 value
func (val *Value) Float64() float64 {
    switch val.Data.(type) {
    case int:
        return float64(val.Data.(int))
    case int8:
        return float64(val.Data.(int8))
    case int16:
        return float64(val.Data.(int16))
    case int32:
        return float64(val.Data.(int32))
    case int64:
        return float64(val.Data.(int64))
    case uint:
        return float64(val.Data.(uint))
    case uint8:
        return float64(val.Data.(uint8))
    case uint16:
        return float64(val.Data.(uint16))
    case uint32:
        return float64(val.Data.(uint32))
    case uint64:
        return float64(val.Data.(uint64))
    case float32:
        return float64(val.Data.(float32))
    case float64:
        return float64(val.Data.(float64))
    case string:
        n, err := strconv.ParseFloat(string(val.Data.(string)), 64)
        if err != nil {
            return 0
        }
        return n
    case []byte:
        n, err := strconv.ParseFloat(string(val.Data.([]byte)), 64)
        if err != nil {
            return 0
        }
        return n
    case []rune:
        n, err := strconv.ParseFloat(string(val.Data.([]rune)), 64)
        if err != nil {
            return 0
        }
        return n
    case bool:
        n := float64(0)
        if val.Data.(bool) {
            n = 1
        }
        return n
    default:
        return 0
    }
    return 0
}

// Boolean get bool value
func (val *Value) Boolean() bool {
    v := val.Uint64()
    if v > 0 {
        return true
    }
    return false
}
