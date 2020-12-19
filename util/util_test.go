package util

import (
	"fmt"
	"testing"
)

func TestUnzip(t *testing.T) {
	fmt.Println(Unzip("testdata/mino-windows.zip", "testdata", "testdata/backup"))
}
