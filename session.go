package mino

import (
	"io"
	"log"
	"sync"
)

// 连接会话
type Session struct {
	// 本地连接
	loc io.ReadWriteCloser
	// 远程连接
	rmt io.ReadWriteCloser
	// 会话错误
	err error
	// 会话关闭状态
	closed bool
	// 关闭状态
	mtxClosed sync.Mutex
	// 错误设置
	mtxErr sync.Mutex
}

// 创建会话
func NewSession(loc, rmt io.ReadWriteCloser) *Session {
	return &Session{
		loc: loc,
		rmt: rmt,
	}
}

// 传输数据
func (s *Session) Transport() (up, down int64, err error) {
	var _closed = make(chan struct{})
	go func() {
		// send local -> remote
		var _err error
		if up, _err = io.Copy(s.rmt, s.loc); _err != nil {
			s.rwErr(_err)
		}
		_ = s.Close()
		_closed <- struct{}{}
	}()
	go func() {
		//send remote -> down
		var _err error
		if down, _err = io.Copy(s.loc, s.rmt); _err != nil {
			s.rwErr(_err)
		}
		_ = s.Close()
		_closed <- struct{}{}
	}()
	<-_closed
	err = s.err
	return
}

func (s *Session) Close() error {
	if !s.closed {
		s.mtxClosed.Lock()
		s.closed = true
		s.mtxClosed.Unlock()
		_ = s.loc.Close()
		_ = s.rmt.Close()
	}
	return s.err
}

func (s *Session) rwErr(err error) {
	s.mtxErr.Lock()
	defer s.mtxErr.Unlock()
	if !s.closed && err != nil {
		log.Println("session read/write error", err)
		s.err = err
		_ = s.Close()
	}
}
