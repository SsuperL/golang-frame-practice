package session

import (
	"sorm/logger"
	"testing"

	"github.com/go-playground/assert"
)

type Person struct {
	Name   string `sorm:"PRIMARY KEY"`
	Gender string
}

func (p *Person) BeforeInsert(s *Session) error {
	logger.Info("BeforeInsert ...")
	return nil
}

func (p *Person) AfterQuery(s *Session) error {
	p.Gender = "Female"
	return nil
}

var _ IBeforeInsert = (*Person)(nil)
var _ IAfterQuery = (*Person)(nil)

func TestHook(t *testing.T) {
	session := NewSession()
	session = session.Model(&Person{})
	session.CreateTable()
	rowAffected, err := session.Insert(&Person{"Jay", "Male"})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	assert.Equal(t, 1, int(rowAffected))
	var persons []Person
	if err := session.Find(&persons); err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	assert.Equal(t, "Jay", persons[0].Name)
	assert.Equal(t, "Female", persons[0].Gender)
	session.DropTable()
}
