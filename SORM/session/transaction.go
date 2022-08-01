package session

import (
	"database/sql"
	"sorm/logger"
)

// Begin start a transaction
func (s *Session) Begin() (tx *sql.Tx, err error) {
	if s.tx, err = s.db.Begin(); err != nil {
		logger.Error("Start transaction failed: %v", err)
		return nil, err
	}

	return s.tx, nil
}

// Commit commit a transaction
func (s *Session) Commit() error {
	if err := s.tx.Commit(); err != nil {
		logger.Error("Commit failed: %v", err)
		return err
	}

	return nil
}

// Rollback rollback a transaction
func (s *Session) Rollback() error {
	if err := s.tx.Rollback(); err != nil {
		logger.Error("Rollback failed: %v", err)
		return err
	}

	return nil
}
