package sqlchecker

import (
    "errors"
    "strings"

    "github.com/whencome/sqlparser"
)

// SQLCheckResult 定义SQL检查对象
type SQLCheckResult struct {
    SQL string    // 执行的SQL
    Command string  // SQL命令
    IsPassed bool   // 检查是否通过
    Error error   // 时报错误信息，默认为nil
}

// NewSQLCheckResult 创建一个空的检查结果
func newSQLCheckResult() *SQLCheckResult {
    return &SQLCheckResult{
        SQL:"",
        Command:"",
        IsPassed:false,
        Error:nil,
    }
}

// Check 执行SQL检查
func Check(query string) *SQLCheckResult {
    result := newSQLCheckResult()
    // 格式化查询SQL
    query, err := normalizeQuery(query)
    if err != nil {
        result.Error = err
        return result
    }
    // 执行SQL检查
    result.SQL = query
    _, err = sqlparser.Parse(query)
    if err != nil {
        result.Error = err
        return result
    }
    // 命令
    t := sqlparser.Preview(query)
    command := sqlparser.StmtType(t)
    if command == "OTHER" {
        command = strings.ToUpper(query[:strings.Index(query, " ")])
    }
    result.Command = strings.ToUpper(command)
    result.IsPassed = true
    // 返回检查结果
    return result
}

// normalizeQuery 格式化查询，去除换行符，移除多余的空格（顺序不能乱）
func normalizeQuery(query string) (string, error) {
    // 移除换行符
    query = strings.ReplaceAll(query, "\n", " ")
    query = strings.ReplaceAll(query, "\r", " ")
    // 去除前后空格
    query = strings.TrimSpace(query)
    if query == "" {
        return "", errors.New("invalid or empty query")
    }
    // 如果传入多条SQL，则只取第一条
    sqlPieces, err := sqlparser.SplitStatementToPieces(query)
    if err != nil {
        return "", err
    }
    // 获取第一条SQL语句
    sql := sqlPieces[0]
    sql = strings.TrimSpace(sql)
    // 返回处理后的结果
    return sql, nil
}
