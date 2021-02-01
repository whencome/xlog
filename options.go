package gomodel

// Options 选项设置，用于扩展设置相关参数
type Options struct {
	PreQuerySyntaxCheck	bool  // 查询前置语法检查
	EnableSharding 		bool  // 是否支持sharding
	DbShardingNum		int64   // 数据库分库数量
	TableShardingNum	int64   // 每个数据库分表数量
}

// NewDefaultOptions 创建一个默认的Options
func NewDefaultOptions() *Options {
	return &Options{
		PreQuerySyntaxCheck : true,
		EnableSharding : false,
		DbShardingNum : 1,
		TableShardingNum : 1,
	}
}

// NewShardingOptions 创建一个分库分表的Options
func NewShardingOptions(tableNum, dbNum int64) *Options {
	return &Options{
		PreQuerySyntaxCheck : true,
		EnableSharding : true,
		DbShardingNum : dbNum,
		TableShardingNum : tableNum,
	}
}

// PreQuerySyntaxCheckEnabled 判断查询前置语法检测是否开启
func (o *Options) IsPreQuerySyntaxCheckEnabled() bool {
	return o.PreQuerySyntaxCheck
}



