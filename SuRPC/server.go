package surpc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"surpc/codec"
	"sync"
)

const MagicNumber = 0x3bef5c

// Option 用于协议协商，后续Header和Body由CodecType决定
// | Option{MgicNumber: 1,CodecType:2} | Header{} | Body{} | Header{} | Body{} ...
// Option 暂定使用JSON编码
type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

var DefaultOption = Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

type Server struct{}

type request struct {
	h *codec.Header
	// 返回结果的入参和返回参数
	argv, replyv reflect.Value
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
	var option Option
	// json解码获得Option
	if err := json.NewDecoder(conn).Decode(&option); err != nil {
		log.Println("Error decoding option, error: ", err)
		return
	}
	log.Println("option:", option)
	// 校验magicnumber
	if option.MagicNumber != MagicNumber {
		log.Println("Wrong magicNumber, ", option.MagicNumber)
		return
	}

	// 校验codecType
	f := codec.CodecFuncMap[option.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type : %s", option.CodecType)
		return
	}

	s.ServerCodec(f(conn))
}

func (s *Server) ServerCodec(cc codec.Codec) {
	// 使用锁保证回复请求的报文是逐个发送的，不然会导致多个报文交织在一起，导致客户端无法解析
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	// 只有在header解析失败时才终止循环
	for {
		// 读取请求
		req, err := s.readRequest(cc)
		log.Println("request:", req)
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
		go s.handleRequest(cc, req, req.h, sending, wg)
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
	fmt.Println("h:", h)
	if err != nil {
		return nil, err
	}

	req := &request{h: h}
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server error: read argv error: ", err)
		// return nil, err
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
	fmt.Println("readHeader h:", h)
	return &h, nil
}

// 处理请求
func (s *Server) handleRequest(cc codec.Codec, req *request, h *codec.Header, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println(req.h, req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("surpc resp %d", req.h.Seq))
	s.sendResponse(cc, h, req.replyv.Interface(), sending)
}

// 返回响应结果
func (s *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error", err)
	}
}
