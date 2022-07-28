package dialect

import (
	"fmt"
	"reflect"
	"time"
)

// Sqlite3 ...
type sqlite3 struct {
}

// ConvertTypeTo convert type to type of database
func (s *sqlite3) ConvertTypeTo(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uintptr, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return "int"
	case reflect.Float32, reflect.Float64:
		return "real"
	case reflect.Int64, reflect.Uint64:
		return "bigint"
	case reflect.String:
		return "text"
	case reflect.Array, reflect.Slice:
		return "blob"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

// TableExistsSQL 判断是否存在table的sql语句
func (s *sqlite3) TableExistsSQL(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	sql := "SELECT name FROM sqlite_master WHERE TYPE= 'table' AND name= ?;"

	return sql, args
}
