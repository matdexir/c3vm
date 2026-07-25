package c3vm

import (
	"fmt"
	"log/slog"
	"runtime"
)

type Platform struct {
	OS   string
	Arch string
}

func Detect() Platform {
	slog.Debug("detecting platform", "GOOS", runtime.GOOS, "GOARCH", runtime.GOARCH)
	osMap := map[string]string{
		"darwin":  "macos",
		"linux":   "linux",
		"netbsd":  "netbsd",
		"openbsd": "openbsd",
		"android": "android",
	}
	archMap := map[string]string{
		"amd64": "x86_64",
		"arm64": "aarch64",
	}
	os := osMap[runtime.GOOS]
	if os == "" {
		os = runtime.GOOS
	}
	arch := archMap[runtime.GOARCH]
	if arch == "" {
		arch = runtime.GOARCH
	}
	p := Platform{OS: os, Arch: arch}
	slog.Debug("platform detected", "os", p.OS, "arch", p.Arch)
	return p
}

func (p Platform) AssetName() string {
	switch p.OS {
	case "macos":
		return fmt.Sprintf("c3-macos")
	case "linux":
		return fmt.Sprintf("c3-linux")
	case "android":
		return fmt.Sprintf("c3-android-%s", p.Arch)
	case "netbsd":
		return fmt.Sprintf("c3-netbsd")
	case "openbsd":
		return fmt.Sprintf("c3-openbsd")
	default:
		return fmt.Sprintf("c3-%s-%s", p.OS, p.Arch)
	}
}

func (p Platform) ArchiveExt() string {
	if p.OS == "macos" {
		return ".zip"
	}
	return ".tar.gz"
}
