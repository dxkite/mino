package proto

import (
	"dxkite.cn/go-gateway/lib/rewind"
	"errors"
)

var ErrUnknownProto = errors.New("unknown proto error")

type ProtoManager struct {
	proto []Proto
	ident []Identifier
}

// 添加协议
func (ic *ProtoManager) Add(proto Proto) {
	ic.proto = append(ic.proto, proto)
	ic.ident = append(ic.ident, proto.NewIdentifier())
}

// 判断协议类型
func (ic *ProtoManager) Proto(conn rewind.Conn) (proto Proto, err error) {
	for i := range ic.ident {
		if err = conn.Rewind(); err != nil {
			return nil, err
		}
		ok, er := ic.ident[i].Check(conn)
		if er != nil {
			return nil, er
		}
		if ok {
			return ic.proto[i], nil
		}
	}
	return nil, ErrUnknownProto
}
