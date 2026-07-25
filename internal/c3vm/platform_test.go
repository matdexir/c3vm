package c3vm

import (
	"testing"
)

func TestAssetName(t *testing.T) {
	tests := []struct {
		platform Platform
		want     string
	}{
		{Platform{"macos", "aarch64"}, "c3-macos"},
		{Platform{"macos", "x86_64"}, "c3-macos"},
		{Platform{"linux", "x86_64"}, "c3-linux"},
		{Platform{"linux", "aarch64"}, "c3-linux"},
		{Platform{"android", "aarch64"}, "c3-android-aarch64"},
		{Platform{"netbsd", "x86_64"}, "c3-netbsd"},
		{Platform{"openbsd", "x86_64"}, "c3-openbsd"},
		{Platform{"freebsd", "x86_64"}, "c3-freebsd-x86_64"},
	}

	for _, tt := range tests {
		got := tt.platform.AssetName()
		if got != tt.want {
			t.Errorf("AssetName(%+v) = %q, want %q", tt.platform, got, tt.want)
		}
	}
}

func TestArchiveExt(t *testing.T) {
	tests := []struct {
		platform Platform
		want     string
	}{
		{Platform{"macos", "aarch64"}, ".zip"},
		{Platform{"macos", "x86_64"}, ".zip"},
		{Platform{"linux", "x86_64"}, ".tar.gz"},
		{Platform{"android", "aarch64"}, ".tar.gz"},
		{Platform{"netbsd", "x86_64"}, ".tar.gz"},
		{Platform{"freebsd", "x86_64"}, ".tar.gz"},
	}

	for _, tt := range tests {
		got := tt.platform.ArchiveExt()
		if got != tt.want {
			t.Errorf("ArchiveExt(%+v) = %q, want %q", tt.platform, got, tt.want)
		}
	}
}

func TestDetect(t *testing.T) {
	p := Detect()
	if p.OS == "" {
		t.Error("Detect() returned empty OS")
	}
	if p.Arch == "" {
		t.Error("Detect() returned empty Arch")
	}
}
