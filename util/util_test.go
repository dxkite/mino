package util

import (
	"fmt"
	"testing"
)

func TestUnzip(t *testing.T) {
	fmt.Println(Unzip("testdata/mino-windows.zip", "testdata", "testdata/backup"))
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		ver1 string
		ver2 string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.1.0", "1.0.0", 1},
		{"1.1.2", "1.1.0", 2},
		{"1.1.2.1000", "1.1.2.100", 900},
		{"1.1.2.100-alpha", "1.1.2.100-beta", -1},
		{"1.1.2.100-beta", "1.1.2.100-alpha", 1},
		{"1.1.2.100-release", "1.1.2.100-beta", 1},
		{"1.1.2.100", "1.1.2.100-beta", 1},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.ver1, tt.ver2), func(t *testing.T) {
			if got := VersionCompare(tt.ver1, tt.ver2); got != tt.want {
				t.Errorf("VersionCompare() = %v, want %v", got, tt.want)
			}
		})
	}
}
