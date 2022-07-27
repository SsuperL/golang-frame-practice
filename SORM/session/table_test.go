package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	Name string `sorm:"index"`
	Age  int
}

func TestSession_CreateTable(t *testing.T) {
	session := NewSession()
	session = session.Model(&User{})
	session.CreateTable()
	exists := session.HasTable("User")
	assert.True(t, true, exists)
}

func TestSession_DropTable(t *testing.T) {
	session := NewSession()
	session = session.Model(&User{})
	session.DropTable()
	exists := session.HasTable("User")
	assert.Equal(t, false, exists)
}
