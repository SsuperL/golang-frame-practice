package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

// GobCodec is a codec of gob encoding
type GobCodec struct {
	// 连接实例
	conn    io.ReadWriteCloser
	encoder *gob.Encoder
	decoder *gob.Decoder
	// 缓冲，防止阻塞
	buf *bufio.Writer
}

var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		// 数据写入buf来优化性能
		encoder: gob.NewEncoder(buf),
		// 直接从conn中读取内容
		decoder: gob.NewDecoder(conn),
		buf:     buf,
	}
}

// ReadBody read body from request
func (g *GobCodec) ReadBody(body interface{}) error {
	return g.decoder.Decode(body)
}

// ReadHeader read header from request
func (g *GobCodec) ReadHeader(h *Header) error {
	return g.decoder.Decode(h)
}

func (g *GobCodec) Write(h *Header, body interface{}) (err error) {
	defer func() {
		// 写入缓冲区
		_ = g.buf.Flush()
		if err != nil {
			g.Close()
			log.Println("flush to buf error:", err)
		}
	}()

	if err = g.encoder.Encode(h); err != nil {
		log.Println("Failed to encode header. err: ", err)
		return
	}
	if err = g.encoder.Encode(body); err != nil {
		log.Println("Failed to encode body. err: ", err)
		return
	}

	return
}

// Close close conn
func (g *GobCodec) Close() error {
	return g.conn.Close()
}
