package sorm

import (
	"database/sql"
	"fmt"
	"sorm/dialect"
	"sorm/logger"
	"sorm/session"
	"strings"

	log "sorm/logger"
)

// Engine entry of orm framework
type Engine struct {
	db      *sql.DB
	dialect dialect.Dialector
}

// NewEngine create an engine
func NewEngine(driver, source string) (engine *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}

	if err = db.Ping(); err != nil {
		log.Error("cannot connect to database")
		return
	}

	dialect, ok := dialect.GetDialect(driver)
	if !ok {
		logger.Errorf("dialect %s not found", driver)
		return
	}
	engine = &Engine{db: db, dialect: dialect}
	log.Info("Connecting to database successfully")

	return engine, nil
}

// NewSession create a session
func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db, engine.dialect)
}

// Close close session
func (engine *Engine) Close() error {
	if err := engine.db.Close(); err != nil {
		log.Error("close session failed :%v", err)
		return err
	}
	log.Info("session closed")

	return nil
}

// TxFunc ...
type TxFunc func(s *session.Session) (interface{}, error)

// Transaction ...
func (engine *Engine) Transaction(fs TxFunc) (result interface{}, err error) {
	session := engine.NewSession()
	tx, err := session.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	return fs(session)
}

// return difference a - b
func difference(a, b []string) []string {
	m := make(map[string]bool)
	for _, v := range b {
		m[v] = true
	}

	var diff []string
	for _, v := range a {
		if _, ok := m[v]; !ok {
			diff = append(diff, v)
		}
	}

	return diff
}

// Migrate only add columns and delete columns
// delete columns 采用先创建临时表，临时表替换原表的方式实现
// CREATE TABLE tmp AS SELECT field1, field2 ... FROM TABLE_SOURCE
// DROP TABLE TABLE_SOURCE;
// ALTER TABLE tmp RENAME TO TABLE_SOURCE;
func (engine *Engine) Migrate(value interface{}) (err error) {
	engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		if !s.Model(value).HasTable() {
			logger.Infof("table %v not exists", s.Model(&value).GetRefTable().Name)
			s.CreateTable()
		}

		table := s.GetRefTable()
		rows, _ := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1;", table.Name)).Query()
		columns, _ := rows.Columns()
		// add columns
		addCols := difference(table.FieldNames, columns)
		// delete columns
		deleteCols := difference(columns, table.FieldNames)
		logger.Infof("add cols %v ; delete cols %v", addCols, deleteCols)

		for _, col := range addCols {
			field := table.GetField(col)
			if _, err = s.Raw(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table.Name, field.Name, field.Type)).Exec(); err != nil {
				return
			}
		}

		if len(deleteCols) == 0 {
			return
		}

		tmp := "tmp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s FROM %s ;", tmp, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s;", table.Name))
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", tmp, table.Name))
		_, err = s.Exec()
		return
	})
	return err
}
