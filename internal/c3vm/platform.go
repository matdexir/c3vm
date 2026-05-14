package c3vm

import (
	"fmt"
	"runtime"
)

type Platform struct {
	OS   string
	Arch string
}

func Detect() Platform {
	osMap := map[string]string{
		"darwin":  "macos",
		"linux":   "linux",
		"windows": "windows",
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
	return Platform{OS: os, Arch: arch}
}

func (p Platform) AssetName() string {
	switch p.OS {
	case "macos":
		return fmt.Sprintf("c3-macos")
	case "windows":
		return fmt.Sprintf("c3-windows")
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
	if p.OS == "macos" || p.OS == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}
