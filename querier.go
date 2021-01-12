package gomodel

import (
    "bytes"
    "database/sql"
    "errors"
    "fmt"
    "strings"

    "github.com/whencome/xlog"

    "github.com/whencome/gomodel/sqlchecker"
)

// 定义联表方式
const (
    innerJoin = "INNER"
    leftJoin = "LEFT"
    rightJoin = "RIGHT"
)

/************************************************************
 ******              SECTION OF JOIN TABLES             *****
 ************************************************************/
// joinTable 定义联表类型
type joinTable struct {
    table       string  // 表名
    condition   string  // 连接条件,只支持字符串
    joinType    string  // 连接方式，inner，left，right
}

// newInnerJoinTable 创建一个内联表
func newInnerJoinTable(tblName string, onCond string) *joinTable {
    return &joinTable{
        table : tblName,
        condition : onCond,
        joinType:innerJoin,
    }
}

// newRightJoinTable 创建一个右联表
func newRightJoinTable(tblName string, onCond string) *joinTable {
    return &joinTable{
        table : tblName,
        condition : onCond,
        joinType:rightJoin,
    }
}

// newLeftJoinTable 创建一个左联表
func newLeftJoinTable(tblName string, onCond string) *joinTable {
    return &joinTable{
        table : tblName,
        condition : onCond,
        joinType:leftJoin,
    }
}


/************************************************************
 ******             SECTION OF QUERY RESULT             *****
 ************************************************************/
// QueryResult 保存一个查询结果（不支持分页）
type QueryResult struct {
    TotalCount      int // 记录总数
    Offset          int // 偏移量，用于分页处理
    RowsCount       int // 当前查询的记录数量
    Columns			[]string   // 用于单独保存字段，以解决显示结果字段顺序不正确的问题
    Rows 			[]map[string]string  // 查询结果，一切皆字符串
}

// NewQueryResult 创建一个新的查询结果
func NewQueryResult() *QueryResult {
    return &QueryResult{
        TotalCount:0,
        Offset:0,
        RowsCount:0,
        Columns:make([]string, 0),
        Rows:make([]map[string]string, 0),
    }
}


/************************************************************
 ******                SECTION OF QUERIER               *****
 ************************************************************/
// Querier 查询对象
type Querier struct {
    queryMaps   map[string]interface{}
    joinTables []*joinTable  // 联表信息
    QuerySQL string // 查询SQL
    Settings *Options // 是否开启查询前的SQL语法检测
}

// NewQuerier 创建一个空的Querier
func NewQuerier() *Querier {
    return &Querier{
        queryMaps:map[string]interface{}{
            "fields" : "",
            "table" : "",
            "join_tables":make([]*joinTable, 0),
            "where":nil,
            "having":nil,
            "order_by":"",
            "group_by":"",
            "offset":0,
            "limit":-1,
        },
        joinTables : make([]*joinTable, 0),
        QuerySQL:"",
        Settings:NewDefaultOptions(),  // 设置一个默认参数配置
    }
}

// NewRawQuerier 根据查询SQL创建一个Querier
func NewRawQuerier(querySQL string) *Querier {
    q := NewQuerier()
    q.QuerySQL = querySQL
    return q
}

// NewModelQuerier 创建一个指定Model的查询对象
func NewModelQuerier(m Modeler) *Querier {
    q := NewQuerier()
    q.queryMaps["table"] = m.GetTableName()
    q.queryMaps["fields"] = "*"
    return q
}

// SetOptions 设置选项配置
func (q *Querier) SetOptions(opts *Options) *Querier {
    if opts == nil {
        return q
    }
    q.Settings = opts
    return q
}

// Select 设置查询字段,fields为以“,”连接的字段列表
func (q *Querier) Select(fields string) *Querier {
     q.queryMaps["fields"] = fields
     return q
}

// From 选择查询的表
func (q *Querier) From(tblName string) *Querier {
    q.queryMaps["table"] = tblName
    return q
}

// Join 设置内联表
func (q *Querier) Join(tblName string, onCond string) *Querier {
    q.joinTables = append(q.joinTables, newInnerJoinTable(tblName, onCond))
    return q
}

