// Package session ...
// 与数据库进行交互
package session

import (
	"database/sql"
	"sorm/clause"
	"sorm/dialect"
	"sorm/logger"
	"sorm/schema"
	"strings"
)

// Session session of the database
type Session struct {
	db      *sql.DB
	dialect dialect.Dialector
	// table schema
	refTable *schema.Schema
	clause   clause.Clause
	sql      strings.Builder
	// sql占位符对应的值
	sqlVars []interface{}
}

// New initiate db session
func New(db *sql.DB, dialect dialect.Dialector) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}

// DB return *sql.DB
func (s *Session) DB() *sql.DB {
	// if s.db == nil {

	// }
	return s.db
}

// Raw contact sql and vars
func (s *Session) Raw(sql string, vars ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, vars...)
	return s
}

// Clear clear sql and vars
func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
}

// Exec execute sql statement
func (s *Session) Exec() (result sql.Result, err error) {
	// 执行完方法清空sql和vars
	// 复用session，一次会话可执行多个sql
	defer s.Clear()
	logger.Info(s.sql.String(), s.sqlVars)
	result, err = s.DB().Exec(s.sql.String(), s.sqlVars...)
	if err != nil {
		logger.Error(err)
		return
	}
	return
}

// QueryRow ...
func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	logger.Info(s.sql.String(), s.sqlVars)
	res := s.DB().QueryRow(s.sql.String(), s.sqlVars...)
	return res
}

// Query ...
func (s *Session) Query() (rows *sql.Rows, err error) {
	defer s.Clear()
	logger.Info(s.sql.String(), s.sqlVars)
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		logger.Error(err)
		return nil, err
	}
	return
}
