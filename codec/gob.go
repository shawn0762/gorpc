package codec

import (
	"bufio"
	"encoding/gob"
	"fmt"
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

// 这里利用强制转换的特性确保GobCodec已经实现了codec.Codec接口
// 如果GobCodec没有实现该接口，无法通过编译
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

// 关闭连接
func (g *GobCodec) Close() error {
	return g.conn.Close()
}

// 从连接中读取header内容
func (g *GobCodec) ReadHeader(h *Header) error {
	return g.dec.Decode(h)
}

// 从连接中读取body内容
func (g *GobCodec) ReadBody(v interface{}) error {
	return g.dec.Decode(v)
}

func (g *GobCodec) Write(h *Header, v interface{}) (err error) {
	defer func() {
		// 将缓冲区的内容写入连接
		// @todo 多次write才刷一次？
		_ = g.buf.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()

	if err := g.enc.Encode(h); err != nil {
		fmt.Println("Rpc codec: gob encoding header err: ", err)
		return err
	}

	if err := g.enc.Encode(v); err != nil {
		fmt.Println("Rpc codec: gob encoding body err: ", err)
		return err
	}
	return nil
}