// LeftJoin 设置左联表
func (q *Querier) LeftJoin(tblName string, onCond string) *Querier {
    q.joinTables = append(q.joinTables, newLeftJoinTable(tblName, onCond))
    return q
}

// RightJoin 设置右联表
func (q *Querier) RightJoin(tblName string, onCond string) *Querier {
    q.joinTables = append(q.joinTables, newRightJoinTable(tblName, onCond))
    return q
}

// Where 设置查询条件
func (q *Querier) Where(cond interface{}) *Querier {
    q.queryMaps["where"] = cond
    return q
}

// OrderBy 设置排序方式
func (q *Querier) OrderBy(orderBy string) *Querier {
    q.queryMaps["order_by"] = orderBy
    return q
}

// GroupBy 设置分组方式
func (q *Querier) GroupBy(groupBy string) *Querier {
    q.queryMaps["group_by"] = groupBy
    return q
}

// Having 设置分组过滤条件
func (q *Querier) Having(cond interface{}) *Querier {
    q.queryMaps["having"] = cond
    return q
}

// Offset 设置查询偏移量
func (q *Querier) Offset(num int) *Querier {
    q.queryMaps["offset"] = num
    return q
}

// Limit 设置查询数量
func (q *Querier) Limit(num int) *Querier {
    q.queryMaps["limit"] = num
    return q
}

// buildCondition 构造查询条件
func (q *Querier) buildCondition() (string, error) {
    where, ok := q.queryMaps["where"]
    if !ok || where == nil {
        return "", nil
    }
    // 根据类型采取不同的构建方式
    condWhere, ok := where.(*Condition)
    if ok {
        return condWhere.Build()
    }
    return NewConditionBuilder().Build(q.queryMaps["where"], "AND")
}

// checkQuery 检查查询语法以及命令是否支持
func (q *Querier) checkQuery(querySQL string) error {
    // 如果没有开启前置检查，则不进行语法检查
    if !q.Settings.IsPreQuerySyntaxCheckEnabled() {
        return nil
    }
    // 检查查询SQL是否存在语法问题
    chkRs := sqlchecker.Check(querySQL)
    if !chkRs.IsPassed {
        if chkRs.Error != nil {
            return chkRs.Error
        }
        return errors.New("SQL syntax check failed：" + querySQL)
    }
    // 检查是否为select语句
    if chkRs.Command != "SELECT" {
        return fmt.Errorf("[%s] command not supported in querier：%s", chkRs.Command, querySQL)
    }
    return nil
}

// buildQuery 构造查询语句
func (q *Querier) buildQuery() error {
    if q.QuerySQL != "" {
        // 检查查询SQL是否存在语法问题
        err := q.checkQuery(q.QuerySQL)
        if err != nil {
           return err
        }
        return nil
    }
    querySQL := bytes.Buffer{}
    querySQL.WriteString("SELECT ")

    // 查询字段
    fields := NewValue(q.queryMaps["fields"]).String()
    if fields == "" {
        fields = "*"
    }
    querySQL.WriteString(fields)

    // 表
    tableName := NewValue(q.queryMaps["table"]).String()
    if tableName == "" {
        return errors.New("query table not specified")
    }
    querySQL.WriteString(" FROM ")
    querySQL.WriteString(quote(tableName))
    
    // 检查联表信息
    if len(q.joinTables) > 0 {
        for _, joinTbl := range q.joinTables {
            if strings.TrimSpace(joinTbl.table) == "" {
                return errors.New("empty join table name")
            }
            if strings.TrimSpace(joinTbl.condition) == "" {
                return errors.New("join condition empty")
            }
            querySQL.WriteString(" ")
            querySQL.WriteString(joinTbl.joinType)
            querySQL.WriteString(" JOIN ")
            querySQL.WriteString(joinTbl.table)
            querySQL.WriteString(" ON ")
            querySQL.WriteString(joinTbl.condition)
        }
    }

    // 查询条件
    condition, err := q.buildCondition()
    if err != nil {
        return err
    }
    if condition != "" {
        querySQL.WriteString(" WHERE ")
        querySQL.WriteString(condition)
    }

    // 检查是否对查询进行分组
    groupBy := NewValue(q.queryMaps["group_by"]).String()
    if groupBy != "" {
        querySQL.WriteString(" GROUP BY ")
        querySQL.WriteString(groupBy)
        // 检查是否有分组过滤
        having, err := NewConditionBuilder().Build(q.queryMaps["having"], "AND")
        if err != nil {
            return err
        }
        if having != "" {
            querySQL.WriteString(" HAVING ")
            querySQL.WriteString(having)
        }
    }

    // 设置排序
    orderBy := NewValue(q.queryMaps["order_by"]).String()
    if orderBy != "" {
        querySQL.WriteString(" ORDER BY ")
        querySQL.WriteString(orderBy)
    }

    // 设置limit信息
    offset := NewValue(q.queryMaps["offset"]).Int64()
    limitNum := NewValue(q.queryMaps["limit"]).Int64()
    if limitNum > 0 {
        querySQL.WriteString(fmt.Sprintf(" LIMIT %d, %d", offset, limitNum))
    }

    // 返回查询SQL
    q.QuerySQL = querySQL.String()
    return nil
}

