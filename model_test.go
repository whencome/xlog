package gomodel

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

/**
CREATE TABLE `user` (
    `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `name` varchar(50) NOT NULL DEFAULT '' COMMENT '用户姓名',
    `email` varchar(100) NOT NULL DEFAULT '' COMMENT '用户邮箱',
    `mobile` varchar(20) NOT NULL DEFAULT '' COMMENT '用户手机号',
    `track` linestring NULL COMMENT '用户轨迹',
    `create_time` int(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '创建时间',
    `update_time` int(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '更新时间',

    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
 */

type Point struct {
	X 		float64
	Y 		float64
}

type User struct{
	ID 					int64	`db:"id" json:"id"`
	Name 				string 	`db:"name" json:"name"`
	Email 				string 	`db:"email" json:"email"`
	Mobile 				string 	`db:"mobile" json:"mobile"`
	Track				[]Point `db:"track" json:"track"`
	CreateTime			int64 	`db:"create_time" json:"create_time"`
	UpdateTime 			int64 	`db:"update_time" json:"update_time"`
}

// GetDatabase 获取数据库名称（返回配置中的名称，不要使用实际数据库名称，因为实际数据库名称在不同环境可能不一样）
func (u *User) GetDatabase() string {
	return "test"
}

// GetTableName 获取数据库数据存放的数据表名称
func (u *User) GetTableName() string {
	return "user"
}

// AutoIncrementField 自增字段名称，如果没有则返回空
func (u *User) AutoIncrementField() string {
	return "id"
}

// GetDBFieldTag 获取数据库字段映射tag
func (u *User) GetDBFieldTag() string {
	return "db"
}

// UserModel model for User
type UserModel struct {
	*ModelManager
}

// NewUserModel create a User Model
func NewUserModel() *UserModel {
	m := &UserModel{
		NewModelManager(&User{}),
	}
	// 设置数据库初始化的方法
	m.SetDBInitFunc(func() (db *sql.DB, e error) {
		// 连接数据库
		conn, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/ddl_test?charset=utf8")
		if err != nil {
			return nil, err
		}
		// 返回连接信息
		return conn, nil
	})
	// 针对具体字段进行处理
	m.SetSqlValueCallback("email", func(v interface{}) string {
		_v := v.(string)
		return "'" + strings.ReplaceAll(_v, "@", "#") + "'"
	})
	m.SetSqlValueCallback("track", func(v interface{}) string {
		points := v.([]Point)
		lintPoints := &bytes.Buffer{}
		lintPoints.WriteString("LINESTRING(")
		for i, p := range points {
			if i > 0 {
				lintPoints.WriteString(",")
			}
			lintPoints.WriteString(fmt.Sprintf("%.6f %.6f", p.X, p.Y))
		}
		lintPoints.WriteString(")")
		return fmt.Sprintf("ST_GeomFromText('%s')", lintPoints)
	})
	// 针对modeler写入前进行预处理
	m.SetPreWriteFunc(func(mod Modeler) Modeler {
		u := mod.(*User)
		if u.ID == 0 {
			u.CreateTime = time.Now().Unix()
		}
		if u.UpdateTime <= 0 {
			u.UpdateTime = time.Now().Unix()
		}
		return u
	})
	// 设置查询前的字段处理
	m.SetPreQueryFieldFunc(func(f string) string {
		if f != "track" {
			return f
		}
		return fmt.Sprintf("ST_astext(%s) as %s", f, f)
	})
	// 读取数据后进行处理
	m.SetPostReadFunc(func(mod Modeler, data map[string]string) Modeler {
		u := mod.(*User)
		// 对手机号进行打马
		u.Mobile = u.Mobile[:3] + "****" + u.Mobile[8:]
		// 对email进行还原
		u.Email = strings.ReplaceAll(u.Email, "#", "@")
		// 解析轨迹
		line, ok := data["track"]
		if ok {
			line = strings.ReplaceAll(line, "LINESTRING(", "")
			line = strings.ReplaceAll(line, ")", "")
			parts := strings.Split(line, ",")
			points := make([]Point, 0)
			for _, part := range parts {
				v := strings.Split(part, " ")
				if len(v) != 2 {
					continue
				}
				p := Point{}
				p.X = NewValue(v[0]).Float64()
				p.Y = NewValue(v[1]).Float64()
				points = append(points, p)
			}
			u.Track = points
		}
		return u
	})
	return m
}

// 创建一个测试用户
var u *User = &User{
	ID:1001,
	Name:"Jack Smith",
	Email:"jack.smith@unknownsite.com",
	Mobile:"12345678900",
	Track:[]Point{
		{X:121.474, Y:31.2345},
		{X:121.472, Y:31.2333},
		{X:121.471, Y:31.2315},
	},
}

// 测试构造插入语句
func TestModelManager_BuildInsertSql(t *testing.T) {
	m := NewUserModel()
	insertSql, e := m.BuildInsertSql(u)
	if e != nil {
		t.Logf("build insert sql failed: %s", e)
		t.Fail()
	}
	t.Log(insertSql)
}

// 测试构造更新语句
func TestModelManager_BuildUpdateSql(t *testing.T) {
	m := NewUserModel()
	updateSql, e := m.BuildUpdateSql(u)
	if e != nil {
		t.Logf("build insert sql failed: %s", e)
		t.Fail()
	}
	t.Log(updateSql)
}

// 测试插入数据
func TestModelManager_Insert(t *testing.T) {
	m := NewUserModel()
	id, e := m.Insert(u)
	if e != nil {
		t.Logf("insert user failed: %s", e)
		t.Fail()
	}
	t.Logf("insert id = %d", id)
}

// 测试读取数据
func TestModelManager_FindOne(t *testing.T) {
	m := NewUserModel()
	conds := m.NewAndCondition()
	conds.Add("name", "Jack Smith")
	u, e := m.FindOne(conds, "id DESC")
	if e != nil {
		t.Logf("query user failed: %s", e)
		t.Fail()
	}
	t.Logf("user = %#v", u)
}