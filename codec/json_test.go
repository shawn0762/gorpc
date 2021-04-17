package codec

import "testing"

func TestJson_Write(t *testing.T) {
	mockConn := &ReadwritecloserTest{}
	jsonCodec := NewJsonCodec(mockConn)

	err := jsonCodec.Write(nil, nil)
	if err != nil {
		t.Error("Err should be nil")
	}
}

// 构造一个实现ReadWriteCloser接口的结构体用于测试
type ReadwritecloserTest struct {
}

func (r *ReadwritecloserTest) Read(p []byte) (n int, err error) { return }

func (r *ReadwritecloserTest) Write(p []byte) (n int, err error) { return }

func (r *ReadwritecloserTest) Close() error { return nil }
