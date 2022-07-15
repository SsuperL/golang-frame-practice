package ccache

import (
	"ccache/singleflight"
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
	peers     PeerPicker
	loadGroup *singleflight.Group
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
		loadGroup: &singleflight.Group{},
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
func (g *Group) Get(key string) (value ByteView, err error) {
	viewi, err := g.loadGroup.Do(key, func() (interface{}, error) {
		if v, ok := g.mainCache.get(key); ok {
			log.Println("Cache hit")
			return v, nil
		}
		return g.load(key)
	})

	if err == nil {
		return viewi.(ByteView), err
	}

	return

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

func (g *Group) load(key string) (value ByteView, err error) {
	// 从远程节点获取值
	if peer, ok := g.peers.PickPeer(key); ok {
		value, err = g.getFromPeer(peer, key)
		return
	}
	return g.getLocally(key)
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("register peers called more than once")
	}
	g.peers = peers
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	value, err := peer.Get(g.name, key)
	if err != nil {
		log.Println("get value from peer failed")
		return ByteView{}, err
	}
	return ByteView{value}, nil
}
