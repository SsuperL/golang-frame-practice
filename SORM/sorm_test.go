package sorm

import (
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Name string `sorm:"PRIMARY KEY"`
	Age  int
}

func TestMigrate(t *testing.T) {
	engine, _ := NewEngine("sqlite3", "sorm.db")
	defer engine.Close()

	session := engine.NewSession()
	session.Raw("DROP TABLE IF EXISTS User;").Exec()
	session.Raw("CREATE TABLE User(Name text PRIMARY KEY, XXX integer);").Exec()
	if err := engine.Migrate(&User{}); err != nil {
		t.Fatalf("Failed to migrate, err:%v", err)
	}

	rows, _ := session.Raw("SELECT * FROM User;").Query()
	columns, _ := rows.Columns()
	if !reflect.DeepEqual(columns, []string{"Name", "Age"}) {
		t.Fatalf("Rows not equal, columns: %v", columns)
	}

}
