package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRepositoryContainsNoLegacyUserFacingBrand(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	legacy := "dev" + "doctor"
	modulePath := "github.com/mahmudddddd/" + "Dev" + "Doctor"

	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "bin":
				return filepath.SkipDir
			}
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if !utf8.Valid(data) {
			return nil
		}
		content := strings.ReplaceAll(string(data), modulePath, "")
		if strings.Contains(strings.ToLower(content), legacy) {
			relative, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			t.Errorf("legacy user-facing brand remains in %s", relative)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadmeUsesDebugDocCommandExamples(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	readme := string(content)
	for _, expected := range []string{
		"# DebugDoc",
		"debugdoc diagnose --path ./my-project",
		"debugdoc diagnose --path ./my-project --format json",
		"debugdoc version",
		"cmd/debugdoc",
	} {
		if !strings.Contains(readme, expected) {
			t.Errorf("README is missing %q", expected)
		}
	}
}
