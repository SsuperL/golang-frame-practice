package surpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"surpc/codec"
	"sync"
	"time"
)

const MagicNumber = 0x3bef5c
const (
	connected        = "200 connected to SuRPC"
	defaultRPCPath   = "/_surpc_"
	defaultDebugPath = "/debug/surpc"
)

// Option 用于协议协商，后续Header和Body由CodecType决定
// | Option{MgicNumber: 1,CodecType:2} | Header{} | Body{} | Header{} | Body{} ...
// Option 暂定使用JSON编码
type Option struct {
	MagicNumber    int
	CodecType      codec.Type
	ConnectTimeout time.Duration // 连接超时时间，默认10s, 0表示不设限
	HandleTimeout  time.Duration // 处理请求超时时间，默认 0ms，0表示不设限
}

var DefaultOption = Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.GobType,
	ConnectTimeout: time.Second * 10,
}

type Server struct {
	serviceMap sync.Map
}

type request struct {
	h *codec.Header
	// 返回结果的入参和返回参数
	argv, replyv reflect.Value
	mType        *methodType
	svc          *service
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()
var invalidRequest = struct{}{}

func (s *Server) Accept(lis net.Listener) {
	for {
		// 循环等待建立socket连接
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc serve accept error: ", err.Error())
			return
		}

		// 连接建立后，交由子协程处理
		go s.ServeConn(conn)
	}
}

func (s *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var option Option
	// json解码获得Option
	if err := json.NewDecoder(conn).Decode(&option); err != nil {
		fmt.Printf("option: %v \n", option)
		log.Println("rpc server: decoding option, error: ", err)
		return
	}
	// 校验magicnumber
	if option.MagicNumber != MagicNumber {
		log.Println("rpc server: wrong magicNumber, ", option.MagicNumber)
		return
	}

	// 校验codecType
	f := codec.CodecFuncMap[option.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type : %s", option.CodecType)
		return
	}

	s.ServerCodec(f(conn), &option)
}

func (s *Server) ServerCodec(cc codec.Codec, option *Option) {
	// 使用锁保证回复请求的报文是逐个发送的，不然会导致多个报文交织在一起，导致客户端无法解析
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	// 只有在header解析失败时才终止循环
	for {
		// 读取请求
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Err = err.Error()
			// 终止本次循环，返回结果
			s.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		// 并发处理请求
		wg.Add(1)
		go s.handleRequest(cc, req, req.h, sending, wg, option.HandleTimeout)
	}

	wg.Wait()
	cc.Close()

}
func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

// 读取请求
func (s *Server) readRequest(cc codec.Codec) (*request, error) {
	// 读取请求头
	h, err := s.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}

	req := &request{h: h}
	req.svc, req.mType, err = s.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}

	// 创建入参实例
	req.argv = req.mType.newArgv()
	req.replyv = req.mType.newReplyv()
	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}

	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server error: read argv error: ", err)
		return req, err
	}

	return req, nil
}

// read header
func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server error: read header error :", err)
		}
		return nil, err
	}
	return &h, nil
}

// 处理请求
func (s *Server) handleRequest(cc codec.Codec, req *request, h *codec.Header, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	called := make(chan struct{})
	sent := make(chan struct{})
	go func() {
		err := req.svc.call(req.mType, req.argv, req.replyv)
		called <- struct{}{}
		if err != nil {
			req.h.Err = err.Error()
			s.sendResponse(cc, req.h, invalidRequest, sending)
			sent <- struct{}{}
			return
		}

		s.sendResponse(cc, req.h, req.replyv.Interface(), sending)
		sent <- struct{}{}
	}()

	if timeout == 0 {
		<-called
		<-sent
		return
	}

	select {
	// 超时
	case <-time.After(timeout):
		req.h.Err = fmt.Errorf("rpc server: request handle time out, expected within %s", timeout).Error()
		s.sendResponse(cc, req.h, invalidRequest, sending)
	case <-called:
		<-sent
	}
}

// 返回响应结果
func (s *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error", err)
	}
}

func (s *Server) Register(rcvr interface{}) error {
	service := newService(rcvr)
	if _, dup := s.serviceMap.LoadOrStore(service.name, service); dup {
		return errors.New("rpc server: service already defined" + service.name)
	}
	return nil
}

func Register(rcvr interface{}) error {
	return DefaultServer.Register(rcvr)
}

// 通过serviceMethod找到对应的service
func (s *Server) findService(serviceMethod string) (svc *service, mType *methodType, err error) {
	// serviceName.methodName (eg."Foo.Sum")
	dotIndex := strings.LastIndex(serviceMethod, ".")
	if dotIndex < 0 {
		err = errors.New("rpc server: service method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dotIndex], serviceMethod[dotIndex+1:]

	svci, ok := s.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can not find service: " + serviceName)
		return
	}
	svc = svci.(*service)
	mType = svc.method[methodName]
	if mType == nil {
		err = errors.New("rpc server: can not find method: " + methodName)
	}
	return
}

// ServeHTTP 接收客户端的HTTP请求，并转换为RPC请求
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 Method Not Allowed")
		return
	}
	// 使用Hijacker接管（劫持）连接，调用后由调用者管理和关闭连接，net/http不再管理连接
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Println("rpc hijacking ", r.RemoteAddr, ":", err.Error())
		return
	}

	_, _ = io.WriteString(conn, " HTTP/1.0 "+connected+"\n\n")
	fmt.Printf("conn: -- %v ", conn)
	s.ServeConn(conn)
}

func (s *Server) HandleHTTP() {
	http.Handle(defaultRPCPath, s)
	http.Handle(defaultDebugPath, debugHTTP{s})
	log.Println("rpc server debug path: ", defaultDebugPath)
}

func HandleHTTP() {
	DefaultServer.HandleHTTP()
}
