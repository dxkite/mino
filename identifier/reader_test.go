package identifier

import (
	"bytes"
	"io"
	"testing"
)

func TestBufferReader_Read(t *testing.T) {
	readNEqual(
		NewBufferedReader([]byte("dxkite"), 6, bytes.NewReader([]byte("dxkite"))),
		255,
		"dxkitedxkite",
		nil,
		t)
	readNEqual(
		NewBufferedReader([]byte("dxkite"), 2, bytes.NewReader([]byte("dxkite"))),
		255,
		"dxdxkite",
		nil,
		t)

	readNEqual(
		NewBufferedReader([]byte("GET / "), 6, bytes.NewReader([]byte("HTTP/1.1"))),
		255,
		"GET / HTTP/1.1",
		nil,
		t)

	buf := NewBufferedReader([]byte("dxkite"), 6, bytes.NewReader([]byte("dxkite")))
	readNEqual(
		buf,
		2,
		"dx",
		nil,
		t)
	readNEqual(
		buf,
		255,
		"kitedxkite",
		nil,
		t)
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
