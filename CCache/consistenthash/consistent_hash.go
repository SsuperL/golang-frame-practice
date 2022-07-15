package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 哈希函数
type Hash func([]byte) uint32

// Map 一致性哈希算法数据结构
type Map struct {
	hash     Hash           // 哈希函数，默认为crc32.ChecksumIEEE
	keys     []int          // 节点列表, 升序排列，便于后续查找最小节点
	hashMap  map[int]string // 虚拟节点与真实节点映射关系表，键为哈希值，值为真实节点名称
	replicas int            //虚拟节点倍数
}

// NewMap initiate map
func NewMap(replicas int, hash Hash) *Map {
	m := &Map{
		hash:     hash,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

// Add 添加真实节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hashed := m.hash([]byte(strconv.Itoa(i) + key))
			m.keys = append(m.keys, int(hashed))
			m.hashMap[int(hashed)] = key
		}
	}
	sort.Ints(m.keys)

}

// Get 获取距离最近的真实节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hashed := int(m.hash([]byte(key)))
	// 二分查找最小的节点（距离最近的)
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hashed
	})

	// keys是环状结构，使用取模来处理最小节点为0的情况
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
