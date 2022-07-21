package session

import (
	"database/sql"
	"sorm/logger"
	"strings"
)

// Session session of the database
type Session struct {
	db  *sql.DB
	sql strings.Builder
	// sql占位符对应的值
	sqlVars []interface{}
}

// New initiate db session
func New(db *sql.DB) *Session {
	return &Session{
		db: db,
	}
}

// DB return *sql.DB
func (s *Session) DB() *sql.DB {
	// if s.db == nil {

	// }
	return s.db
}

// Raw contact sql and vars
func (s *Session) Raw(sql string, vars ...interface{}) {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, vars...)
}

// Clear clear sql and vars
func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
}

// Exec execute sql statement
func (s *Session) Exec(sql string, vars ...interface{}) sql.Result {
	// 执行完方法清空sql和vars
	// 复用session，一次会话可执行多个sql
	defer s.Clear()
	logger.Info(sql, vars)
	res, err := s.DB().Exec(sql, vars...)
	if err != nil {
		logger.Error(err)
	}
	return res
}

// QueryRow ...
func (s *Session) QueryRow(sql string, vars ...interface{}) *sql.Row {
	defer s.Clear()
	logger.Info(sql, vars)
	res := s.DB().QueryRow(sql, vars...)
	return res
}

// Query ...
func (s *Session) Query(sql string, vars ...interface{}) (rows *sql.Rows, err error) {
	defer s.Clear()
	logger.Info(sql, vars)
	if rows, err = s.DB().Query(sql, vars...); err != nil {
		logger.Error(err)
		return nil, err
	}
	return
}