// buildCountQuery 构造count查询语句，用于统计查询数据的数量
func (q *Querier) buildCountQuery() (string, error) {
    // 根据原始查询语句构造Count语句
    if q.QuerySQL != "" && q.queryMaps["where"] == nil {
        return q.buildCountQueryFromRawQuery()
    }
    // 根据条件构造Count语句
    return q.buildCountQueryFromConditions()
}

// buildCountQueryFromConditions 根据条件构造count语句
func (q *Querier) buildCountQueryFromConditions() (string, error) {
    querySQL := bytes.Buffer{}
    querySQL.WriteString("SELECT COUNT(0)")

    // 表
    tableName := NewValue(q.queryMaps["table"]).String()
    if tableName == "" {
        return "", errors.New("query table not specified")
    }
    querySQL.WriteString(" FROM ")
    querySQL.WriteString(quote(tableName))

    // 检查联表信息
    if len(q.joinTables) > 0 {
        for _, joinTbl := range q.joinTables {
            if strings.TrimSpace(joinTbl.table) == "" {
                return "", errors.New("empty join table name")
            }
            if strings.TrimSpace(joinTbl.condition) == "" {
                return "", errors.New("join condition empty")
            }
            querySQL.WriteString(" ")
            querySQL.WriteString(joinTbl.joinType)
            querySQL.WriteString(" JOIN ")
            querySQL.WriteString(joinTbl.table)
            querySQL.WriteString(" ON ")
            querySQL.WriteString(joinTbl.condition)
        }
    }

    // 查询条件
    condition, err := q.buildCondition()
    if err != nil {
        return "", err
    }
    if condition != "" {
        querySQL.WriteString(" WHERE ")
        querySQL.WriteString(condition)
    }

    // 返回查询SQL
    return querySQL.String(), nil
}

// buildCountQueryFromRawQuery 根据原始查询构造count语句
func (q *Querier) buildCountQueryFromRawQuery() (string, error) {
    if q.QuerySQL == "" {
        return "", errors.New("query sql can not be empty")
    }
    // 检查查询SQL
    err := q.checkQuery(q.QuerySQL)
    if err != nil {
        return "", err
    }
    // 构造count语句
    querySQL := bytes.Buffer{}
    querySQL.WriteString("SELECT COUNT(0) ")

    // 先简单处理(逻辑上有问题，后续再解决)
    lowerQuerySQL := strings.ToLower(q.QuerySQL)
    fromPos := strings.Index(lowerQuerySQL, " from ")
    // wherePos := strings.Index(lowerQuerySQL, "where")
    orderPos := strings.LastIndex(lowerQuerySQL, " order ")
    groupPos := strings.LastIndex(lowerQuerySQL, " group ")
    limitPos := strings.LastIndex(lowerQuerySQL, " limit ")
    endPos := 0
    if orderPos > 0 {
        endPos = orderPos
    }
    if groupPos > 0 && groupPos < endPos {
        endPos = groupPos
    }
    if limitPos > 0 && limitPos < endPos {
        endPos = limitPos
    }
    countPart := ""
    if endPos > 0 && endPos > fromPos {
        countPart = q.QuerySQL[fromPos:endPos]
    } else {
        countPart = q.QuerySQL[fromPos:]
    }
    querySQL.WriteString(countPart)

    // 返回查询SQL
    return querySQL.String(), nil
}

