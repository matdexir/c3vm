package github

import (
	"testing"
)

func TestDownloadURL(t *testing.T) {
	tests := []struct {
		tag       string
		assetName string
		want      string
	}{
		{"v0.8.2", "c3-macos.zip", "https://github.com/c3lang/c3c/releases/download/v0.8.2/c3-macos.zip"},
		{"v0.8.1", "c3-linux.tar.gz", "https://github.com/c3lang/c3c/releases/download/v0.8.1/c3-linux.tar.gz"},
		{"latest", "c3-macos.zip", "https://github.com/c3lang/c3c/releases/download/latest/c3-macos.zip"},
	}

	for _, tt := range tests {
		got := DownloadURL(tt.tag, tt.assetName)
		if got != tt.want {
			t.Errorf("DownloadURL(%q, %q) = %q, want %q", tt.tag, tt.assetName, got, tt.want)
		}
	}
}
