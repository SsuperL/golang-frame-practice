/*
封装lru缓存淘汰策略以及并发控制
*/
package ccache

import (
	"ccache/lru"
	"sync"
)

type cache struct {
	// 使用Mutex封装lru的方法
	mu         sync.Mutex // guards
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value lru.Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}

	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	v, ok := c.lru.Get(key)
	if !ok {
		return
	}

	return v.(ByteView), true
}
