package c3vm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathHelpers(t *testing.T) {
	v := &C3VM{HomeDir: "/home/user/.c3vm"}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"versionsDir", v.versionsDir(), "/home/user/.c3vm/versions"},
		{"binDir", v.binDir(), "/home/user/.c3vm/bin"},
		{"currentLink", v.currentLink(), "/home/user/.c3vm/current"},
		{"defaultFile", v.defaultFile(), "/home/user/.c3vm/default"},
		{"c3cBin", v.c3cBin(), "/home/user/.c3vm/bin/c3c"},
		{"c3cStdLib", v.c3cStdLib(), "/home/user/.c3vm/lib"},
		{"VersionDir(v0.8.0)", v.VersionDir("v0.8.0"), "/home/user/.c3vm/versions/v0.8.0"},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestResolveVersion(t *testing.T) {
	v := &C3VM{}

	tag, err := v.ResolveVersion("v0.8.0")
	if err != nil {
		t.Fatalf("ResolveVersion(v0.8.0) returned error: %v", err)
	}
	if tag != "v0.8.0" {
		t.Errorf("ResolveVersion(v0.8.0) = %q, want %q", tag, "v0.8.0")
	}
}

func TestBinInPath(t *testing.T) {
	v := &C3VM{HomeDir: "/home/user/.c3vm"}

	os.Unsetenv("PATH")
	if v.BinInPath() {
		t.Error("BinInPath() = true with empty PATH")
	}

	os.Setenv("PATH", "/usr/bin:/bin")
	if v.BinInPath() {
		t.Error("BinInPath() = true when .c3vm/bin is not in PATH")
	}

	os.Setenv("PATH", "/home/user/.c3vm/bin:/usr/bin")
	if !v.BinInPath() {
		t.Error("BinInPath() = false when .c3vm/bin is in PATH")
	}

	os.Setenv("PATH", "/usr/bin:/home/user/.c3vm/bin")
	if !v.BinInPath() {
		t.Error("BinInPath() = false when .c3vm/bin is at end of PATH")
	}
}

func TestInitScript(t *testing.T) {
	v := &C3VM{}

	os.Unsetenv("SHELL")
	got := v.InitScript()
	want := `export PATH="$HOME/.c3vm/bin:$PATH"`
	if got != want {
		t.Errorf("InitScript() with no SHELL = %q, want %q", got, want)
	}

	os.Setenv("SHELL", "/bin/zsh")
	got = v.InitScript()
	if got != `echo 'export PATH="$HOME/.c3vm/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc` {
		t.Errorf("InitScript() with zsh returned unexpected value: %q", got)
	}

	os.Setenv("SHELL", "/bin/bash")
	got = v.InitScript()
	if got != `echo 'export PATH="$HOME/.c3vm/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc` {
		t.Errorf("InitScript() with bash returned unexpected value: %q", got)
	}

	os.Setenv("SHELL", "/usr/bin/fish")
	got = v.InitScript()
	want = `echo 'set -gx PATH "$HOME/.c3vm/bin" $PATH' >> ~/.config/fish/config.fish && source ~/.config/fish/config.fish`
	if got != want {
		t.Errorf("InitScript() with fish = %q, want %q", got, want)
	}
}

func TestFileSize(t *testing.T) {
	got := fileSize("/nonexistent/path")
	if got != "unknown" {
		t.Errorf("fileSize(nonexistent) = %q, want %q", got, "unknown")
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	got = fileSize(file)
	want := "5 bytes"
	if got != want {
		t.Errorf("fileSize(test.txt) = %q, want %q", got, want)
	}
}
