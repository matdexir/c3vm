package c3vm

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/matdexir/c3vm/internal/github"
)

type C3VM struct {
	HomeDir string
}

func New() (*C3VM, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	c3vmHome := filepath.Join(home, ".c3vm")
	return &C3VM{HomeDir: c3vmHome}, nil
}

func (v *C3VM) versionsDir() string {
	return filepath.Join(v.HomeDir, "versions")
}

func (v *C3VM) binDir() string {
	return filepath.Join(v.HomeDir, "bin")
}

func (v *C3VM) currentLink() string {
	return filepath.Join(v.HomeDir, "current")
}

func (v *C3VM) defaultFile() string {
	return filepath.Join(v.HomeDir, "default")
}

func (v *C3VM) c3cBin() string {
	return filepath.Join(v.binDir(), "c3c")
}

func (v *C3VM) versionDir(version string) string {
	return filepath.Join(v.versionsDir(), version)
}

func (v *C3VM) ensureDirs() error {
	for _, d := range []string{v.versionsDir(), v.binDir()} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}
	return nil
}

func (v *C3VM) Install(version string) error {
	var tag string
	if version == "latest" {
		release, err := github.GetLatestRelease()
		if err != nil {
			return err
		}
		tag = release.TagName
	} else {
		tag = version
	}

	destDir := v.versionDir(tag)
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("version %s is already installed", tag)
	}

	plat := Detect()
	assetName := plat.AssetName()
	ext := plat.ArchiveExt()
	archiveName := assetName + ext

	url := github.DownloadURL(tag, archiveName)

	fmt.Printf("Downloading %s ...\n", url)
	tmpDir, err := os.MkdirTemp("", "c3vm-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, archiveName)
	if err := downloadFile(url, archivePath); err != nil {
		return err
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return err
	}

	fmt.Println("Extracting ...")
	if err := extractArchive(archivePath, extractDir); err != nil {
		return err
	}

	// Find c3c binary in extracted files
	var c3cSrc string
	filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == "c3c" {
			c3cSrc = path
		}
		if strings.HasSuffix(info.Name(), ".exe") && strings.HasPrefix(info.Name(), "c3c") {
			c3cSrc = path
		}
		return nil
	})

	if c3cSrc == "" {
		return fmt.Errorf("could not find c3c binary in the archive")
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create version directory: %w", err)
	}

	// Copy c3c binary to version dir
	destC3C := filepath.Join(destDir, "c3c")
	input, err := os.ReadFile(c3cSrc)
	if err != nil {
		return fmt.Errorf("failed to read c3c binary: %w", err)
	}
	if err := os.WriteFile(destC3C, input, 0755); err != nil {
		return fmt.Errorf("failed to write c3c binary: %w", err)
	}

	fmt.Printf("Installed c3c %s\n", tag)

	// Auto-use if no current version is set
	current, _ := os.Readlink(v.currentLink())
	if current == "" {
		if err := v.Use(tag); err != nil {
			return err
		}
	}

	if !v.BinInPath() {
		fmt.Println()
		fmt.Println("To use c3c from any directory, add ~/.c3vm/bin to your PATH:")
		fmt.Println()
		fmt.Printf("  export PATH=\"$HOME/.c3vm/bin:$PATH\"\n")
		fmt.Println()
		fmt.Println("Run 'c3vm init' to print shell-specific setup instructions.")
	}

	return nil
}

func (v *C3VM) Remove(version string) error {
	dir := v.versionDir(version)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("version %s is not installed", version)
	}

	// Check if it's the current version
	current, _ := v.Current()
	if current == version {
		return fmt.Errorf("cannot remove the active version; switch to another version first")
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove %s: %w", version, err)
	}
	return nil
}

func (v *C3VM) List() ([]string, error) {
	dir := v.versionsDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read versions directory: %w", err)
	}

	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	return versions, nil
}

func (v *C3VM) ListRemote() ([]string, error) {
	tags, err := github.ListAllTags()
	if err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(tags)))
	return tags, nil
}

func (v *C3VM) Current() (string, error) {
	link := v.currentLink()
	target, err := os.Readlink(link)
	if err != nil {
		return "", nil
	}
	return filepath.Base(target), nil
}

func (v *C3VM) Use(version string) error {
	dir := v.versionDir(version)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("version %s is not installed", version)
	}

	link := v.currentLink()

	// Remove existing symlink if any
	os.Remove(link)

	if err := os.Symlink(dir, link); err != nil {
		return fmt.Errorf("failed to set current symlink: %w", err)
	}

	if err := v.ensureDirs(); err != nil {
		return err
	}

	// Update bin/c3c symlink
	binC3C := v.c3cBin()
	os.Remove(binC3C)

	if err := os.Symlink(filepath.Join(link, "c3c"), binC3C); err != nil {
		return fmt.Errorf("failed to create bin/c3c symlink: %w", err)
	}

	fmt.Printf("Now using c3c %s\n", version)
	return nil
}

func (v *C3VM) Default() (string, error) {
	data, err := os.ReadFile(v.defaultFile())
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(data)), nil
}

func (v *C3VM) SetDefault(version string) error {
	dir := v.versionDir(version)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("version %s is not installed", version)
	}
	if err := os.WriteFile(v.defaultFile(), []byte(version+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write default version: %w", err)
	}
	fmt.Printf("Default version set to %s\n", version)
	return nil
}

func (v *C3VM) BinInPath() bool {
	return strings.Contains(os.Getenv("PATH"), v.binDir())
}

func (v *C3VM) InitScript() string {
	shell := os.Getenv("SHELL")
	switch {
	case strings.Contains(shell, "zsh"):
		return `echo 'export PATH="$HOME/.c3vm/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc`
	case strings.Contains(shell, "bash"):
		return `echo 'export PATH="$HOME/.c3vm/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc`
	case strings.Contains(shell, "fish"):
		return `echo 'set -gx PATH "$HOME/.c3vm/bin" $PATH' >> ~/.config/fish/config.fish && source ~/.config/fish/config.fish`
	default:
		return `export PATH="$HOME/.c3vm/bin:$PATH"`
	}
}

func (v *C3VM) Which(version string) (string, error) {
	if version == "" {
		current, err := v.Current()
		if err != nil || current == "" {
			return "", fmt.Errorf("no active version set")
		}
		version = current
	}

	dir := v.versionDir(version)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", fmt.Errorf("version %s is not installed", version)
	}

	c3cPath := filepath.Join(dir, "c3c")
	if _, err := os.Stat(c3cPath); os.IsNotExist(err) {
		return "", fmt.Errorf("c3c binary not found in version %s", version)
	}

	return c3cPath, nil
}
