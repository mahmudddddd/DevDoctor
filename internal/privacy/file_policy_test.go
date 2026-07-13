package privacy

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReadMetadataAllowsPackageJSON(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	want := []byte(`{"name":"safe"}`)
	if err := os.WriteFile(filepath.Join(root, "package.json"), want, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := NewFilePolicy().ReadMetadata(root, "package.json")
	if err != nil {
		t.Fatalf("ReadMetadata() error = %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("ReadMetadata() = %q, want %q", got, want)
	}
}

func TestResolveMetadataPathDeniesSensitiveAndEscapingPaths(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	policy := NewFilePolicy()

	for _, path := range []string{".env", "credentials.json", "../package.json", filepath.Join("nested", "package.json")} {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			_, err := policy.ResolveMetadataPath(root, path)
			if !errors.Is(err, ErrFileDenied) {
				t.Fatalf("ResolveMetadataPath(%q) error = %v, want ErrFileDenied", path, err)
			}
		})
	}
}

func TestReadMetadataEnforcesSizeLimit(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(strings.Repeat("x", 17)), 0o600); err != nil {
		t.Fatal(err)
	}

	policy := FilePolicy{MaxBytes: 16}
	_, err := policy.ReadMetadata(root, "package.json")
	if !errors.Is(err, ErrFileDenied) {
		t.Fatalf("ReadMetadata() error = %v, want ErrFileDenied", err)
	}
}

func TestMetadataNameCaseMatchesFilesystemSemantics(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "PACKAGE.JSON"), []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := NewFilePolicy().ReadMetadata(root, "PACKAGE.JSON")
	if runtime.GOOS == "windows" && err != nil {
		t.Fatalf("ReadMetadata() error = %v on case-insensitive Windows filesystem", err)
	}
	if runtime.GOOS != "windows" && !errors.Is(err, ErrFileDenied) {
		t.Fatalf("ReadMetadata() error = %v, want ErrFileDenied on case-sensitive platform", err)
	}
}

func TestResolveMetadataPathDeniesSymlinkEscape(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.json")
	if err := os.WriteFile(outside, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "package.json")); err != nil {
		t.Skipf("symbolic links are unavailable in this environment: %v", err)
	}

	_, err := NewFilePolicy().ResolveMetadataPath(root, "package.json")
	if !errors.Is(err, ErrFileDenied) {
		t.Fatalf("ResolveMetadataPath() error = %v, want ErrFileDenied", err)
	}
}
