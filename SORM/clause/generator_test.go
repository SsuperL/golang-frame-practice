package clause

import (
	"testing"

	"github.com/go-playground/assert"
)

func TestSelect(t *testing.T) {
	sql, vars := _select("User", []string{"age", "name"})
	assert.Equal(t, "SELECT age,name FROM User", sql)
	assert.Equal(t, 0, len(vars))
}
