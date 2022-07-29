package session

import (
	"database/sql"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	TestDB, _ = sql.Open("sqlite3", "./sorm.db")
	TestDB.Exec("DROP TABLE IF EXISTS User;")
	TestDB.Exec("DROP TABLE IF EXISTS Person;")
	code := m.Run()
	TestDB.Exec("DROP TABLE IF EXISTS User;")
	_ = TestDB.Close()
	if _, err := os.Stat("sorm.db"); !os.IsNotExist(err) {
		// 清理，移除测试用db文件
		os.Remove("sorm.db")
	}
	os.Exit(code)
}
