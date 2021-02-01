package gomodel

import (
	"database/sql"
)

// Commander 执行者，用于执行数据库查询等操作
type Commander struct{
	inTrans		bool // 是否在执行事务中
	Command     string  // 需要执行的SQL
	Settings    *Options // 相关配置
	conn        *sql.DB  // 数据库连接
	tx 			*sql.Tx  // 事务
}

// NewCommander 创建一个新的执行者对象
func NewCommander(opts *Options) *Commander {
	if opts == nil {
		opts = NewDefaultOptions()
	}
	return &Commander{
		inTrans:false,
		Settings:opts,
		conn:nil,
	}
}

// SetOptions 设置选项参数
func (c *Commander) SetOptions(opts *Options) *Commander {
	c.Settings = opts
	return c
}

// Connect 设置数据库连接
func (c *Commander) Connect(conn *sql.DB) *Commander {
	if conn != nil {
		c.conn = conn
	}
	return c
}

// BeginTransaction 开启事务
func (c *Commander) BeginTransaction() error {
	tx, err := c.conn.Begin()
	if err != nil {
		return err
	}
	c.tx = tx
	return nil
}

// Commit 提交事务
func (c *Commander) Commit() error {
	if !c.inTrans {
		return nil
	}
	return c.tx.Commit()
}

// Rollback 回滚事务
func (c *Commander) Rollback() error {
	if !c.inTrans {
		return nil
	}
	return c.tx.Rollback()
}

// Execute 执行SQL命令
func (c *Commander) Execute(command string, args ...interface{}) (sql.Result, error) {
	if c.inTrans {
		return c.tx.Exec(command, args...)
	}
	return c.conn.Exec(command, args...)
}

// RawQuery 执行原始的查询
func (c *Commander) RawQuery(command string, args ...interface{}) (*sql.Rows, error) {
	if c.inTrans {
		return c.tx.Query(command, args...)
	}
	return c.conn.Query(command, args...)
}

// Query 查询满足条件的全部数据
func (c *Commander) Query(command string, args ...interface{}) (*QueryResult, error) {
	result := NewQueryResult()
	rows, err := c.RawQuery(command, args...)
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

// QueryRow 查询单行数据
func (c *Commander) QueryRow(command string, args ...interface{}) (map[string]string, error) {
	rows, err := c.RawQuery(command, args...)
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	// 创建临时切片用于保存数据
	row := make([]interface{}, len(columns))
	// 创建存储数据的字节切片2维数组data
	tmpData := make([][]byte, len(columns))
	for i, _ := range row {
		row[i] = &tmpData[i]
	}
	// 开始读取数据
	data := make(map[string]string)
	err = rows.Scan(row...)
	if err != nil {
		return nil, err
	}
	for i, v := range row {
		k := columns[i]
		if v == nil {
			data[k] = ""
		} else {
			data[k] = string(*(v.(*[]uint8)))
		}
	}
	// 返回查询结果
	return data, nil
}

// QueryScalar 查询单个值
func (c *Commander) QueryScalar(command string, args ...interface{}) (string, error) {
	rows, err := c.RawQuery(command, args...)
	if err != nil {
		return "", err
	}
	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}
	// 创建临时切片用于保存数据
	row := make([]interface{}, len(columns))
	// 创建存储数据的字节切片2维数组data
	tmpData := make([][]byte, len(columns))
	for i, _ := range row {
		row[i] = &tmpData[i]
	}
	// 开始读取数据
	data := make(map[string]string)
	err = rows.Scan(row...)
	if err != nil {
		return "", err
	}
	for i, v := range row {
		k := columns[i]
		if v == nil {
			data[k] = ""
		} else {
			data[k] = string(*(v.(*[]uint8)))
		}
	}
	// 查询第一个字段
	firstColumn := columns[0]
	// 返回查询结果
	return data[firstColumn], nil
}