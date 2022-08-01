package sorm

import (
	"database/sql"
	"sorm/dialect"
	"sorm/logger"
	"sorm/session"

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
