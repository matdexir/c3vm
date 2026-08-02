package c3vm

import (
	"fmt"
	"io"
	"log/slog"
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

func (v *C3VM) c3cStdLib() string {
	return filepath.Join(v.HomeDir, "lib")
}

func (v *C3VM) VersionDir(version string) string {
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

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath)
	})
}

func populateVersionDir(destDir, extractDir string) error {
	// Find c3c binary and std lib in extracted files
	var c3cSrc string
	var c3cStdLib string
	filepath.WalkDir(extractDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "c3c" {
			c3cSrc = path
		}

		if d.IsDir() && d.Name() == "lib" {
			c3cStdLib = path
		}
		return nil
	})

	slog.Debug("located files in archive", "binary", c3cSrc, "stdlib", c3cStdLib)

	if c3cSrc == "" {
		return fmt.Errorf("could not find c3c binary in the archive")
	}

	if c3cStdLib == "" {
		return fmt.Errorf("could not find c3c standard library in the archive")
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create version directory: %w", err)
	}
	slog.Debug("created version directory", "path", destDir)

	// Copy c3c binary to version dir
	destC3C := filepath.Join(destDir, "c3c")

	slog.Debug("copying binary", "src", c3cSrc, "dest", destC3C)
	if err := copyFile(c3cSrc, destC3C); err != nil {
		return fmt.Errorf("failed to copy c3c binary: %w", err)
	}

	// Copy c3c std lib to version dir. c3c resolves its standard library
	// as lib/std relative to its installed directory (the parent of the
	// c3c binary), so the std lib must live under <version>/lib.
	destC3CStdLib := filepath.Join(destDir, "lib")
	slog.Debug("copying standard library", "src", c3cStdLib, "dest", destC3CStdLib)
	if err := copyDir(c3cStdLib, destC3CStdLib); err != nil {
		return fmt.Errorf("failed to copy c3c standard library: %w", err)
	}

	return nil
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	if !strings.HasPrefix(version, "v") {
		return "v" + version
	}
	return version
}

func (v *C3VM) ResolveVersion(version string) (string, error) {	if version == "latest" {
		release, err := github.GetLatestRelease()
		if err != nil {
			return "", err
		}
		return release.TagName, nil
	}
	return normalizeVersion(version), nil
}

func (v *C3VM) Install(version string) error {
	tag, err := v.ResolveVersion(version)
	if err != nil {
		return err
	}

	slog.Debug("resolved version", "input", version, "tag", tag)

	destDir := v.VersionDir(tag)
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("version %s is already installed", tag)
	}

	slog.Debug("version not installed, starting download", "tag", tag, "dest", destDir)

	plat := Detect()
	slog.Debug("detected platform", "os", plat.OS, "arch", plat.Arch, "asset", plat.AssetName())

	assetName := plat.AssetName()
	ext := plat.ArchiveExt()
	archiveName := assetName + ext

	url := github.DownloadURL(tag, archiveName)

	slog.Info("downloading", "url", url)
	tmpDir, err := os.MkdirTemp("", "c3vm-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() {
		slog.Debug("cleaning up temp directory", "path", tmpDir)
		os.RemoveAll(tmpDir)
	}()

	archivePath := filepath.Join(tmpDir, archiveName)
	slog.Debug("downloading archive", "src", url, "dest", archivePath)
	if err := downloadFile(url, archivePath); err != nil {
		return err
	}
	slog.Debug("download complete", "size", fileSize(archivePath))

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return err
	}

	slog.Info("extracting", "archive", archivePath)
	if err := extractArchive(archivePath, extractDir); err != nil {
		return err
	}
	slog.Debug("extraction complete", "extractDir", extractDir)

	if err := populateVersionDir(destDir, extractDir); err != nil {
		return err
	}

	slog.Info("installed", "version", tag)

	// Auto-use if no current version is set
	current, _ := os.Readlink(v.currentLink())
	if current == "" {
		slog.Debug("no active version set, auto-switching", "version", tag)
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
	version = normalizeVersion(version)
	dir := v.VersionDir(version)
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
	tag, err := v.ResolveVersion(version)
	if err != nil {
		return err
	}
	slog.Debug("switching version", "input", version, "tag", tag)

	dir := v.VersionDir(tag)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("version %s is not installed", tag)
	}

	if err := v.ensureDirs(); err != nil {
		return err
	}

	link := v.currentLink()
	slog.Debug("removing existing symlink", "path", link)
	os.Remove(link)

	slog.Debug("creating current symlink", "target", dir, "link", link)
	if err := os.Symlink(dir, link); err != nil {
		return fmt.Errorf("failed to set current symlink: %w", err)
	}

	binC3C := v.c3cBin()
	slog.Debug("removing existing bin/c3c symlink", "path", binC3C)
	os.Remove(binC3C)

	stdLibC3C := v.c3cStdLib()
	slog.Debug("removing existing lib symlink", "path", stdLibC3C)
	os.RemoveAll(stdLibC3C)

	slog.Debug("creating bin/c3c symlink", "target", filepath.Join(link, "c3c"), "link", binC3C)
	if err := os.Symlink(filepath.Join(link, "c3c"), binC3C); err != nil {
		return fmt.Errorf("failed to create bin/c3c symlink: %w", err)
	}

	slog.Debug("creating lib symlink", "target", filepath.Join(dir, "lib"), "link", stdLibC3C)
	if err := os.Symlink(filepath.Join(dir, "lib"), stdLibC3C); err != nil {
		return fmt.Errorf("failed to create standard library symlink: %w", err)
	}

	slog.Info("now using", "version", tag)
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
	version = normalizeVersion(version)
	dir := v.VersionDir(version)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("version %s is not installed", version)
	}
	slog.Debug("setting default version", "version", version, "file", v.defaultFile())
	if err := os.WriteFile(v.defaultFile(), []byte(version+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write default version: %w", err)
	}
	slog.Info("default version set", "version", version)
	return nil
}

func (v *C3VM) BinInPath() bool {
	path := os.Getenv("PATH")
	bin := v.binDir()
	for _, p := range strings.Split(path, ":") {
		if p == bin {
			return true
		}
	}
	return false
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

func fileSize(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%d bytes", info.Size())
}

func (v *C3VM) Which(version string) (string, error) {
	if version == "" {
		current, err := v.Current()
		if err != nil || current == "" {
			return "", fmt.Errorf("no active version set")
		}
		version = current
	}

	version = normalizeVersion(version)
	dir := v.VersionDir(version)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", fmt.Errorf("version %s is not installed", version)
	}

	c3cPath := filepath.Join(dir, "c3c")
	if _, err := os.Stat(c3cPath); os.IsNotExist(err) {
		return "", fmt.Errorf("c3c binary not found in version %s", version)
	}

	return c3cPath, nil
}
