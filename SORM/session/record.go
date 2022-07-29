package session

import (
	"errors"
	"reflect"
	"sorm/clause"
)

// Insert orm insert
// session.Insert(&User{Name:"Tom",Age:12})
func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		// value为要映射的表实例
		table := s.Model(value).GetRefTable()
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		recordValues = append(recordValues, table.RecordValues(value))
	}
	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Find find of orm
// var users []Users
// session.Find(&users)
// 根据获取到的值构造出相应对象
func (s *Session) Find(values interface{}) error {
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	table := s.Model(reflect.New(destType).Elem().Interface()).GetRefTable()

	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	rows, err := s.Raw(sql, vars...).Query()
	if err != nil {
		return err
	}
	for rows.Next() {
		dest := reflect.New(destType).Elem()
		var values []interface{}
		for _, fieldName := range table.FieldNames {
			values = append(values, dest.FieldByName(fieldName).Addr().Interface())
		}
		if err := rows.Scan(values...); err != nil {
			return err
		}
		destSlice.Set(reflect.Append(destSlice, dest))
	}

	return rows.Close()
}

// Update update of orm
// 1. map[string]interface{}
// 2. k-v list : "Name","Tome","Age",12
func (s *Session) Update(kvs ...interface{}) (int64, error) {
	m, ok := kvs[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		for i := 0; i < len(kvs); i += 2 {
			m[kvs[i].(string)] = kvs[i+1]
		}
	}

	s.clause.Set(clause.UPDATE, s.refTable.Name, m)
	sql, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Delete delete of orm
func (s *Session) Delete(values ...interface{}) (int64, error) {
	// DELETE FROM $tableName
	table := values[0]
	s.clause.Set(clause.DELETE, table)
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Count count of orm
func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.refTable.Name)
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, vars...).QueryRow()
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Limit chain链式调用
func (s *Session) Limit(limit int) *Session {
	s.clause.Set(clause.LIMIT, limit)
	return s
}

// Where chain链式调用
func (s *Session) Where(desc string, values ...interface{}) *Session {
	var vars []interface{}
	s.clause.Set(clause.WHERE, append(append(vars, desc), values...)...)
	return s
}

// OrderBy chain链式调用
func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}

// First return the first record of result
// session.Where(...).First()
func (s *Session) First(value interface{}) error {
	dest := reflect.Indirect(reflect.ValueOf(value))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("Record Not Found")
	}

	dest.Set(destSlice.Index(0))

	return nil
}
