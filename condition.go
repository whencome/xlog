package gomodel

import (
    "github.com/whencome/gomodel/builder"
)

// Condition 定义一个sql条件组
type Condition struct {
    Logic       string  // 条件逻辑，AND / OR
    Conds       []*Condition  // 条件数组
    condData    []map[string]interface{} // 条件组数据，优先级高于Conds
}

// NewAndCondition 创建一个And条件组
func NewAndCondition() *Condition {
    return &Condition{
        Logic : "AND",
        Conds : make([]*Condition, 0),
        condData : make([]map[string]interface{}, 0),
    }
}

// NewOrCondition 创建一个Or条件组
func NewOrCondition() *Condition {
    return &Condition{
        Logic : "OR",
        Conds : make([]*Condition, 0),
        condData : make([]map[string]interface{}, 0),
    }
}

// Add 添加一个条件
func (c *Condition) Add(field string, val interface{}) {
    c.condData = append(c.condData, map[string]interface{}{field:val})
}

// AddBatch 批量添加条件
func (c *Condition) AddBatch(batchConds []map[string]interface{}) {
    for _, bc := range batchConds {
        c.condData = append(c.condData, bc)
    }
}

// AddCondition 田间一个条件组
func (c *Condition) AddCondition(cc *Condition) {
    c.Conds = append(c.Conds, cc)
}

// Build 构造条件
func (c *Condition) Build() (string, error) {
    patch, err := builder.NewConditionBuilder().Build(c.condData, c.Logic)
    if err != nil {
        return "", err
    }
    if len(c.Conds) > 0 {
        for _, cond := range c.Conds {
            p, err := cond.Build()
            if err != nil {
                return "", err
            }
            if patch != "" {
                patch += " " + c.Logic + " "
            }
            patch += " (" + p + ") "
        }
    }
    return patch, nil
}

// BuildCondition 根据任意条件参数构造条件
func BuildCondition(conds interface{}) (string, error) {
    if conds == nil {
        return "", nil
    }
    // 根据类型采取不同的构建方式
    condWhere, ok := conds.(*Condition)
    if ok {
        return condWhere.Build()
    }
    return builder.NewConditionBuilder().Build(conds, "AND")
}


