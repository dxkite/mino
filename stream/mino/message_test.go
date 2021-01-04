package mino

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestRequestMessage_marshal(t *testing.T) {
	tests := []struct {
		name    string
		m       *RequestMessage
		want    []byte
		wantErr bool
	}{
		{
			"simple-ipv4",
			&RequestMessage{
				Network: "tcp",
				Address: "127.0.0.1:1080",
			},
			[]byte("\x02" + string(byte(0b00000000)) + "\x7f\x00\x00\x01\x04\x38"),
			false,
		},
		{
			"simple-ipv6",
			&RequestMessage{
				Network: "udp",
				Address: "[::1]:1080",
			},
			[]byte("\x02" + string(byte(0b00110000)) + "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x04\x38"),
			false,
		},
		{
			"simple-host-password",
			&RequestMessage{
				Network:  "udp",
				Address:  "baidu.com:443",
				Username: "dxkite",
				Password: "123456",
			},
			[]byte("\x02" + string(byte(0b01010001)) + "\x09baidu.com\x01\xbb\x06\x06dxkite123456"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.marshal()
			if (err != nil) != tt.wantErr {
				t.Errorf("marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("marshal() got = %08b %x, want %08b %x", got[0], got[1:], tt.want[0], tt.want[1:])
			}
		})
	}
}

func TestRequestMessage_unmarshal(t *testing.T) {

	tests := []struct {
		name    string
		r       io.Reader
		m       *RequestMessage
		wantErr bool
	}{
		{
			"simple-ipv4",
			bytes.NewBufferString(string(byte(0b00000000)) + "\x7f\x00\x00\x01\x04\x38"),
			&RequestMessage{
				Network: "tcp",
				Address: "127.0.0.1:1080",
			},
			false,
		},
		{
			"simple-ipv6",
			bytes.NewBufferString(string(byte(0b00110000)) + "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x04\x38"),
			&RequestMessage{
				Network: "udp",
				Address: "[::1]:1080",
			},
			false,
		},
		{
			"simple-host-password",
			bytes.NewBufferString(string(byte(0b01010001)) + "\x09baidu.com\x01\xbb\x06\x06dxkite123456"),
			&RequestMessage{
				Network:  "udp",
				Address:  "baidu.com:443",
				Username: "dxkite",
				Password: "123456",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := new(RequestMessage)
			if err := mm.unmarshal(tt.r); (err != nil) != tt.wantErr {
				t.Errorf("unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if reflect.DeepEqual(mm, tt.m) == false {
				t.Errorf("unmarshal() got = %v, want %v", mm, tt.m)
			}
		})
	}
}
