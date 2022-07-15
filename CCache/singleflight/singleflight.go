package singleflight

/*
Package singleflight
保证重复请求只执行一次，防止缓存击穿
*/
import "sync"

// call 表示正在进行的或已经完成的Do请求
type call struct {
	wg  sync.WaitGroup // 保证并发
	val interface{}
	err error
}

// Group 用于保证请求只执行一次
type Group struct {
	mu sync.Mutex       // guards
	m  map[string]*call //延迟初始化
}

// Do 确保重复发起的Do请求（fn函数）只执行一次
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok { // 是否有正在进行的请求
		g.mu.Unlock()
		c.wg.Wait()         // 如果请求正在进行中，则等待
		return c.val, c.err // 请求结束，返回结果
	}

	c := new(call)
	c.wg.Add(1) // 发起请求前加锁
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done() // 请求结束

	g.mu.Lock()
	delete(g.m, key) // 更新g.m
	g.mu.Unlock()

	return c.val, c.err
}
