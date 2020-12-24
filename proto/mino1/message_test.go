package mino1

import (
	"net"
	"reflect"
	"testing"
)

func TestResponseMessage_marshal(t *testing.T) {
	tests := []struct {
		name    string
		Code    uint8
		Message string
		want    []byte
	}{
		{
			"simple",
			0,
			"OK",
			[]byte{0, 2, 'O', 'K'},
		},
		{
			"simple-empty",
			10,
			"",
			[]byte{10, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ResponseMessage{
				Code:    tt.Code,
				Message: tt.Message,
			}
			if got := m.marshal(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponseMessage_unmarshal(t *testing.T) {

	tests := []struct {
		name    string
		Code    uint8
		Message string
		p       []byte
		wantErr bool
	}{
		{
			"simple",
			0,
			"OK",
			[]byte{0, 2, 'O', 'K'},
			false,
		},
		{
			"simple-empty",
			10,
			"",
			[]byte{10, 0},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ResponseMessage{
				Code:    tt.Code,
				Message: tt.Message,
			}
			v := new(ResponseMessage)
			if err := v.unmarshal(tt.p); (err != nil) != tt.wantErr {
				t.Errorf("unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(m, v) {
				t.Errorf("unmarshal() got = %v, want %v", v, m)
			}
		})
	}
}

func TestRequestMessage_unmarshal(t *testing.T) {
	tests := []struct {
		name       string
		Network    uint8
		Address    string
		Username   string
		Password   string
		MacAddress []net.HardwareAddr
		p          []byte
		wantErr    bool
	}{
		{
			"simple",
			uint8(NetworkUdp),
			"dxkite.cn:443",
			"dxkite",
			"p@ssw0rd",
			[]net.HardwareAddr{
				[]byte("\x00\x11\x22\x33\x44\x55"),
			},
			[]byte("\x01\x0Ddxkite.cn:443\x06dxkite\x08p@ssw0rd\x01\x00\x11\x22\x33\x44\x55"),
			false,
		},
		{
			"simple-empty-auth",
			uint8(NetworkUdp),
			"dxkite.cn:443",
			"",
			"",
			[]net.HardwareAddr{
				[]byte("\x00\x11\x22\x33\x44\x55"),
			},
			[]byte("\x01\x0Ddxkite.cn:443\x00\x00\x01\x00\x11\x22\x33\x44\x55"),
			false,
		},
		{
			"simple-empty-auth-address",
			uint8(NetworkUdp),
			"dxkite.cn:443",
			"",
			"",
			[]net.HardwareAddr{},
			[]byte("\x01\x0Ddxkite.cn:443\x00\x00\x00"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RequestMessage{
				Network:    tt.Network,
				Address:    tt.Address,
				Username:   tt.Username,
				Password:   tt.Password,
				MacAddress: tt.MacAddress,
			}
			v := new(RequestMessage)
			if err := v.unmarshal(tt.p); (err != nil) != tt.wantErr {
				t.Errorf("unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(m, v) {
				t.Errorf("unmarshal() got = %v, want %v", v, m)
			}

		})
	}
}
