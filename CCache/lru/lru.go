package lru

import (
	"container/list"
)

// Cache structure to implement LRU
/*
LRU(least Recently Used) lru缓存淘汰策略
维护一个队列，如果记录被访问了，则移至队尾，队首的元素是最近最少访问的，可以被移除
*/
type Cache struct {
	// 已使用内存
	usedBytes int64
	// 最大内存, 为0时表示不设限
	maxBytes int64
	// 元素指向底层双向链表中的节点，保存键值映射关系
	cache map[string]*list.Element
	// 底层双向链表，保存值
	linkedList *list.List
	// 当entry（访问记录）被移除时执行
	onEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

// Value 计算使用了多少内存
type Value interface {
	Len() int
}

// New initiate
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:   maxBytes,
		linkedList: list.New(),
		cache:      make(map[string]*list.Element),
		onEvicted:  onEvicted,
	}
}

// Get get element from cache
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// 如果能在cache中查找到对应key，将该key对应的元素移至队尾（假设front是队尾）
		c.linkedList.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

// RemoveOldest 删除最近最久未被使用的队首
func (c *Cache) RemoveOldest() {
	ele := c.linkedList.Back()
	if ele != nil {
		// 从底层双向链表中移除对应节点
		c.linkedList.Remove(ele)
		kv := ele.Value.(*entry)
		// 从cache中删除对应key
		delete(c.cache, kv.key)
		c.usedBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}
	}
}

// Add add entry
func (c *Cache) Add(key string, value Value) {
	// 不存在记录则添加至队尾，存在则更新
	if ele, ok := c.cache[key]; !ok {
		ele := c.linkedList.PushFront(&entry{key: key, value: value})
		c.usedBytes += int64(len(key)) + int64(value.Len())
		// 关联cache中key和linkedlist节点
		c.cache[key] = ele
	} else {
		// 访问次数增加，将节点移至队首
		c.linkedList.MoveToFront(ele)
		kv := ele.Value.(*entry)
		// 更新节点增加的容量
		c.usedBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	}

	// 如果已使用容量超过最大容量，移除队首（最近最久未使用）的节点
	for c.maxBytes != 0 && c.usedBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

// Len return length of linkedList
func (c *Cache) Len() int {
	return c.linkedList.Len()
}
