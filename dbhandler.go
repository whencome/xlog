package gomodel

// Options 选项设置，用于扩展设置相关参数
type Options struct {
	PreQuerySyntaxCheck	bool  // 查询前置语法检查
}

// NewDefaultOptions 创建一个默认的Options
func NewDefaultOptions() *Options {
	return &Options{
		PreQuerySyntaxCheck : true,
	}
}

// PreQuerySyntaxCheckEnabled 判断查询前置语法检测是否开启
func (o *Options) IsPreQuerySyntaxCheckEnabled() bool {
	return o.PreQuerySyntaxCheck
}



