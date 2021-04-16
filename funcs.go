package gomodel

import (
	"bytes"
	"fmt"
	"strings"
)

// quote 对字段进行处理
func quote(field string) string {
	field = strings.TrimSpace(field)
	// 自带反引号或者内置方法调用，不需要quote操作
	if strings.Contains(field, "`") || strings.Contains(field, "(") {
		return field
	}
	// a.b格式
	if strings.Contains(field, ".") {
		fieldParts := strings.Split(field, ".")
		size := len(fieldParts)
		if size == 1 {
			return fmt.Sprintf("`%s`", fieldParts[0])
		}
		if fieldParts[size-1] == "*" {
			return fmt.Sprintf("`%s`", strings.Join(fieldParts[:size-1], "`.`")) + ".*"
		}
		return fmt.Sprintf("`%s`", strings.Join(fieldParts, "`.`"))
	}
	// 检查是否包含别名
	if strings.Contains(field, " ") {
		fieldParts := strings.Split(field, " ")
		size := len(fieldParts)
		if size == 1 {
			return fmt.Sprintf("`%s`", fieldParts[0])
		}
		var quoteRs bytes.Buffer
		for i := 0; i < size; i++ {
			if i == 0 || i == size - 1 {
				quoteRs.WriteString(fmt.Sprintf("`%s` ", fieldParts[i]))
			} else {
				quoteRs.WriteString(fieldParts[i])
				quoteRs.WriteString(" ")
			}
		}
		return quoteRs.String()
	}
	// 普通字段/表名
	return fmt.Sprintf("`%s`", field)
}

// transValue2Array 将值转换成数组
func transValue2Array(value interface{}) []interface{} {
	inVales := make([]interface{}, 0)
	switch value.(type) {
	case []interface{}:
		inVales = value.([]interface{})
	case []string:
		strArrs := value.([]string)
		for _, sa := range strArrs {
			inVales = append(inVales, sa)
		}
	case []int:
		intArrs := value.([]int)
		for _, i := range intArrs {
			inVales = append(inVales, i)
		}
	case []int64:
		intArrs := value.([]int64)
		for _, i := range intArrs {
			inVales = append(inVales, i)
		}
	case []int32:
		intArrs := value.([]int32)
		for _, i := range intArrs {
			inVales = append(inVales, i)
		}
	case []int16:
		intArrs := value.([]int16)
		for _, i := range intArrs {
			inVales = append(inVales, i)
		}
	case []int8:
		intArrs := value.([]int8)
		for _, i := range intArrs {
			inVales = append(inVales, i)
		}
	case []uint:
		intArrs := value.([]uint)
		for _, i := range intArrs {
			inVales = append(inVales, i)
		}
	case []uint64:
		intArrs := value.([]uint64)
		for _, i := range intArrs {
			inVales = append(inVales, i)
		}
	case []uint32:
		intArrs := value.([]uint32)
		for _, i := range intArrs {
			inVales = append(inVales, i)
		}
	case []uint16:
		intArrs := value.([]uint16)
		for _, i := range intArrs {
			inVales = append(inVales, i)
		}
	case []uint8:
		intArrs := value.([]uint8)
		for _, i := range intArrs {
			inVales = append(inVales, i)
		}
	case []float64:
		floadArrs := value.([]float64)
		for _, f := range floadArrs {
			inVales = append(inVales, f)
		}
	case []float32:
		floadArrs := value.([]float32)
		for _, f := range floadArrs {
			inVales = append(inVales, f)
		}
	}
	return inVales
}
