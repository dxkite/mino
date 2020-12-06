package rewind

import (
	"bytes"
	"io"
	"testing"
)

func Test_rewindReader_Read(t *testing.T) {
	buf1 := NewRewindReaderSize(bytes.NewReader([]byte("dxkite12345")), 6)
	readNEqual(buf1, 2, "dx", nil, t)
	readNEqual(buf1, 4, "kite", nil, t)
	readNEqual(buf1, 5, "12345", nil, t)
	if e := buf1.Rewind(); e != ErrRewindSize {
		t.Error("want", ErrRewindSize, "got", e)
	}
	buf2 := NewRewindReaderSize(bytes.NewReader([]byte("dxkite12345678910")), 6)
	readNEqual(buf2, 2, "dx", nil, t)
	if e := buf2.Rewind(); e != nil {
		t.Error("want", ErrRewindSize, "got", e)
	}
	readFullNEqual(buf2, 6, "dxkite", nil, t)
	if e := buf2.Rewind(); e != nil {
		t.Error("want", ErrRewindSize, "got", e)
	}
	readFullNEqual(buf2, 6, "dxkite", nil, t)
	if e := buf2.Rewind(); e != nil {
		t.Error("want", ErrRewindSize, "got", e)
	}
	readFullNEqual(buf2, 11, "dxkite12345", nil, t)
	if e := buf2.Rewind(); e != ErrRewindSize {
		t.Error("want", ErrRewindSize, "got", e)
	}
	readFullNEqual(buf2, 6, "678910", nil, t)
	if e := buf2.Rewind(); e != ErrRewindSize {
		t.Error("want", ErrRewindSize, "got", e)
	}
	readFullNEqual(buf2, 6, "", io.EOF, t)

	buf3 := NewRewindReaderSize(bytes.NewReader([]byte("dxkite12345")), 6)
	readNEqual(buf3, 100, "dxkite12345", nil, t)
}

func readNEqual(r io.Reader, n int, rd string, err error, t *testing.T) {
	buf := make([]byte, n)
	rs, e := r.Read(buf)
	if e != err {
		t.Error("read", n, "real read", rs, "want error", err, "got", e)
	}
	if string(buf[:rs]) != rd {
		t.Error("read", n, "real read", rs, "want", rd, "got", string(buf[:rs]))
	}
}

func readFullNEqual(r io.Reader, n int, rd string, err error, t *testing.T) {
	buf := make([]byte, n)
	rs, e := io.ReadFull(r, buf)
	if e != err {
		t.Error("read", n, "real read", rs, "want error", err, "got", e)
	}
	if string(buf[:rs]) != rd {
		t.Error("read", n, "real read", rs, "want", rd, "got", string(buf[:rs]))
	}
}
