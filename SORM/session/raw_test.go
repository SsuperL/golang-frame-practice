package session

import (
	"database/sql"
	"os"
	"sorm/dialect"
	"testing"

	"github.com/go-playground/assert"
	_ "github.com/mattn/go-sqlite3"
)

var (
	TestDB              *sql.DB
	sqlite3Dialector, _ = dialect.GetDialect("sqlite3")
)

func TestMain(m *testing.M) {
	TestDB, _ = sql.Open("sqlite3", "./sorm.db")
	code := m.Run()
	_ = TestDB.Close()
	if _, err := os.Stat("sorm.db"); !os.IsNotExist(err) {
		// 清理，移除测试用db文件
		os.Remove("sorm.db")
	}
	os.Exit(code)
}

func NewSession() *Session {
	return New(TestDB, sqlite3Dialector)
}

func TestQueryRow(t *testing.T) {
	session := NewSession()
	session.Exec("DROP TABLE IF EXISTS User;")
	session.Exec("CREATE TABLE User (name text);")
	session.Exec("INSERT INTO User VALUES(?)", "test")
	row := session.QueryRow("SELECT * FROM User WHERE name = ?", "test")
	var name string
	row.Scan(&name)
	assert.Equal(t, "test", name)
}

func TestQuery(t *testing.T) {
	session := NewSession()
	session.Exec("DROP TABLE IF EXISTS User;")
	session.Exec("CREATE TABLE User (name text);")
	session.Exec("INSERT INTO User VALUES(?),(?)", "A", "B")
	rows, err := session.Query("SELECT COUNT(*) FROM User")
	var count int
	if rows.Next() {
		rows.Scan(&count)
	}
	if err != nil || count == 0 {
		t.Fatal("query failed")
	}
	assert.Equal(t, 2, count)

}
