package gomodel

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/whencome/xlog"
)

type ShardingModelManager struct {
	*ModelManager
	Sharding 		int64
}

// NewShardingModelManager 创建一个ShardingModelManager
func NewShardingModelManager(m Modeler, opts *Options) *ShardingModelManager {
	mm := NewModelManager(m)
	mm.SetOptions(opts)
	return &ShardingModelManager{
		ModelManager : mm,
		Sharding: 0,
	}
}

// UseSharding 设置使用的sharding值
func (m *ShardingModelManager) UseSharding(v int64) *ShardingModelManager {
	m.Sharding = v
	return m
}

// GetSharding 获取Sharding值
func (m *ShardingModelManager) GetSharding() (int64, int64, error) {
	if !m.Settings.EnableSharding || m.Settings.DbShardingNum <= 0 || m.Settings.TableShardingNum <= 0 {
		return 0, 0, fmt.Errorf("SHARDING_UNAVAILABLE")
	}
	if m.Sharding <= 0 {
		return 0, 0, fmt.Errorf("SHARDING_VALUE_INVALID")
	}
	tblSharding := m.Sharding % m.Settings.TableShardingNum
	dbSharding := int64(math.Floor(float64(tblSharding) / float64(m.Settings.DbShardingNum)))
	return tblSharding, dbSharding, nil
}

// GetTableName 获取Model对应的数据表名
func (m *ShardingModelManager) GetTableName() string {
	if m.Model == nil {
		return ""
	}
	tblName := m.Model.GetTableName()
	ti, _, _ := m.GetSharding()
	return fmt.Sprintf("%s_%d", tblName, ti)
}

// GetDatabase 获取数据库名称（返回配置中的名称，不要使用实际数据库名称，因为实际数据库名称在不同环境可能不一样）
func (m *ShardingModelManager) GetDatabase() string {
	if m.Model == nil {
		return ""
	}
	_, di, _ := m.GetSharding()
	return fmt.Sprintf("%s_%d", m.Model.GetDatabase(), di)
}

// NewQuerier 创建一个查询对象
func (m *ShardingModelManager) NewQuerier() *Querier {
	conn, err := m.GetConnection()
	if err != nil {
		xlog.Errorf("get db [%s] connection failed: %s", m.GetDatabase(), err)
		conn = nil
	}
	return NewModelQuerier(m.Model).Connect(conn).SetOptions(m.Settings).Select(m.QueryFieldsString())
}

// NewRawQuerier 创建一个查询对象
func (m *ShardingModelManager) NewRawQuerier(querySQL string) *Querier {
	// 获取数据库连接
	conn, err := m.GetConnection()
	if err != nil {
		xlog.Errorf("get db [%s] connection failed: %s", m.GetDatabase(), err)
		conn = nil
	}
	return NewRawQuerier(querySQL).SetOptions(m.Settings).Connect(conn)
}

// NewCommander 创建一个Commander对象
func (m *ShardingModelManager) NewCommander() *Commander {
	conn, err := m.GetConnection()
	if err != nil {
		xlog.Errorf("get db [%s] connection failed: %s", m.GetDatabase(), err)
		conn = nil
	}
	return NewCommander(m.Settings).Connect(conn)
}

// BuildBatchInsertSql 构造批量插入语句
func (m *ShardingModelManager) BuildBatchInsertSql(data interface{}) (string, error) {
	if data == nil {
		return "", errors.New("can not insert nil data")
	}
	var objects []interface{} = make([]interface{}, 0)
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice, reflect.Array:
		valData := reflect.ValueOf(data)
		arrSize := valData.Len()
		if arrSize == 0 {
			return "", errors.New("empty params")
		}
		for i := 0; i < arrSize; i++ {
			objects = append(objects, valData.Index(i).Interface())
		}
	default:
		return "", errors.New("invalid params")
	}
	// 先获取字段列表
	insertFields := m.getInsertFields()
	insertSql := fmt.Sprintf("INSERT INTO %s(`%s`) VALUES", m.GetTableName(), strings.Join(insertFields, "`,`"))
	insertCount := 0
	for i, object := range objects {
		modelObj, ok := m.convert2Model(object)
		if !ok {
			continue
		}
		values := make([]string, 0)
		rv := reflect.ValueOf(modelObj)
		for _, field := range insertFields {
			propName := m.FieldMaps[field]
			val := m.GetSqlValue(field, rv.Elem().FieldByName(propName).Interface())
			values = append(values, val)
		}
		if i > 0 {
			insertSql += ","
		}
		insertSql += fmt.Sprintf("(%s)", strings.Join(values, ","))
		insertCount++
	}
	if insertCount <= 0 {
		return "", errors.New("no any qualified data to insert")
	}
	return insertSql, nil
}

