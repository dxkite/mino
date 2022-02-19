package proxy

import (
	"fmt"
	"net/url"
	"testing"
	"time"
)

func TestUrl(t *testing.T) {
	pu, err := url.Parse("mino://127.0.0.1:21080")
	if err != nil {
		fmt.Println("parse", err)
	}
	err = Test("https://google.com", pu, 3*time.Second)
	fmt.Println("test", err)
}
