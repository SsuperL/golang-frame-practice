package ccache

import (
	"log"
	"sync"
)

// Getter 缓存未命中时，获取源数据的回调函数，暴露给用户自定义，可定义多个适配器
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc callback func
type GetterFunc func(key string) ([]byte, error)

// Group namespace of cache
type Group struct {
	// name of namespace
	name string
	// callback when not hit cache
	getter    Getter
	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// Get callback
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// NewGroup create a group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}

	groups[name] = g
	return g
}

// GetGroup get a group
func GetGroup(name string) *Group {
	// read lock
	mu.RLock()
	defer mu.RUnlock()

	g := groups[name]
	return g
}

// Get value from cache if exists, else get value from other resources using callback function
func (g *Group) Get(key string) (ByteView, error) {
	v, ok := g.mainCache.get(key)
	if ok {
		log.Println("Cache hit")
		return v, nil
	} else {
		return g.load(key)
	}
}

// 单机调用
func (g *Group) getLocally(key string) (ByteView, error) {
	b, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(b)}
	// write cache
	g.populateCache(key, value)

	return value, nil
}

func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}