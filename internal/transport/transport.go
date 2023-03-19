package transport

import (
	"io"
	"net"
)

type Transport struct {
	src, dst net.Conn
}

func CreateTransport(src, dst net.Conn) *Transport {
	return &Transport{src: src, dst: dst}
}

func (t *Transport) DoTransport() (up, down int64, err error) {
	var closeCh = make(chan struct{})
	var errCh = make(chan error)

	go func() {
		// remote -> local
		var _err error
		if down, _err = io.Copy(t.src, t.dst); _err != nil {
			errCh <- _err
			return
		}
		closeCh <- struct{}{}
	}()

	go func() {
		// local -> remote
		var _err error
		if up, _err = io.Copy(t.dst, t.src); _err != nil {
			errCh <- _err
			return
		}
		closeCh <- struct{}{}
	}()

	select {
	case err = <-errCh:
		return
	case <-closeCh:
		return
	}
}
