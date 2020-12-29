package transport

import (
	"errors"
	"io"
	"sync"
)

// 建立隧道
type Tunnel struct {
	// 流量统计
	Flow Flow

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
	// 关闭通知
	closeCh chan struct{}
}

// 建立隧道
func NewTunnel(loc, rmt io.ReadWriteCloser) *Tunnel {
	return &Tunnel{
		loc:     NewFlow(loc),
		rmt:     rmt,
		closeCh: make(chan struct{}),
	}
}

// 传输数据
func (s *Tunnel) Transport() (up, down int64, err error) {
	var _closed = make(chan struct{})
	go func() {
		// remote -> local
		var _err error
		if down, _err = io.Copy(s.loc, s.rmt); _err != nil {
			s.rwErr("read", _err)
		}
		_closed <- struct{}{}
	}()
	// local -> remote
	var _err error
	if up, _err = io.Copy(s.rmt, s.loc); _err != nil {
		s.rwErr("write", _err)
	}
	_ = s.close()
	<-_closed
	err = s.err
	// 通知关闭
	s.notifyClose()
	return
}

func (s *Tunnel) notifyClose() {
	s.closeCh <- struct{}{}
}

func (s *Tunnel) CloseNotify() <-chan struct{} {
	return s.closeCh
}

func (s *Tunnel) close() error {
	if !s.closed {
		s.mtxClosed.Lock()
		s.closed = true
		s.mtxClosed.Unlock()
		_ = s.loc.Close()
		_ = s.rmt.Close()
	}
	return s.err
}

func (s *Tunnel) rwErr(name string, err error) {
	s.mtxErr.Lock()
	defer s.mtxErr.Unlock()
	if !s.closed && err != nil {
		s.err = errors.New("tunnel error:" + name + ":" + err.Error())
		_ = s.close()
	}
}

type Session struct {
	*Tunnel
	Id  int    `json:"sid"`
	Src string `json:"src"`
	Dst string `json:"dst"`
}

func NewSession(sid int, tun *Tunnel, src, dst string) *Session {
	return &Session{
		Tunnel: tun,
		Id:     sid,
		Src:    src,
		Dst:    dst,
	}
}

type Flow struct {
	rwc io.ReadWriteCloser
	R   int64 `json:"r"`
	W   int64 `json:"w"`
	C   bool  `json:"c"`
}

func NewFlow(rw io.ReadWriteCloser) *Flow {
	return &Flow{
		R:   0,
		W:   0,
		C:   false,
		rwc: rw,
	}
}

func (f *Flow) Read(p []byte) (n int, err error) {
	n, err = f.rwc.Read(p)
	f.R += int64(n)
	return
}

func (f *Flow) Write(p []byte) (n int, err error) {
	n, err = f.rwc.Write(p)
	f.W += int64(n)
	return
}

func (f *Flow) Close() error {
	f.C = true
	return f.rwc.Close()
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
	go func() {
		<-session.CloseNotify()
		delete(sg[group], id)
	}()
}