// BuildInsertSql 构造单条插入语句
func (m *ShardingModelManager) BuildInsertSql(object interface{}) (string, error) {
	// 类型检查与转换
	modelObj, ok := m.convert2Model(object)
	if !ok {
		return "", fmt.Errorf("insert action expect a %T object, but %T found", m.Model, object)
	}
	// 先获取字段列表
	insertFields := m.getInsertFields()
	insertSql := fmt.Sprintf("INSERT INTO %s(`%s`) VALUES", m.GetTableName(), strings.Join(insertFields, "`,`"))
	values := make([]string, 0)
	rv := reflect.ValueOf(modelObj)
	for _, field := range insertFields {
		propName := m.FieldMaps[field]
		val := m.GetSqlValue(field, rv.Elem().FieldByName(propName).Interface())
		values = append(values, val)
	}
	insertSql += fmt.Sprintf("(%s)", strings.Join(values, ","))
	return insertSql, nil
}

// BuildReplaceIntoSql 构造REPLACE INTO语句
func (mm *ShardingModelManager) BuildReplaceIntoSql(data interface{}) (string, error) {
	if data == nil {
		return "", errors.New("can not replace into nil data")
	}
	var objects []interface{} = make([]interface{}, 0)
	ele := reflect.TypeOf(data)
	if ele.Kind() == reflect.Ptr {
		ele = ele.Elem()
	}
	switch ele.Kind() {
	case reflect.Slice, reflect.Array:
		valData := reflect.ValueOf(data)
		arrSize := valData.Len()
		if arrSize == 0 {
			return "", errors.New("empty params")
		}
		for i := 0; i < arrSize; i++ {
			objects = append(objects, valData.Index(i).Interface())
		}
	case reflect.Struct:
		objects = append(objects, data)
	default:
		return "", errors.New("invalid params")
	}
	// 先获取字段列表
	allFields := mm.Fields
	replaceSql := fmt.Sprintf("REPLACE INTO %s(`%s`) VALUES", quote(mm.GetTableName()), strings.Join(allFields, "`,`"))
	count := 0
	for i, object := range objects {
		modelObj, ok := mm.convert2Model(object)
		if !ok {
			continue
		}
		values := make([]string, 0)
		rv := reflect.ValueOf(modelObj)
		for _, field := range allFields {
			propName := mm.FieldMaps[field]
			val := mm.GetSqlValue(field, rv.Elem().FieldByName(propName).Interface())
			values = append(values, val)
		}
		if i > 0 {
			replaceSql += ","
		}
		replaceSql += fmt.Sprintf("(%s)", strings.Join(values, ","))
		count++
	}
	if count <= 0 {
		return "", errors.New("no any qualified data to replace into")
	}
	return replaceSql, nil
}

// BuildUpdateSql 构造更新语句
func (m *ShardingModelManager) BuildUpdateSql(object interface{}) (string, error) {
	// 类型检查与转换
	modelObj, ok := m.convert2Model(object)
	if !ok {
		return "", fmt.Errorf("insert action expect a %T object, but %T found", m.Model, object)
	}
	// 先获取字段列表
	updateFields := m.getInsertFields()
	updateSQL := fmt.Sprintf("UPDATE `%s` SET ", m.GetTableName())
	// 构造更新数据
	rv := reflect.ValueOf(modelObj)
	for i, field := range updateFields {
		propName := m.FieldMaps[field]
		val := m.GetSqlValue(field, rv.Elem().FieldByName(propName).Interface())
		if i > 0 {
			updateSQL += ", "
		}
		updateSQL += fmt.Sprintf(" `%s` = %s", field, val)
	}
	// 自增ID
	autoIncrementField := m.Model.AutoIncrementField()
	propName := m.FieldMaps[autoIncrementField]
	idVal := m.GetSqlValue(autoIncrementField, rv.Elem().FieldByName(propName).Interface())
	updateSQL += fmt.Sprintf(" WHERE `%s` = %s ", autoIncrementField, idVal)
	return updateSQL, nil
}

