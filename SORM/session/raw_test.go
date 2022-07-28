package session

import (
	"database/sql"
	"sorm/dialect"
	"testing"

	"github.com/go-playground/assert"
	_ "github.com/mattn/go-sqlite3"
)

var (
	TestDB              *sql.DB
	sqlite3Dialector, _ = dialect.GetDialect("sqlite3")
)

func NewSession() *Session {
	return New(TestDB, sqlite3Dialector)
}

func TestQueryRow(t *testing.T) {
	session := NewSession()
	session.Raw("DROP TABLE IF EXISTS User;").Exec()
	session.Raw("CREATE TABLE User (name text);").Exec()
	session.Raw("INSERT INTO User VALUES(?)", "test").Exec()
	row := session.Raw("SELECT * FROM User WHERE name = ?", "test").QueryRow()
	var name string
	row.Scan(&name)
	assert.Equal(t, "test", name)
}

func TestQuery(t *testing.T) {
	session := NewSession()
	session.Raw("DROP TABLE IF EXISTS User;").Exec()
	session.Raw("CREATE TABLE User (name text);").Exec()
	session.Raw("INSERT INTO User VALUES(?),(?)", "A", "B").Exec()
	rows, err := session.Raw("SELECT COUNT(*) FROM User").Query()
	var count int
	for rows.Next() {
		rows.Scan(&count)
	}
	if err != nil || count == 0 {
		t.Fatal("query failed")
	}
	assert.Equal(t, 2, count)

}
