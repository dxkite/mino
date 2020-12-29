package transport

import (
	"errors"
	"io"
	"net"
	"sync"
)

// 会话
type Session struct {
	// 会话ID
	Id int `json:"id"`
	// 分组
	Group string `json:"group"`
	// 来源
	Src string `json:"src"`
	// 目的
	Dst string `json:"dst"`
	// 上传
	Up int64 `json:"up"`
	// 下载
	Down int64 `json:"down"`
	// 关闭
	Closed bool `json:"closed"`
	// 本地连接
	loc io.ReadWriteCloser
	// 远程连接
	rmt io.ReadWriteCloser
	// 会话错误
	err error
	// 关闭状态
	mtxClosed sync.Mutex
	// 错误设置
	mtxErr sync.Mutex
	// 关闭通知
	chClosed chan struct{}
	chRead   chan int64
	chWrite  chan int64
}

// 传输数据
func (s *Session) Transport() (up, down int64, err error) {
	var _closed = make(chan struct{})
	go func() {
		// remote -> local
		var _err error
		if down, _err = io.Copy(s, s.rmt); _err != nil {
			s.rwErr("read", _err)
		}
		_closed <- struct{}{}
	}()
	// local -> remote
	var _err error
	if up, _err = io.Copy(s.rmt, s); _err != nil {
		s.rwErr("write", _err)
	}
	_ = s.close()
	<-_closed
	err = s.err
	// 通知关闭
	s.notifyClose()
	return
}

func (s *Session) notifyClose() {
	go func() { s.chClosed <- struct{}{} }()
}

func (s *Session) CloseNotify() <-chan struct{} {
	return s.chClosed
}
func (s *Session) ReadNotify() <-chan int64 {
	return s.chRead
}
func (s *Session) WriteNotify() <-chan int64 {
	return s.chWrite
}

func (s *Session) close() error {
	if !s.Closed {
		s.mtxClosed.Lock()
		s.Closed = true
		s.mtxClosed.Unlock()
		_ = s.loc.Close()
		_ = s.rmt.Close()
	}
	return s.err
}

func (s *Session) rwErr(name string, err error) {
	s.mtxErr.Lock()
	defer s.mtxErr.Unlock()
	if !s.Closed && err != nil {
		s.err = errors.New("tunnel error:" + name + ":" + err.Error())
		_ = s.close()
	}
}

func NewSession(sid int, group string, loc, rmt net.Conn, dst string) *Session {
	return &Session{
		loc:      loc,
		rmt:      rmt,
		chClosed: make(chan struct{}),
		chRead:   make(chan int64),
		chWrite:  make(chan int64),
		Id:       sid,
		Group:    group,
		Src:      loc.RemoteAddr().String(),
		Dst:      dst,
		Up:       0,
		Down:     0,
		Closed:   false,
	}
}

func (f *Session) Read(p []byte) (n int, err error) {
	n, err = f.loc.Read(p)
	f.Up += int64(n)
	go func() {
		if !f.Closed {
			f.chRead <- f.Up
		}
	}()
	return
}

func (f *Session) Write(p []byte) (n int, err error) {
	n, err = f.loc.Write(p)
	f.Down += int64(n)
	go func() {
		if !f.Closed {
			f.chWrite <- f.Down
		}
	}()
	return
}

type SessionGroup map[string]map[string]*Session

func NewSessionGroup() SessionGroup {
	return map[string]map[string]*Session{}
}

func (sg SessionGroup) AddSession(group, id string, session *Session) {
	if sg[group] == nil {
		sg[group] = map[string]*Session{}
	}
	sg[group][id] = session
}

func (sg SessionGroup) DelSession(group, id string) {
	delete(sg[group], id)
}
