package codec

import "io"

// Header request or response header
type Header struct {
	// ServiceMethod request method
	ServiceMethod string
	// Seq sequence number of request
	Seq uint64
	Err string
}

// Codec used to implement different codec
type Codec interface {
	io.Closer
	ReadBody(interface{}) error
	ReadHeader(*Header) error
	Write(*Header, interface{}) error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

var CodecFuncMap map[Type]NewCodecFunc

func init() {
	CodecFuncMap = make(map[Type]NewCodecFunc)
	CodecFuncMap[GobType] = NewGobCodec
}
