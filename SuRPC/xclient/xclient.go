package xclient

import (
	"context"
	"io"
	"reflect"
	. "surpc"
	"sync"
)

type XClient struct {
	d      Discovery
	mode   SelectMode
	option *Option
	mu     sync.Mutex
	// 保存成功创建的Client实例，复用socket连接
	clients map[string]*Client
}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, option *Option, mode SelectMode) *XClient {
	return &XClient{d: d, option: option, mode: mode, clients: make(map[string]*Client)}
}

func (x *XClient) Close() error {
	x.mu.Lock()
	defer x.mu.Unlock()
	for key, client := range x.clients {
		_ = client.Close()
		delete(x.clients, key)
	}
	return nil
}

func (x *XClient) dial(rpcAddr string) (*Client, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	// client是否在缓存列表中
	client, ok := x.clients[rpcAddr]
	// 存在缓存列表，但为不可用状态，删除该实例
	if ok && !client.IsAvailable() {
		client.Close()
		delete(x.clients, rpcAddr)
		client = nil
	}
	if client == nil {
		var err error
		client, err = XDial(rpcAddr, x.option)
		if err != nil {
			return nil, err
		}
		x.clients[rpcAddr] = client
	}
	return client, nil
}

func (x *XClient) call(ctx context.Context, rpcAddr string, serviceMethod string, args, reply interface{}) error {
	client, err := x.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}
func (x *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	rpcAddr, err := x.d.Get(x.mode)
	if err != nil {
		return err
	}
	return x.call(ctx, rpcAddr, serviceMethod, args, reply)
}

// Broadcast 广播转发请求
func (x *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	servers, err := x.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var e error
	replyDone := reply == nil
	ctx, cancel := context.WithCancel(ctx)
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var clonedReply interface{}
			if reply != nil {
				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := x.call(ctx, rpcAddr, serviceMethod, args, clonedReply)
			mu.Lock()
			if err != nil && e == nil {
				e = err
				cancel()
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}
			mu.Unlock()
		}(rpcAddr)
	}
	wg.Wait()
	return e
}
