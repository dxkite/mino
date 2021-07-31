package rewind

import (
	"errors"
	"io"
)

type Reader interface {
	io.Reader
	Rewind() error
	Cached() []byte
}

// 预读取的数据大于缓存区的数据不能回退
var ErrRewindSize = errors.New("rewind size error")

type rewindReader struct {
	r   io.Reader // 读取院
	buf []byte    // 缓存的位置
	rd  int       // 读取到的位置
	use int       // 缓冲区使用的大小
	max int       // 缓冲区大小
}

// 创建可缓冲
func NewRewindReaderSize(r io.Reader, size int) Reader {
	return &rewindReader{
		r:   r,
		buf: make([]byte, size),
		rd:  0,
		use: 0,
		max: size,
	}
}

// 读取数据
// 如果缓冲区有数据，读取缓冲区数据
// 如果超出缓存区数据，缓冲区失效，直接读取数据
func (rr *rewindReader) Read(p []byte) (n int, err error) {
	// 无缓冲区/缓冲区失效
	if len(rr.buf) == 0 {
		return rr.r.Read(p)
	}
	// 缓冲区能兜住部分数据
	if rr.rd < rr.use {
		lp := len(p)
		n := 0
		if rr.rd+lp > rr.use {
			// 读取的内容超过了缓冲区，只读取缓冲区有的数据
			n = copy(p, rr.buf[rr.rd:rr.use])
		} else {
			// 缓冲区能完全兜住
			n = copy(p, rr.buf[rr.rd:rr.rd+lp])
		}
		rr.rd += n
		return n, err
	}
	// 从数据源读取
	if n, er := rr.r.Read(p); er != nil {
		return 0, er
	} else {
		// 当前读取的数据大于缓冲区的数据
		if n+rr.rd > rr.max {
			// 缓冲区失效
			// 降级为普通读取
			rr.buf = rr.buf[0:0]
			return n, er
		}
		// 将数据复制到缓冲区
		copy(rr.buf[rr.rd:], p)
		rr.rd += n
		rr.use += n
		return n, nil
	}
}

// 重置读取位置，从头读取
func (rr *rewindReader) Rewind() error {
	if len(rr.buf) != rr.max {
		return ErrRewindSize
	}
	rr.rd = 0
	return nil
}

// 获取缓存数据
func (rr *rewindReader) Cached() []byte {
	return rr.buf[:rr.use]
}
