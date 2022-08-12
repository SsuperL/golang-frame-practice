package surpc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"surpc/codec"
	"sync"
	"time"
)

// Call 进行调用的条件,方法必须对外可见
// 包含两个参数，一个为参数，一个用于接收返回值
type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Done          chan *Call
	Error         error
}

func (call *Call) done() {
	// 用于支持异步调用，调用结束时通知调用方
	call.Done <- call
}

// Client sturcture of client
type Client struct {
	cc  codec.Codec
	opt *Option
	seq uint64
	// 互斥锁，用于保证请求的有序发送
	sending sync.Mutex
	mu      sync.Mutex
	header  codec.Header
	// 用于存放未处理完的请求，key是请求call.Seq，值是call实例
	pending map[uint64]*Call
	// closing和shutdown任一为true表示不可用，closing表示主动关闭
	closing bool
	// 表示出现故障，或者发生错误
	shutdown bool
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("Connection is shutdown")

// Close close the connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.shutdown {
		return ErrShutdown
	}

	c.closing = true
	return c.cc.Close()
}

// IsAvailable 判断client是否可用
func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.shutdown && !c.closing
}

// 注册请求并返回call序列号Seq
func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.shutdown || c.closing {
		return 0, ErrShutdown
	}
	// 注册未处理请求
	call.Seq = c.seq
	c.pending[call.Seq] = call
	c.seq++
	return call.Seq, nil
}

func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	call := c.pending[seq]
	delete(c.pending, seq)
	return call
}

// 客户端或服务端发生错误时调用
func (c *Client) terminateCalls(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shutdown = true
	for _, call := range c.pending {
		call.Error = err
		call.done()
	}
}

// 接收响应
func (c *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		// 解析reply请求头出错
		if err = c.cc.ReadHeader(&h); err != nil {
			break
		}
		call := c.removeCall(h.Seq)
		switch {
		//请求已经被移除或服务端处理时部分出错
		case call == nil:
			err = c.cc.ReadBody(nil)
		case h.Err != "":
			// reply包含错误信息
			call.Error = fmt.Errorf(h.Err)
			err = c.cc.ReadBody(nil)
			call.done()
		default:
			// 请求结果为正常响应
			err = c.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body error: " + err.Error())
			}
			call.done()

		}
	}

	c.terminateCalls(err)

}

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	f := codec.CodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type: %s", opt.CodecType)
		log.Println("rpc client: codec error: ", err)
		return nil, err
	}

	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}

	// fmt.Println(opt)
	return newCientCodec(f(conn), opt), nil

}

func newCientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		cc:      cc,
		opt:     opt,
		seq:     1,
		pending: make(map[uint64]*Call),
	}
	go client.receive()

	return client
}

func parseOptions(opts ...*Option) (*Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return &DefaultOption, nil
	}

	if len(opts) != 1 {
		return nil, fmt.Errorf("number of options is greater than 1")
	}

	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}

	return opt, nil

}

type clientResult struct {
	client *Client
	err    error
}

type newClientFunc func(conn net.Conn, opt *Option) (client *Client, err error)

// 超时机制，连接超时
// 入参f为创建client的初始化函数NewClient
func dialTimeout(f newClientFunc, network, addr string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	// fmt.Printf("dial opt: %v", opt)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout(network, addr, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}

	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()

	ch := make(chan clientResult)
	go func() {
		client, err := f(conn, opt)
		ch <- clientResult{client: client, err: err}
	}()

	if opt.ConnectTimeout == 0 {
		result := <-ch
		return result.client, result.err
	}

	select {
	// 如果超时则返回错误
	case <-time.After(opt.ConnectTimeout):
		return nil, fmt.Errorf("rpc client: connect timeout, expected within %s", opt.ConnectTimeout)
	case result := <-ch:
		return result.client, result.err
	}

}

// Dial 连接至具体服务端
func Dial(network, addr string, opts ...*Option) (client *Client, err error) {
	return dialTimeout(NewClient, network, addr, opts...)
}

// 发送请求
func (c *Client) send(call *Call) {
	c.sending.Lock()
	defer c.sending.Unlock()

	seq, err := c.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// request header
	c.header.ServiceMethod = call.ServiceMethod
	c.header.Seq = seq
	c.header.Err = ""

	if c.cc.Write(&c.header, call.Args); err != nil {
		call := c.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

// Go 异步接口，暴露给用户调用
func (c *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered.")
	}

	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}

	// fmt.Printf("c: %#v \n", c)
	c.send(call)
	return call
}

// Call 暴露给用户调用
// 使用context实现超时机制，context控制权交由用户
func (c *Client) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	call := c.Go(serviceMethod, args, reply, make(chan *Call, 1))
	select {
	case <-ctx.Done():
		c.removeCall(call.Seq)
		return errors.New("rpc client: call failed: " + ctx.Err().Error())
	case call := <-call.Done:
		return call.Error
	}
}

func NewHTTPClient(conn net.Conn, opt *Option) (*Client, error) {
	io.WriteString(conn, fmt.Sprintf("CONNECT %s HTTP/1.0\n\n", defaultRPCPath))

	// 切换至RPC响应之前，需要接受HTTP响应
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == connected {
		return NewClient(conn, opt)
	}

	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}

	return nil, err
}

// DialHTTP 发起HTTP请求
func DialHTTP(network, addr string, opts ...*Option) (*Client, error) {
	return dialTimeout(NewHTTPClient, network, addr, opts...)
}

func XDial(rpcAddr string, opts ...*Option) (*Client, error) {
	parts := strings.Split(rpcAddr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("rpc client err: wrong format '%s', expected protocol@addr", rpcAddr)
	}

	protocol, addr := parts[0], parts[1]
	switch protocol {
	case "http":
		return DialHTTP("tcp", addr, opts...)
	default:
		return Dial(protocol, addr, opts...)
	}
}
