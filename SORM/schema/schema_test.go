package schema

import (
	"sorm/dialect"
	"testing"

	"github.com/go-playground/assert"
)

type User struct {
	Name string `sorm:"name"`
	Age  int    `sorm:"age"`
}

var sqlite3Dialector, _ = dialect.GetDialect("sqlite3")

func TestParse(t *testing.T) {
	schema := Parse(&User{}, sqlite3Dialector)
	if schema.Name != "User" && len(schema.Fields) != 2 {
		t.Fatal("Failed to parse user schema.")
	}
	assert.Equal(t, schema.GetField("Name").Tag, "name")
	assert.Equal(t, schema.GetField("Name").Name, "Name")
}
