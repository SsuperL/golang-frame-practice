package surpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"surpc/codec"
	"sync"
)

// Call 进行调用的条件,方法必须对外可见
// 包含两个参数，一个为参数，一个用于接收返回值
type Call struct {
	Seq           int
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
	seq int
	// 互斥锁，用于保证请求的有序发送
	sending sync.Mutex
	mu      sync.Mutex
	header  codec.Header
	// 用于存放未处理完的请求，key是请求call.Seq，值是call实例
	pending map[int]*Call
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

// 注册请求并返回client序列号seq
func (c *Client) registerCall(call *Call) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.shutdown || c.closing {
		return 0, ErrShutdown
	}
	// 注册未处理请求
	c.pending[call.Seq] = call
	c.seq++
	return c.seq, nil
}

func (c *Client) removeCall(seq int) *Call {
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
func (c *Client) recieve() {
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
			err = errors.New(h.Err)
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

	return newCientCodec(f(conn), opt), nil

}

func newCientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		cc:      cc,
		opt:     opt,
		seq:     1,
		pending: make(map[int]*Call),
	}
	go client.recieve()

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

// Dial 连接至具体服务端
func Dial(network, addr string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()

	return NewClient(conn, opt)
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

	c.send(call)
	return call
}

//Call 暴露给用户调用
func (c *Client) Call(serviceMethod string, args, reply interface{}) error {
	call := <-c.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}
