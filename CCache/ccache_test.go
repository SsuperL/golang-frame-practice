package ccache

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var db = map[string]string{
	"A": "A",
	"B": "B",
	"C": "C",
}

type peer struct{}

func (p *peer) Get(key string) ([]byte, error) {
	return []byte(db[key]), nil
}

func TestGetterFunc(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Fatalf("callback failed.")
	}
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	key := "A"
	// 没有缓存的情况下
	ccache := NewGroup("test", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				if _, exists := loadCounts[key]; !exists {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, nil
		}))

	value, err := ccache.Get(key)
	assert.Equal(t, err, nil)
	assert.Equal(t, db[key], value.String())
	assert.Equal(t, 1, loadCounts[key])

	// 已缓存
	_, _ = ccache.Get(key)
	assert.Equal(t, 1, loadCounts[key])
}
