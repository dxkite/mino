package transport

type Handler interface {
	Conn(sessionId int)
}
