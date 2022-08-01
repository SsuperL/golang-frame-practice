package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type User struct {
	Name string `sorm:"PRIMARY KEY"`
	Age  int
}

func TestSession_CreateTable(t *testing.T) {
	session := NewSession()
	session = session.Model(&User{})
	if session.HasTable() {
		session.DropTable()
	}
	session.CreateTable()
	exists := session.HasTable()
	require.True(t, true, exists)
}

func TestSession_HasTable(t *testing.T) {
	session := NewSession()
	session = session.Model(&User{})
	session.Raw("DROP TABLE IF EXISTS User;").Exec()
	session.CreateTable()
	exists := session.HasTable()
	assert.True(t, exists)
}
