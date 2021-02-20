package gomodel

import (
    "database/sql"
    "errors"
    "fmt"
    "reflect"
    "strconv"
    "strings"
    
    "github.com/whencome/xlog"
)


/************************************************************
 ******                SECTION OF MODELER               *****
 ************************************************************/

// Modeler 定义一个model接口
type Modeler interface {
    GetDatabase() string
    GetTableName() 	string
    AutoIncrementField() string
    GetDBFieldTag() string
}

/************************************************************
 ******            SECTION OF MODEL MANAGER             *****
 ************************************************************/

type Manager interface {
    SetDBInitFunc(func()(*sql.DB, error))
    GetConnection() (*sql.DB, error)
}

// 定义ModelManager结构体，用于数据model操作管理
type ModelManager struct {
    Manager
    Model 			Modeler
    Fields          []string
    FieldMaps       map[string]string
    Settings        *Options
    GetDBFunc       func()(*sql.DB, error)
}

// NewModelManager 创建一个新的ModelManager
func NewModelManager(m Modeler) *ModelManager {
    fieldMaps := map[string]string{}
    fields := make([]string, 0)
    // 获取tag中的内容
    rt := reflect.TypeOf(m)
    // 获取字段数量
    fieldsNum := rt.Elem().NumField()
    for i := 0; i < fieldsNum; i++ {
        field := rt.Elem().Field(i)
        fieldName := field.Name
        tableFieldName := field.Tag.Get(m.GetDBFieldTag())
        if tableFieldName == "" {
            continue
        }
        fields = append(fields, tableFieldName)
        fieldMaps[tableFieldName] = fieldName
    }
    return &ModelManager{
        Model:     m,
        Fields:    fields,
        FieldMaps: fieldMaps,
        Settings: NewDefaultOptions(),
    }
}

// NewCustomModelManager 创建一个定制化的ModelManager
func NewCustomModelManager(m Modeler, opts *Options) *ModelManager {
    mm := NewModelManager(m)
    mm.SetOptions(opts)
    return mm
}

// SetOptions 设置选项
func (mm *ModelManager) SetOptions(opts *Options) {
    if opts == nil {
        return
    }
    mm.Settings = opts
}


// GetTableName 获取Model对应的数据表名
func (mm *ModelManager) GetTableName() string {
    if mm.Model == nil {
        return ""
    }
    return mm.Model.GetTableName()
}

// GetDatabase 获取数据库名称（返回配置中的名称，不要使用实际数据库名称，因为实际数据库名称在不同环境可能不一样）
func (mm *ModelManager) GetDatabase() string {
    if mm.Model == nil {
        return ""
    }
    return mm.Model.GetDatabase()
}

// SetDBInitFunc 设置数据库初始化函数
func (mm *ModelManager) SetDBInitFunc(f func()(*sql.DB, error)) {
    mm.GetDBFunc = f
}

// GetConnection 获取数据库连接
func (mm *ModelManager) GetConnection() (*sql.DB, error) {
    return mm.GetDBFunc()
}

// NewAndCondition 创建一个AND条件组
func (mm *ModelManager) NewAndCondition() *Condition {
    return NewAndCondition()
}

// NewAndCondition 创建一个OR条件组
func (mm *ModelManager) NewOrCondition() *Condition {
    return NewOrCondition()
}

// NewQuerier 创建一个查询对象
func (mm *ModelManager) NewQuerier() *Querier {
    conn, _ := mm.GetConnection()
    return NewModelQuerier(mm.Model).Connect(conn).SetOptions(mm.Settings)
}

// NewRawQuerier 创建一个查询对象
func (mm *ModelManager) NewRawQuerier(querySQL string) *Querier {
    // 获取数据库连接
    conn, _ := mm.GetConnection()
    return NewRawQuerier(querySQL).SetOptions(mm.Settings).Connect(conn)
}

// NewCommander 创建一个Commander对象
func (mm *ModelManager) NewCommander() *Commander {
    conn, _ := mm.GetConnection()
    return NewCommander(mm.Settings).Connect(conn)
}

// getInsertFields 获取插入的字段列表
func (mm *ModelManager) getInsertFields() []string {
    fields := make([]string, 0)
    for _, field := range mm.Fields {
        if field == mm.Model.AutoIncrementField() {
            continue
        }
        fields = append(fields, field)
    }
    return fields
}