// Query 执行查询,此处返回为切片，以保证返回值结果顺序与查询字段顺序一致
func (q *Querier) Query(conn *sql.DB) (*QueryResult, error) {
    // 构建查询
    err := q.buildQuery()
    if err != nil {
        xlog.Errorf("[querier] build query failed: %s", err)
        return nil, err
    }
    xlog.Debugf("[querier] Query : %s", q.QuerySQL)
    // 执行查询
    result := NewQueryResult()
    rows, err := conn.Query(q.QuerySQL)
    if err != nil {
        return nil, err
    }
    result.Columns, err = rows.Columns()
    if err != nil {
        return nil, err
    }
    // 创建临时切片用于保存数据
    row := make([]interface{}, len(result.Columns))
    // 创建存储数据的字节切片2维数组data
    tmpData := make([][]byte, len(result.Columns))
    for i, _ := range row {
        // 将字节切片地址赋值给临时切片,这样row才是真正存放数据
        row[i] = &tmpData[i]
    }
    // 开始读取数据
    count := 0
    for rows.Next() {
        err = rows.Scan(row...)
        if err != nil {
            return nil, err
        }
        data := make(map[string]string)
        for i, v := range row {
            k := result.Columns[i]
            if v == nil {
                data[k] = ""
            } else {
                data[k] = string(*(v.(*[]uint8)))
            }
        }
        result.Rows = append(result.Rows, data)
        count++
    }
    result.TotalCount = count
    result.RowsCount = count
    // 返回查询结果
    return result, nil
}

// 查询记录总数
func (q *Querier) queryTotalCount(conn *sql.DB) (int, error) {
    countQuery, err := q.buildCountQuery()
    if err != nil {
        return 0, err
    }
    xlog.Debugf("querier queryTotalCount : %s", countQuery)
    countRow := conn.QueryRow(countQuery)
    var totalCount int
    err = countRow.Scan(&totalCount)
    if err != nil {
        return 0, err
    }
    return totalCount, nil
}

// Count 查询记录总数
func (q *Querier) Count(conn *sql.DB) (int, error) {
    return q.queryTotalCount(conn)
}

// QueryPage 查询分页信息
func (q *Querier) QueryPage(conn *sql.DB, page, pageSize int) (*QueryResult, error) {
    // 将page和pageSize转换成limit
    offset := (page - 1) * pageSize
    q.Offset(offset).Limit(pageSize)
    // 开始查询，查询分两步
    // 1. 查询总数量
    totalCount, err := q.queryTotalCount(conn)
    if err != nil {
        return nil, err
    }
    // 2. 查询当前分页的数据
    queryResult, err := q.Query(conn)
    if err != nil {
        return nil, err
    }
    // 重置总数
    queryResult.TotalCount = totalCount
    // 返回查询结果
    return queryResult, nil
}

// QueryRow 查询单条记录
func (q *Querier) QueryRow(conn *sql.DB) (map[string]string, error) {
    q.Limit(1)
    queryResult, err := q.Query(conn)
    if err != nil {
        return nil, err
    }
    if queryResult.RowsCount == 0 {
        return nil, nil
    }
    return queryResult.Rows[0], nil
}

// QueryScalar 查询单个值
func (q *Querier) QueryScalar(conn *sql.DB) (string, error) {
    queryResult, err := q.Query(conn)
    if err != nil {
        return "", err
    }
    if queryResult.RowsCount == 0 ||
        len(queryResult.Columns) == 0 {
        return "", nil
    }
    firstField := queryResult.Columns[0]
    v, _ := queryResult.Rows[0][firstField]
    return v, nil
}
