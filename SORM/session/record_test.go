package session

import (
	"fmt"
	"testing"

	"github.com/go-playground/assert"
)

func initTest(t *testing.T) *Session {
	t.Helper()
	session := NewSession()
	session = session.Model(&User{})
	if session.HasTable("User") {
		session.DropTable()
	}
	err2 := session.CreateTable()
	if err2 != nil {
		err := fmt.Errorf("err2:%v", err2)
		t.Fatalf("failed to init : %v", err)
	}
	session.Raw("INSERT INTO User VALUES (?,?),(?,?)", "Tom", 12, "John", 14).Exec()
	return session
}
func TestSession_Insert(t *testing.T) {
	session := initTest(t)
	rowsAffected, err := session.Insert(&User{"April", 15})
	if err != nil {
		t.Fatal("Insert failed")
	}
	assert.Equal(t, 1, int(rowsAffected))
}

func TestSession_Find(t *testing.T) {
	session := initTest(t)
	var users []User
	err := session.Find(&users)
	if err != nil {
		t.Fatal("Find failed")
	}
	assert.Equal(t, 2, len(users))
}
