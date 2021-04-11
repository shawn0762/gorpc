package codec

import (
	"bufio"
	"encoding/gob"
	"io"
)

// 通过gob实现的编码器
type GobCodec struct {
	// 请求的connection实例
	conn io.ReadWriteCloser

	// 定义一个缓冲器
	buf *bufio.Writer

	// 使用Gob的编/解码器
	dec *gob.Decoder
	enc *gob.Encoder
}

var _ Codec = (*GobCodec)(nil)

// 按照要求，实现一个构造函数
func NewGobCodec(conn io.ReadWriteCloser) Codec {
	// 缓冲器的内容最终要写入到connection，然后返回给客户端
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf: buf,
		dec: gob.NewDecoder(conn), // 从connection中读取并解码请求体
		enc: gob.NewEncoder(buf), // 这里不直接写入connection，而是先写到缓冲器，最后才一次性写入connection
	}
}

func (g *GobCodec) Close() error {
	panic("implement me")
}

func (g *GobCodec) ReadHeader(h *Header) error {
	panic("implement me")
}

func (g *GobCodec) ReadBody(h *Header) error {
	panic("implement me")
}

func (g *GobCodec) Write(h *Header, v interface{}) error {
	panic("implement me")
}

