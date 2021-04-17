package codec

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

type Json struct {
	// 请求的connection实例
	conn io.ReadWriteCloser

	// 定义一个缓冲器
	buf *bufio.Writer

	enc *json.Encoder
	dec *json.Decoder
}

// 限制Json必须实现Codec接口，否则无法通过编译
var _ Codec = (*Json)(nil)

func NewJsonCodec(conn io.ReadWriteCloser) Codec {
	// 缓冲器buf临时存放要写入connection的内容
	// 最终buf中的内容要写到connection中
	buf := bufio.NewWriter(conn)
	return &Json{
		conn: conn,
		buf:  buf,
		enc:  json.NewEncoder(conn),
		dec:  json.NewDecoder(conn),
	}
}

func (j *Json) Close() error {
	return j.conn.Close()
}

// 从conn中读取header内容
func (j *Json) ReadHeader(header *Header) error {
	return j.dec.Decode(header)
}

// 从conn中读取body内容
func (j *Json) ReadBody(i interface{}) error {
	return j.dec.Decode(i)
}

// 将返回内容写入到conn
func (j *Json) Write(header *Header, i interface{}) (err error) {
	defer func() {
		// 写完后一次性写入conn
		err = j.buf.Flush()
		if err != nil {
			_ = j.conn.Close()
		}
	}()

	if err := j.enc.Encode(header); err != nil {
		fmt.Println("Rpc codec: gob encoding header err: ", err)
		return err
	}

	if err := j.enc.Encode(i); err != nil {
		fmt.Println("Rpc codec: gob encoding header err: ", err)
		return err
	}
	return nil
}
