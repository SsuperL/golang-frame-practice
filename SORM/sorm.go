package sorm

import (
	"database/sql"
	"sorm/session"

	log "sorm/logger"
)

// Engine entry of orm framework
type Engine struct {
	db *sql.DB
}

// NewEngine create an engine
func NewEngine(driver, source string) (*Engine, error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		log.Error("cannot connect to database")
		return nil, err
	}

	engine := &Engine{db: db}
	log.Info("Connecting to database successfully")

	return engine, nil
}

// NewSession create a session
func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db)
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
