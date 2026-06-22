package transporter

import (
	"testing"
	"time"

	"dxkite.cn/mino/config"
	"dxkite.cn/mino/stream"
)

func TestBasicAuth(t *testing.T) {
	today := time.Now().Format(expireAtLayout)
	tomorrow := time.Now().AddDate(0, 0, 1).Format(expireAtLayout)
	yesterday := time.Now().AddDate(0, 0, -1).Format(expireAtLayout)

	tests := []struct {
		name     string
		cfg      *config.Config
		info     *stream.AuthInfo
		expected bool
	}{
		{
			name:     "empty config allows anonymous access",
			cfg:      &config.Config{},
			info:     &stream.AuthInfo{},
			expected: true,
		},
		{
			name: "valid credentials without expiration",
			cfg: &config.Config{
				Username: "alice",
				Password: "secret",
			},
			info: &stream.AuthInfo{
				Username: "alice",
				Password: "secret",
			},
			expected: true,
		},
		{
			name: "invalid credentials",
			cfg: &config.Config{
				Username: "alice",
				Password: "secret",
			},
			info: &stream.AuthInfo{
				Username: "alice",
				Password: "wrong",
			},
			expected: false,
		},
		{
			name: "future expiration allows access",
			cfg: &config.Config{
				Username: "alice",
				Password: "secret",
				ExpireAt: tomorrow,
			},
			info: &stream.AuthInfo{
				Username: "alice",
				Password: "secret",
			},
			expected: true,
		},
		{
			name: "today expiration is expired after midnight",
			cfg: &config.Config{
				Username: "alice",
				Password: "secret",
				ExpireAt: today,
			},
			info: &stream.AuthInfo{
				Username: "alice",
				Password: "secret",
			},
			expected: false,
		},
		{
			name: "past expiration denies access",
			cfg: &config.Config{
				Username: "alice",
				Password: "secret",
				ExpireAt: yesterday,
			},
			info: &stream.AuthInfo{
				Username: "alice",
				Password: "secret",
			},
			expected: false,
		},
		{
			name: "invalid expiration denies access",
			cfg: &config.Config{
				Username: "alice",
				Password: "secret",
				ExpireAt: "2026/05/08",
			},
			info: &stream.AuthInfo{
				Username: "alice",
				Password: "secret",
			},
			expected: false,
		},
		{
			name: "valid user from users list",
			cfg: &config.Config{
				Users: []config.User{
					{
						Username: "alice",
						Password: "secret",
						ExpireAt: tomorrow,
					},
					{
						Username: "bob",
						Password: "hunter2",
						ExpireAt: tomorrow,
					},
				},
			},
			info: &stream.AuthInfo{
				Username: "bob",
				Password: "hunter2",
			},
			expected: true,
		},
		{
			name: "expired user from users list denies access",
			cfg: &config.Config{
				Users: []config.User{
					{
						Username: "alice",
						Password: "secret",
						ExpireAt: yesterday,
					},
					{
						Username: "bob",
						Password: "hunter2",
						ExpireAt: tomorrow,
					},
				},
			},
			info: &stream.AuthInfo{
				Username: "alice",
				Password: "secret",
			},
			expected: false,
		},
		{
			name: "top-level user remains compatible with users list",
			cfg: &config.Config{
				Username: "legacy",
				Password: "secret",
				ExpireAt: tomorrow,
				Users: []config.User{
					{
						Username: "alice",
						Password: "secret",
						ExpireAt: tomorrow,
					},
				},
			},
			info: &stream.AuthInfo{
				Username: "legacy",
				Password: "secret",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tsp := New(tt.cfg)
			if got := tsp.basicAuth(tt.info); got != tt.expected {
				t.Fatalf("basicAuth() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAuthRequired(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		expected bool
	}{
		{
			name:     "empty config does not require auth",
			cfg:      &config.Config{},
			expected: false,
		},
		{
			name: "top-level user requires auth",
			cfg: &config.Config{
				Username: "alice",
				Password: "secret",
			},
			expected: true,
		},
		{
			name: "users list requires auth",
			cfg: &config.Config{
				Users: []config.User{
					{Username: "alice", Password: "secret"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tsp := New(tt.cfg)
			if got := tsp.authRequired(); got != tt.expected {
				t.Fatalf("authRequired() = %v, want %v", got, tt.expected)
			}
		})
	}
}
