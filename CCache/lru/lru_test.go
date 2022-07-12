package lru

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	cache := New(10, nil)
	key, value := "test", String("value")
	cache.Add(key, value)
	v, ok := cache.Get(key)
	ele := cache.linkedList.Front()
	assert.Equal(t, v, value)
	assert.Equal(t, ele.Value.(*entry).value, value)
	assert.True(t, ok)

	notExist, ok := cache.Get("123")
	assert.Equal(t, notExist, nil)
	assert.Equal(t, ok, false)
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"
	cap := int64(len(k1+k2) + len(v1+v2))
	cache := New(cap, nil)
	cache.Add(k1, String(v1))
	cache.Add(k2, String(v2))
	cache.Add(k3, String(v3))

	assert.Equal(t, cache.Len(), 2)
	if _, ok := cache.Get(k1); ok || cache.Len() != 2 {
		t.Fatalf("RemoveOldest key1 failed.")
	}
}

func TestOnEvicted(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"
	cap := int64(len(k1+k2) + len(v1+v2))
	keys := make([]string, 0)
	cache := New(cap, func(key string, value Value) {
		keys = append(keys, key)
	})
	cache.Add(k1, String(v1))
	cache.Add(k2, String(v2))
	cache.Add(k3, String(v3))

	assert.Equal(t, 1, len(keys))
}

func TestUpdate(t *testing.T) {
	cache := New(10, nil)
	key, value := "test", String("value")
	cache.Add(key, value)
	newValue := String("value2")
	cache.Add(key, newValue)
	v, _ := cache.Get(key)
	assert.Equal(t, v, newValue)

}
