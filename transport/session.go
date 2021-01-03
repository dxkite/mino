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
	// 打印数据流
	dump bool
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

func (s *Session) Read(p []byte) (n int, err error) {
	n, err = s.loc.Read(p)
	s.Up += int64(n)
	go func() {
		if !s.Closed {
			s.chRead <- s.Up
		}
	}()
	return
}

func (s *Session) Write(p []byte) (n int, err error) {
	n, err = s.loc.Write(p)
	s.Down += int64(n)
	go func() {
		if !s.Closed {
			s.chWrite <- s.Down
		}
	}()
	return
}

type SessionMap struct {
	mtx   sync.Mutex
	group map[string]*Session
}

func NewSessionGroup() *SessionMap {
	return &SessionMap{group: map[string]*Session{}}
}

func (sg *SessionMap) AddSession(id string, session *Session) {
	sg.mtx.Lock()
	defer sg.mtx.Unlock()
	sg.group[id] = session
}

func (sg *SessionMap) DelSession(id string) {
	sg.mtx.Lock()
	defer sg.mtx.Unlock()
	delete(sg.group, id)
}

func (sg *SessionMap) Group() map[string]*Session {
	return sg.group
}