// BuildUpdateSqlByCond 构造更新语句
func (m *ShardingModelManager) BuildUpdateSqlByCond(params map[string]interface{}, cond interface{}) (string, error) {
	if len(params) <= 0 {
		return "", errors.New("nothing to update")
	}
	where, err := NewConditionBuilder().Build(cond, "AND")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(where) == "" {
		return "", errors.New("update condition can not be empty")
	}
	// 构造更新语句
	updateSQL := fmt.Sprintf("UPDATE `%s` SET ", m.GetTableName())
	counter := 0
	for field, iv := range params {
		// val := NewValue(iv).SQLValue()
		val := m.GetSqlValue(field, iv)
		if counter > 0 {
			updateSQL += ", "
		}
		updateSQL += fmt.Sprintf(" `%s` = %s", field, val)
		counter++
	}
	updateSQL += fmt.Sprintf(" WHERE %s ", where)
	return updateSQL, nil
}

// BuildDeleteSql 构造删除语句
func (m *ShardingModelManager) BuildDeleteSql(conds interface{}) (string, error) {
	delSQL := fmt.Sprintf("DELETE FROM `%s` WHERE ", m.GetTableName())
	where, err := BuildCondition(conds)
	if err != nil {
		return "", err
	}
	// 不支持无条件删除
	if where == "" {
		return "", fmt.Errorf("delete condition can not be empty")
	}
	delSQL += where
	return delSQL, nil
}

// Insert 插入一条新数据
func (m *ShardingModelManager) Insert(obj interface{}) (int64, error) {
	// 构造插入语句
	insertSQL, err := m.BuildInsertSql(obj)
	xlog.Debugf("* Insert : %s", insertSQL)
	if err != nil {
		return 0, err
	}
	// 获取数据库连接
	conn, err := m.GetConnection()
	if err != nil {
		return 0, err
	}
	// 执行插入操作
	result, err := conn.Exec(insertSQL)
	if err != nil {
		xlog.Error("exec insert failed : ", err, ";  sql : ", insertSQL)
		return 0, err
	}
	return result.LastInsertId()
}

// InsertBatch 批量插入数据
func (m *ShardingModelManager) InsertBatch(objs interface{}) (int64, error) {
	// 构造插入语句
	insertSQL, err := m.BuildBatchInsertSql(objs)
	xlog.Debugf("* Batch Insert : %s", insertSQL)
	if err != nil {
		return 0, err
	}
	// 获取数据库连接
	conn, err := m.GetConnection()
	if err != nil {
		return 0, err
	}
	// 执行插入操作
	_, err = conn.Exec(insertSQL)
	if err != nil {
		xlog.Error("exec batch insert failed : ", err, ";  sql : ", insertSQL)
		return 0, err
	}
	// 只返回是否成功
	return 1, nil
}

// ReplaceInto 批量插入/更新数据
func (mm *ShardingModelManager) ReplaceInto(objs interface{}) (int64, error) {
	replaceSQL, err := mm.BuildReplaceIntoSql(objs)
	xlog.Debugf("* SQL : %s", replaceSQL)
	if err != nil {
		return 0, err
	}
	// 获取数据库连接
	conn, err := mm.GetConnection()
	if err != nil {
		return 0, err
	}
	// 执行插入操作
	_, err = conn.Exec(replaceSQL)
	if err != nil {
		xlog.Error("exec failed : ", err, ";  sql : ", replaceSQL)
		return 0, err
	}
	// 只返回是否成功
	return 1, nil
}

// Update 更新数据
func (m *ShardingModelManager) Update(obj interface{}) (int64, error) {
	// 构造更新语句
	updateSQL, err := m.BuildUpdateSql(obj)
	xlog.Debugf("* Update : %s", updateSQL)
	if err != nil {
		return 0, err
	}
	// 获取数据库连接
	conn, err := m.GetConnection()
	if err != nil {
		return 0, err
	}
	// 执行插入操作
	result, err := conn.Exec(updateSQL)
	if err != nil {
		xlog.Error("exec update failed : ", err, ";  sql : ", updateSQL)
		return 0, err
	}
	return result.RowsAffected()
}

// UpdateByCond 根据条件更新数据
func (m *ShardingModelManager) UpdateByCond(params map[string]interface{}, cond interface{}) (int64, error) {
	// 构造更新语句
	updateSQL, err := m.BuildUpdateSqlByCond(params, cond)
	xlog.Debugf("* UpdateByCond : %s", updateSQL)
	if err != nil {
		return 0, err
	}
	// 获取数据库连接
	conn, err := m.GetConnection()
	if err != nil {
		return 0, err
	}
	// 执行更新操作
	result, err := conn.Exec(updateSQL)
	if err != nil {
		xlog.Error("exec update failed : ", err, ";  sql : ", updateSQL)
		return 0, err
	}
	return result.RowsAffected()
}

