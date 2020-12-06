package session

import (
	"io"
)

// 连接会话
type Session struct {
	// 本地连接
	loc io.ReadWriteCloser
	// 远程连接
	rmt io.ReadWriteCloser

	// 关闭状态
	sndError error
	revError error
}

// 创建会话
func NewSession(loc, rmt io.ReadWriteCloser) *Session {
	return &Session{
		loc: loc,
		rmt: rmt,
	}
}

// 传输数据
func (s *Session) Transport() (up, down int64) {
	var _closed = make(chan struct{})
	go func() {
		// send local -> remote
		up, s.sndError = io.Copy(s.rmt, s.loc)
		_closed <- struct{}{}
	}()
	go func() {
		// send remote -> down
		down, s.revError = io.Copy(s.loc, s.rmt)
		_closed <- struct{}{}
	}()
	<-_closed
	_ = s.loc.Close()
	_ = s.rmt.Close()
	return
}
