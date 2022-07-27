// Package session ...
// 数据库表相关操作
package session

import (
	"fmt"
	"reflect"
	"sorm/logger"
	"sorm/schema"
	"strings"
)

// Model init refTable of session
func (s *Session) Model(value interface{}) *Session {
	// refTable==nil 或传入的类型发生变化才更新refTable
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = schema.Parse(value, s.dialect)
	}
	return s
}

// GetRefTable get refTable
func (s *Session) GetRefTable() *schema.Schema {
	if s.refTable == nil {
		logger.Error("Invalid refTable")
	}
	return s.refTable
}

// CreateTable create a new table
func (s *Session) CreateTable() error {
	var columns []string
	table := s.refTable
	for _, col := range table.Fields {
		// 拼接字段名、类型和标签
		columns = append(columns, fmt.Sprintf("%s %s %s", col.Name, col.Type, col.Tag))
	}

	desc := strings.Join(columns, ",")
	_, err := s.db.Exec(`CREATE TABLE %s (%s) `, table.Name, desc)
	return err
}

// DropTable drop a table
func (s *Session) DropTable() error {
	_, err := s.db.Exec(`DROP TABLE %s`, s.refTable.Name)
	return err
}

// HasTable check if table exists
func (s *Session) HasTable(tableName string) bool {
	sql, args := s.dialect.TableExistsSQL(tableName)
	row := s.db.QueryRow(sql, args...)
	var tmp string
	_ = row.Scan(&tmp)
	return tmp == s.GetRefTable().Name
}
