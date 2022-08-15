package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

// 服务发现与负载均衡策略，实现随机选择策略与Round Robin策略

// SelectMode 选择策略
type SelectMode int

const (
	//RandomSelect 随机选择策略
	RandomSelect SelectMode = iota
	// RoundRobinSelect Robbin轮询策略，每次调度选择 i=(i+1)mod n
	RoundRobinSelect
)

// Discovery discovery servers
type Discovery interface {
	// Refresh 从注册中心更新服务列表
	Refresh() error
	// Update 更新服务列表
	Update(servers []string) error
	// Get 获取服务实例
	Get(mode SelectMode) (string, error)
	// GetAll 获取所有服务实例
	GetAll() ([]string, error)
}

var _ Discovery = (*MultiServerDiscovery)(nil)

// MultiServerDiscovery 不需要注册中心，手动维护服务实例
type MultiServerDiscovery struct {
	// 生成随机数
	rand    *rand.Rand
	mu      sync.RWMutex
	servers []string
	// 记录Robbin算法轮询到的位置
	index int
}

func NewMultiServerDiscovery(servers []string) *MultiServerDiscovery {
	d := &MultiServerDiscovery{
		servers: servers,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.rand.Intn(math.MaxInt32 - 1)
	return d
}

// Refresh 从注册中心更新服务列表
func (m *MultiServerDiscovery) Refresh() error {
	return nil
}

// Update 更新服务列表
func (m *MultiServerDiscovery) Update(servers []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers = servers
	return nil
}

// Get 根据策略模式获取服务实例
func (m *MultiServerDiscovery) Get(mode SelectMode) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	length := len(m.servers)
	if length == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}
	switch mode {
	case RandomSelect:
		return m.servers[m.rand.Intn(length)], nil
	case RoundRobinSelect:
		s := m.servers[m.index%length]
		m.index = (m.index + 1) % length
		return s, nil
	default:
		return "", errors.New("rpc discovery: not supported mode")
	}
}

// GetAll 获取所有服务实例
func (m *MultiServerDiscovery) GetAll() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	servers := make([]string, len(m.servers), len(m.servers))
	copy(servers, m.servers)
	return servers, nil
}