// MatchObject 匹配对象，检查对象类型是否匹配
func (mm *ModelManager) MatchObject(obj interface{}) bool {
    if obj == nil {
        return false
    }
    modelObj, ok := obj.(Modeler)
    if !ok {
        return false
    }
    if mm.Model == nil || modelObj.GetTableName() != mm.Model.GetTableName() {
        return false
    }
    return true
}

// BuildBatchInsertSql 构造批量插入语句
func (mm *ModelManager) BuildBatchInsertSql(data interface{}) (string, error) {
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
    insertFields := mm.getInsertFields()
    insertSql := fmt.Sprintf("INSERT INTO %s(`%s`) VALUES", mm.GetTableName(), strings.Join(insertFields, "`,`"))
    insertCount := 0
    for i, object := range objects {
        if !mm.MatchObject(object) {
            continue
        }
        modelObj, _ := object.(Modeler)
        values := make([]string, 0)
        rv := reflect.ValueOf(modelObj)
        for _, field := range insertFields {
            propName := mm.FieldMaps[field]
            val := NewValue(rv.Elem().FieldByName(propName).Interface()).SQLValue()
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
func (mm *ModelManager) BuildInsertSql(object interface{}) (string, error) {
    // 类型检查
    if !mm.MatchObject(object) {
        return "", fmt.Errorf("insert action expect a %T object, but %T found", mm.Model, object)
    }
    // 先获取字段列表
    insertFields := mm.getInsertFields()
    insertSql := fmt.Sprintf("INSERT INTO %s(`%s`) VALUES", mm.GetTableName(), strings.Join(insertFields, "`,`"))
    modelObj, _ := object.(Modeler)
    values := make([]string, 0)
    rv := reflect.ValueOf(modelObj)
    for _, field := range insertFields {
        propName := mm.FieldMaps[field]
        val := NewValue(rv.Elem().FieldByName(propName).Interface()).SQLValue()
        values = append(values, val)
    }
    insertSql += fmt.Sprintf("(%s)", strings.Join(values, ","))
    return insertSql, nil
}

// BuildUpdateSql 构造更新语句
func (mm *ModelManager) BuildUpdateSql(object interface{}) (string, error) {
    // 类型检查
    if !mm.MatchObject(object) {
        return "", fmt.Errorf("update action expect a %T object, but %T found", mm.Model, object)
    }
    // 先获取字段列表
    updateFields := mm.getInsertFields()
    updateSQL := fmt.Sprintf("UPDATE `%s` SET ", mm.GetTableName())
    modelObj, _ := object.(Modeler)
    rv := reflect.ValueOf(modelObj)
    for i, field := range updateFields {
        propName := mm.FieldMaps[field]
        val := NewValue(rv.Elem().FieldByName(propName).Interface()).SQLValue()
        if i > 0 {
            updateSQL += ", "
        }
        updateSQL += fmt.Sprintf(" `%s` = %s", field, val)
    }
    // 自增ID
    autoIncrementField := mm.Model.AutoIncrementField()
    propName := mm.FieldMaps[autoIncrementField]
    idVal := NewValue(rv.Elem().FieldByName(propName).Interface()).SQLValue()
    updateSQL += fmt.Sprintf(" WHERE `%s` = %s ", autoIncrementField, idVal)
    return updateSQL, nil
}

// BuildUpdateSqlByCond 构造更新语句
func (mm *ModelManager) BuildUpdateSqlByCond(params map[string]interface{}, cond interface{}) (string, error) {
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
    updateSQL := fmt.Sprintf("UPDATE `%s` SET ", mm.GetTableName())
    counter := 0
    for field, iv := range params {
        val := NewValue(iv).SQLValue()
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
func (mm *ModelManager) BuildDeleteSql(conds interface{}) (string, error) {
    delSQL := fmt.Sprintf("DELETE FROM `%s` WHERE ", mm.GetTableName())
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
func (mm *ModelManager) Insert(obj interface{}) (int64, error) {
    // 构造插入语句
    insertSQL, err := mm.BuildInsertSql(obj)
    xlog.Debugf("* Insert : %s", insertSQL)
    if err != nil {
        return 0, err
    }
    // 获取数据库连接
    conn, err := mm.GetConnection()
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
func (mm *ModelManager) InsertBatch(objs interface{}) (int64, error) {
    // 构造插入语句
    insertSQL, err := mm.BuildBatchInsertSql(objs)
    xlog.Debugf("* Batch Insert : %s", insertSQL)
    if err != nil {
        return 0, err
    }
    // 获取数据库连接
    conn, err := mm.GetConnection()
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

// Update 更新数据
func (mm *ModelManager) Update(obj interface{}) (int64, error) {
    // 构造更新语句
    updateSQL, err := mm.BuildUpdateSql(obj)
    xlog.Debugf("* Update : %s", updateSQL)
    if err != nil {
        return 0, err
    }
    // 获取数据库连接
    conn, err := mm.GetConnection()
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
func (mm *ModelManager) UpdateByCond(params map[string]interface{}, cond interface{}) (int64, error) {
    // 构造更新语句
    updateSQL, err := mm.BuildUpdateSqlByCond(params, cond)
    xlog.Debugf("* UpdateByCond : %s", updateSQL)
    if err != nil {
        return 0, err
    }
    // 获取数据库连接
    conn, err := mm.GetConnection()
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
func (mm *ModelManager) Delete(cond interface{}) (int64, error) {
    // 构造删除语句
    delSQL, err := mm.BuildDeleteSql(cond)
    xlog.Debugf("* Delete : %s", delSQL)
    if err != nil {
        return 0, err
    }
    // 获取数据库连接
    conn, err := mm.GetConnection()
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
func (mm *ModelManager) MapToModeler(data map[string]string) Modeler {
    if len(data) == 0 || mm.Model == nil {
        return nil
    }
    // 创建对象并进行转换
    t := reflect.TypeOf(mm.Model)
    // 指针类型获取真正type需要调用Elem
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    // 调用反射创建对象
    newModel := reflect.New(t)
    // 遍历字段列表并设置值
    for field, val := range data {
        // 1. 检查model是否包含该字段
        propName, ok := mm.FieldMaps[field]
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
    return newModel.Interface().(Modeler)
}

// Map 将model转换为map
func (mm *ModelManager) Map(obj Modeler) map[string]interface{} {
    if !mm.MatchObject(obj) {
        return nil
    }
    retData := make(map[string]interface{})
    fields := mm.Fields
    rv := reflect.ValueOf(obj)
    for _, field := range fields {
        propName := mm.FieldMaps[field]
        val := rv.Elem().FieldByName(propName).Interface()
        retData[field] = val
    }
    // 返回结果
    return retData
}

// FindPage 分页查询
func (mm *ModelManager) FindPage(conds interface{}, orderBy string, page, pageSize int) (*QueryResult, error) {
    return mm.NewQuerier().From(mm.GetTableName()).Where(conds).OrderBy(orderBy).QueryPage(page, pageSize)
}

// FindOne 查询单条数据
func (mm *ModelManager) FindOne(conds interface{}, orderBy string) (Modeler, error) {
    data, err := mm.NewQuerier().From(mm.GetTableName()).Where(conds).OrderBy(orderBy).QueryRow()
    if err != nil {
        return nil, err
    }
    if data == nil {
        return nil, nil
    }
    mData := mm.MapToModeler(data)
    return mData, nil
}

// FindAll 查询满足条件的全部数据
func (mm *ModelManager) FindAll(conds interface{}, orderBy string) ([]interface{}, error) {
    queryRs, err := mm.NewQuerier().From(mm.GetTableName()).Where(conds).OrderBy(orderBy).Query()
    if err != nil {
        return nil, err
    }
    if queryRs.RowsCount == 0 {
        return nil, nil
    }
    list := make([]interface{}, 0)
    for _, d := range queryRs.Rows {
        v := mm.MapToModeler(d)
        list = append(list, v)
    }
    return list, nil
}

// FindOne 查询单条数据
func (mm *ModelManager) Count(conds interface{}) (int, error) {
    data, err := mm.NewQuerier().Select("COUNT(0)").From(mm.GetTableName()).Where(conds).QueryScalar()
    if err != nil {
        return 0, err
    }
    return strconv.Atoi(data)
}

// QueryRaw 根据SQL查询满足条件的全部数据
func (mm *ModelManager) QueryAll(querySql string) (*QueryResult, error) {
    queryRs, err := mm.NewRawQuerier(querySql).Query()
    if err != nil {
        return nil, err
    }
    return queryRs, nil
}

// QueryRow 根据SQL查询满足条件的全部数据
func (mm *ModelManager) QueryRow(querySql string) (map[string]string, error) {
    row, err := mm.NewRawQuerier(querySql).Limit(1).QueryRow()
    if err != nil {
        return nil, err
    }
    return row, nil
}



