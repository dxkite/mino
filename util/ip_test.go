package util

import "testing"

func TestIsLoopback(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"127.0.0.1", true},
		{"::1", true},
		{"only.dxkite.dx", true},
		{"baidu.com", false},
		{"xxxxx.xxxxxx-baidu.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := IsLoopback(tt.host); got != tt.want {
				t.Errorf("IsLoopback() = %v, want %v", got, tt.want)
			}
		})
	}
}
