package surpc

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

func TestClient_Timeout(t *testing.T) {
	t.Parallel()
	l, _ := net.Listen("tcp", ":0")
	f := func(conn net.Conn, option *Option) (client *Client, err error) {
		_ = conn.Close()
		time.Sleep(time.Second)
		return nil, nil
	}

	t.Run("timeout", func(t *testing.T) {
		_, err := dialTimeout(f, "tcp", l.Addr().String(), &Option{ConnectTimeout: time.Second})
		_assert(err != nil && strings.Contains(err.Error(), "connect timeout"), "")
	})

	t.Run("0", func(t *testing.T) {
		_, err := dialTimeout(f, "tcp", l.Addr().String(), &Option{ConnectTimeout: 0})
		_assert(err == nil, "0 means no limit")
	})
}

type Bar int

func (b Bar) Timeout(argv int, replyv *int) error {
	time.Sleep(time.Second)
	return nil
}

func startServer(addr chan string) {
	var b Bar
	Register(&b)
	lis, _ := net.Listen("tcp", ":0")
	addr <- lis.Addr().String()
	Accept(lis)
}

func TestTimeout_client(t *testing.T) {
	t.Parallel()
	addr := make(chan string)
	go startServer(addr)
	time.Sleep(time.Second)
	// 客户端设置超时时间，服务端无限制
	t.Run("connect timeout", func(t *testing.T) {
		client, _ := Dial("tcp", <-addr)
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		var reply int
		err := client.Call(ctx, "Bar.Timeout", 1, &reply)
		_assert(err != nil && strings.Contains(err.Error(), ctx.Err().Error()), "expect a timeout error")
	})

	// 服务端设置超时时间，客户端无限制
	t.Run("handle timeout", func(t *testing.T) {
		client, _ := Dial("tcp", <-addr, &Option{HandleTimeout: time.Second})
		var reply int
		err := client.Call(context.Background(), "Bar.Timeout", 1, &reply)
		_assert(err != nil && strings.Contains(err.Error(), "handle timeout"), "handle timeout")
	})
}