// Delete 删除数据
func (m *ShardingModelManager) Delete(cond interface{}) (int64, error) {
	// 构造删除语句
	delSQL, err := m.BuildDeleteSql(cond)
	xlog.Debugf("* Delete : %s", delSQL)
	if err != nil {
		return 0, err
	}
	// 获取数据库连接
	conn, err := m.GetConnection()
	if err != nil {
		return 0, err
	}
	// 执行删除操作
	result, err := conn.Exec(delSQL)
	if err != nil {
		xlog.Error("exec delete failed : ", err, ";  sql : ", delSQL)
		return 0, err
	}
	return result.RowsAffected()
}

// MapToModeler 将map转换为Modeler对象(待测试)
func (m *ShardingModelManager) MapToModeler(data map[string]string) Modeler {
	if len(data) == 0 || m.Model == nil {
		return nil
	}
	// 创建对象并进行转换
	t := reflect.TypeOf(m.Model)
	// 指针类型获取真正type需要调用Elem
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// 调用反射创建对象
	newModel := reflect.New(t)
	// 遍历字段列表并设置值
	for field, val := range data {
		// 1. 检查model是否包含该字段
		propName, ok := m.FieldMaps[field]
		if !ok {
			continue
		}
		// 设置值
		reflectField := newModel.Elem().FieldByName(propName)
		propTypeKind := reflectField.Type().Kind()
		switch propTypeKind {
		case reflect.String:
			reflectField.SetString(NewValue(val).String())
		case reflect.Bool:
			reflectField.SetBool(NewValue(val).Boolean())
		case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8:
			reflectField.SetInt(NewValue(val).Int64())
		case reflect.Uint64, reflect.Uint, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			reflectField.SetUint(NewValue(val).Uint64())
		case reflect.Float64:
			reflectField.SetFloat(NewValue(val).Float64())
		default:   // 其他类型暂不支持
			break
		}
	}
	// 返回结果
	// return newModel.Interface().(Modeler)
	// 读取后的数据处理
	mod := newModel.Interface().(Modeler)
	if m.postReadFunc != nil {
		mod = m.postReadFunc(mod, data)
	}
	// 返回结果
	return mod
}

// FindPage 分页查询
func (m *ShardingModelManager) FindPage(conds interface{}, orderBy string, page, pageSize int) (*QueryResult, error) {
	return m.NewQuerier().From(m.GetTableName()).Where(conds).OrderBy(orderBy).QueryPage(page, pageSize)
}

// FindOne 查询单条数据
func (m *ShardingModelManager) FindOne(conds interface{}, orderBy string) (Modeler, error) {
	data, err := m.NewQuerier().From(m.GetTableName()).Where(conds).OrderBy(orderBy).QueryRow()
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	mData := m.MapToModeler(data)
	return mData, nil
}

// FindAll 查询满足条件的全部数据
func (m *ShardingModelManager) FindAll(conds interface{}, orderBy string) ([]interface{}, error) {
	queryRs, err := m.NewQuerier().From(m.GetTableName()).Where(conds).OrderBy(orderBy).Query()
	if err != nil {
		return nil, err
	}
	if queryRs.RowsCount == 0 {
		return nil, nil
	}
	list := make([]interface{}, 0)
	for _, d := range queryRs.Rows {
		v := m.MapToModeler(d)
		list = append(list, v)
	}
	return list, nil
}

// FindOne 查询单条数据
func (m *ShardingModelManager) Count(conds interface{}) (int, error) {
	data, err := m.NewQuerier().Select("COUNT(0)").From(m.GetTableName()).Where(conds).QueryScalar()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(data)
}

// QueryRaw 根据SQL查询满足条件的全部数据
func (m *ShardingModelManager) QueryAll(querySql string) (*QueryResult, error) {
	queryRs, err := m.NewRawQuerier(querySql).Query()
	if err != nil {
		return nil, err
	}
	return queryRs, nil
}

// QueryRow 根据SQL查询满足条件的全部数据
func (m *ShardingModelManager) QueryRow(querySql string) (map[string]string, error) {
	row, err := m.NewRawQuerier(querySql).Limit(1).QueryRow()
	if err != nil {
		return nil, err
	}
	return row, nil
}


