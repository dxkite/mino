package proxy

import (
	"context"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
	"time"
)

func init() {
	proxy.RegisterDialerType("mino", MinoDialer)
}

func MinoDialer(url *url.URL, dialer proxy.Dialer) (proxy.Dialer, error) {
	proxyAddr := url.Host
	user := url.User.Username()
	password, _ := url.User.Password()
	return &Dialer{
		ProxyDial:    dialer.Dial,
		Username:     user,
		Password:     password,
		ProxyAddress: proxyAddr,
	}, nil
}

func Test(remote string, proxyURL *url.URL, timeout time.Duration) error {
	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return err
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return DialContext(ctx, dialer, network, addr)
		},
	}
	client := &http.Client{Transport: transport}
	client.Timeout = timeout
	req, err := http.NewRequest("GET", remote, nil)
	if err != nil {
		return err
	}
	_, err = client.Do(req)
	return err
}

func DialContext(ctx context.Context, d proxy.Dialer, network, address string) (conn net.Conn, err error) {
	done := make(chan struct{}, 1)
	go func() {
		conn, err = d.Dial(network, address)
		close(done)
		if conn != nil && ctx.Err() != nil {
			_ = conn.Close()
		}
	}()
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-done:
	}
	return conn, err
}
