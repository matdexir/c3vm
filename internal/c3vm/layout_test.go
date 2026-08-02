package c3vm

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPopulateVersionDirLayout verifies the on-disk layout of an installed
// version matches what the c3c compiler expects: the c3c binary at the top
// level and the std lib under <version>/lib (c3c resolves lib/std relative
// to its installed directory).
func TestPopulateVersionDirLayout(t *testing.T) {
	extractDir := t.TempDir()
	destDir := t.TempDir()

	archiveBin := filepath.Join(extractDir, "bin")
	if err := os.MkdirAll(archiveBin, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(archiveBin, "c3c"), []byte("binary"), 0755); err != nil {
		t.Fatal(err)
	}

	stdDir := filepath.Join(extractDir, "lib", "std")
	if err := os.MkdirAll(stdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stdDir, "std.c3i"), []byte("stdlib"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := populateVersionDir(destDir, extractDir); err != nil {
		t.Fatalf("populateVersionDir returned error: %v", err)
	}

	// The binary must be at <version>/c3c.
	binaryPath := filepath.Join(destDir, "c3c")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Errorf("expected c3c binary at %s", binaryPath)
	}

	// The std lib must be at <version>/lib/std (c3c's expectation).
	stdLibPath := filepath.Join(destDir, "lib", "std", "std.c3i")
	if _, err := os.Stat(stdLibPath); os.IsNotExist(err) {
		t.Errorf("expected std lib at %s, want <version>/lib/std layout", stdLibPath)
	}

	// The std lib must NOT be under <version>/bin.
	wrongPath := filepath.Join(destDir, "bin", "std", "std.c3i")
	if _, err := os.Stat(wrongPath); err == nil {
		t.Errorf("std lib found at %s, but <version>/bin layout is unsupported by c3c", wrongPath)
	}
}

// TestPopulateVersionDirMissingFiles verifies the install fails cleanly when
// the archive is missing the binary or the std lib.
func TestPopulateVersionDirMissingFiles(t *testing.T) {
	extractDir := t.TempDir()
	destDir := t.TempDir()

	if err := populateVersionDir(destDir, extractDir); err == nil {
		t.Error("expected error when archive has no c3c binary or std lib")
	}

	// Only binary, no std lib.
	binDir := filepath.Join(extractDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "c3c"), []byte("binary"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := populateVersionDir(destDir, extractDir); err == nil {
		t.Error("expected error when archive is missing std lib")
	}
}

// TestUseLayout verifies the symlinks created by Use() point to the correct
// locations: ~/.c3vm/lib must point at <version>/lib, not <version>/bin.
func TestUseLayout(t *testing.T) {
	home := t.TempDir()
	v := &C3VM{HomeDir: home}

	versionDir := v.VersionDir("v0.8.0")
	if err := os.MkdirAll(filepath.Join(versionDir, "lib"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(versionDir, "c3c"), []byte("binary"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(versionDir, "lib", "std"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := v.Use("v0.8.0"); err != nil {
		t.Fatalf("Use returned error: %v", err)
	}

	// The lib symlink must resolve to <version>/lib.
	libTarget, err := os.Readlink(v.c3cStdLib())
	if err != nil {
		t.Fatalf("failed to read lib symlink: %v", err)
	}
	wantLibTarget := filepath.Join(versionDir, "lib")
	if libTarget != wantLibTarget {
		t.Errorf("lib symlink target = %q, want %q", libTarget, wantLibTarget)
	}

	// The lib symlink must expose std so that c3c's lib/std resolution works.
	if _, err := os.Stat(filepath.Join(v.c3cStdLib(), "std")); os.IsNotExist(err) {
		t.Errorf("expected std to be reachable via %s", filepath.Join(v.c3cStdLib(), "std"))
	}

	// The current symlink must point to the version dir.
	currentTarget, err := os.Readlink(v.currentLink())
	if err != nil {
		t.Fatalf("failed to read current symlink: %v", err)
	}
	if currentTarget != versionDir {
		t.Errorf("current symlink target = %q, want %q", currentTarget, versionDir)
	}
}
