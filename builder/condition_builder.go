package builder

import (
    "bytes"
    "fmt"
    "strings"

    "github.com/whencome/gomodel/utils"
)

// ConditionBuilder 条件构造器，构造SQL查询条件
type ConditionBuilder struct {}

// NewConditionBuilder 创建一个新的条件构造器
func NewConditionBuilder() *ConditionBuilder {
    return &ConditionBuilder{}
}

// Build 构造SQL条件
func (cb *ConditionBuilder) Build(conds interface{}, logic string) (string, error) {
    return cb.buildCondition(conds, logic)
}

// addSQLCondition 写入SQL查询条件
func (cb *ConditionBuilder) addSQLCondition(buffer *bytes.Buffer, logic string, sqlPatch string) {
    if buffer.Len() > 0 {
        buffer.WriteString(" ")
        buffer.WriteString(logic)
        buffer.WriteString(" ")
    }
    buffer.WriteString(" ( ")
    buffer.WriteString(sqlPatch)
    buffer.WriteString(" ) ")
}

// BuildCondition 构造逻辑查询条件
func (cb *ConditionBuilder) buildCondition(conds interface{}, logic string) (string, error) {
    // 如果条件为空，则认为查询全部
    if conds == nil {
        return "", nil
    }
    // 构造查询条件
    // 查询逻辑，logic = AND/OR
    logic = strings.ToUpper(strings.TrimSpace(logic))
    if logic == "" {
        logic = "AND"
    }
    buffer := &bytes.Buffer{}
    // 检查条件是否为已经写好的SQL段
    switch conds.(type) {
    // 查询内容为纯粹的sql段，无需处理
    case string:
        sqlPatch := string(conds.(string))
        cb.addSQLCondition(buffer, logic, sqlPatch)
    case []uint8:
        sqlPatch := string(conds.([]uint8))
        cb.addSQLCondition(buffer, logic, sqlPatch)
    case []rune:
        sqlPatch := string(conds.([]rune))
        cb.addSQLCondition(buffer, logic, sqlPatch)
    case []interface{}:
        condList := conds.([]interface{})
        if len(condList) == 0 {
            break
        }
        for _, v := range condList {
            sqlPatch, err := cb.buildCondition(v, logic)
            if err != nil {
                return "", err
            }
            cb.addSQLCondition(buffer, logic, sqlPatch)
        }
    case map[string]interface{}:
        mapCond := conds.(map[string]interface{})
        sqlPatch, err := cb.buildMapCondition(mapCond, logic)
        if err != nil {
            return "", err
        }
        cb.addSQLCondition(buffer, logic, sqlPatch)
    case []map[string]interface{}:
        listMapConds := conds.([]map[string]interface{})
        for _, mapConds := range listMapConds {
            sqlPatch, err := cb.buildMapCondition(mapConds, logic)
            if err != nil {
                return "", err
            }
            cb.addSQLCondition(buffer, logic, sqlPatch)
        }
    default:
        return "", fmt.Errorf("unsupported condition data type %T of %#v", conds, conds)
    }
    return buffer.String(), nil
}

// buildMapCondition 根据map参数构造
func (cb *ConditionBuilder) buildMapCondition(conds map[string]interface{}, logic string) (string, error) {
    buffer := &bytes.Buffer{}
    for k, v := range conds {
        k = strings.TrimSpace(k)
        mapLogic := strings.ToUpper(k)
        // K如果是指定查询逻辑
        if mapLogic == "AND" || mapLogic == "OR" {
            sqlPatch, err := cb.buildCondition(v, mapLogic)
            if err != nil {
            }
            cb.addSQLCondition(buffer, mapLogic, sqlPatch)
            continue
        }
        // K如果是指定查询字段
        field := k
        matchLogic := "="
        logicSep := strings.Index(k, " ")
        if logicSep > 0 {
            field = k[:logicSep]
            matchLogic = k[logicSep+1:]
        }
        sqlPatch, err := cb.buildMatchLogicQuery(field, matchLogic, v)
        if err != nil {
            return "", err
        }
        cb.addSQLCondition(buffer, logic, sqlPatch)
        continue
    }
    return buffer.String(), nil
}

// buildMatchLogicQuery 构造匹配条件
func (cb *ConditionBuilder) buildMatchLogicQuery(field, matchLogic string, value interface{}) (string, error) {
    matchLogic = strings.ToUpper(strings.TrimSpace(matchLogic))
    if matchLogic == "" {
        matchLogic = "="
    }
    field = strings.ReplaceAll(field, "`", "")
    switch matchLogic {
    case "=","!=",">",">=","<","<=","<>","LIKE","NOT LIKE","IS":
        fieldValue := utils.NewValue(value).SQLValue()
        return fmt.Sprintf("%s %s %s", cb.quoteField(field), matchLogic, fieldValue), nil
    case "IN","NOT IN":
        inVales := transValue2Array(value)
        if len(inVales) == 0 {
            return "", fmt.Errorf("[%s] value not qualified", matchLogic)
        }
        fieldValues := make([]string, 0)
        for _, v := range inVales {
            vv := utils.NewValue(v).SQLValue()
            fieldValues = append(fieldValues, vv)
        }
        return fmt.Sprintf("%s %s (%s)", cb.quoteField(field), matchLogic, strings.Join(fieldValues, ", ")), nil
    case "BETWEEN", "NOT BETWEEN":
        betweenVales := transValue2Array(value)
        if len(betweenVales) != 2 {
            return "", fmt.Errorf("[%s] value count not qualified", matchLogic)
        }
        firstV := utils.NewValue(betweenVales[0]).SQLValue()
        secondV := utils.NewValue(betweenVales[1]).SQLValue()
        return fmt.Sprintf("%s %s %s AND %s", cb.quoteField(field), matchLogic, firstV, secondV), nil
    default:
        return "", fmt.Errorf("unsupported match logic %s", matchLogic)
    }
}

// quoteField 对字段进行处理
func (cb *ConditionBuilder) quoteField(field string) string {
    if strings.Contains(field, "`") {
        return field
    }
    if !strings.Contains(field, ".") {
        return fmt.Sprintf("`%s`", field)
    }
    fieldParts := strings.Split(field, ".")
    return fmt.Sprintf("`%s`", strings.Join(fieldParts, "`,`"))
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


