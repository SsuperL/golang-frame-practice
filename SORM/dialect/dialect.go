package dialect

import "reflect"

// Dialector 支持不同数据库的差异
type Dialector interface {
	//ConvertTypeTo 用于将Go的类型转换为对应数据库的数据类型
	ConvertTypeTo(typ reflect.Value) string
	// TableExistsSQL 判断表是否存在的SQL语句
	TableExistsSQL(tableName string) (string, []interface{})
}

// key 是数据库类型
var dialectsMap = map[string]Dialector{}

func init() {
	RegisterDialect("sqlite3", &sqlite3{})
}

// RegisterDialect 注册dialect
func RegisterDialect(name string, dialect Dialector) {
	if _, ok := dialectsMap[name]; !ok {
		dialectsMap[name] = dialect
	}
}

// GetDialect get dialector
func GetDialect(name string) (Dialector, bool) {
	dialect, ok := dialectsMap[name]
	return dialect, ok
}
