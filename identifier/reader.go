package identifier

import (
	"io"
)

func NewBufferedReader(buf []byte, used int, r io.Reader) io.Reader {
	return &BufferedReader{
		buf:  buf,
		used: used,
		r:    r,
	}
}

type BufferedReader struct {
	buf  []byte
	rd   int
	used int
	r    io.Reader
}

func (c *BufferedReader) Read(p []byte) (int, error) {
	if c.rd >= c.used {
		return c.r.Read(p)
	}

	n := copy(p, c.buf[c.rd:])
	c.rd += n
	return n, nil
}
