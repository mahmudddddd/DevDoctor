package detect

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mahmudddddd/DevDoctor/internal/model"
	"github.com/mahmudddddd/DevDoctor/internal/privacy"
)

func TestDiscoverNodeNPMFixture(t *testing.T) {
	t.Parallel()
	detector := NewProjectDetector(privacy.NewFilePolicy())
	root := filepath.Join("..", "..", "tests", "fixtures", "node-npm")

	summary, err := detector.Discover(context.Background(), root)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	if summary.Name != "fixture-node-npm" {
		t.Fatalf("Name = %q, want fixture-node-npm", summary.Name)
	}
	assertDetection(t, summary.Languages, "javascript")
	assertDetection(t, summary.Languages, "typescript")
	assertDetection(t, summary.Frameworks, "nextjs")
	assertDetection(t, summary.Frameworks, "react")
	assertRuntime(t, summary.Runtimes, "nodejs", ">=22")
	assertPackageManager(t, summary.PackageManagers, "npm", "11.0.0")
	assertWorkspace(t, summary.Workspaces, "package-workspaces", []string{"packages/*"})
	if len(summary.Warnings) != 0 {
		t.Fatalf("Warnings = %#v, want none", summary.Warnings)
	}
}

func TestDiscoverPNPMWorkspaceFixture(t *testing.T) {
	t.Parallel()
	detector := NewProjectDetector(privacy.NewFilePolicy())
	root := filepath.Join("..", "..", "tests", "fixtures", "node-pnpm-workspace")

	summary, err := detector.Discover(context.Background(), root)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	assertDetection(t, summary.Frameworks, "nestjs")
	assertDetection(t, summary.Frameworks, "fastify")
	assertPackageManager(t, summary.PackageManagers, "pnpm", "10.12.0")
	assertWorkspace(t, summary.Workspaces, "package-workspaces", []string{"packages/*"})
	assertWorkspace(t, summary.Workspaces, "pnpm-workspace", []string{})
}

func TestDiscoverReportsConflictingLockfiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTestFile(t, root, "package.json", `{"name":"conflict"}`)
	writeTestFile(t, root, "package-lock.json", `{}`)
	writeTestFile(t, root, "yarn.lock", "")

	summary, err := NewProjectDetector(privacy.NewFilePolicy()).Discover(context.Background(), root)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	assertWarning(t, summary.Warnings, "package-manager.conflicting-lockfiles")
}

func TestDiscoverReportsInvalidPackageJSON(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTestFile(t, root, "package.json", `{not-json`)

	summary, err := NewProjectDetector(privacy.NewFilePolicy()).Discover(context.Background(), root)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	assertWarning(t, summary.Warnings, "manifest.package-json-invalid")
}

func TestDiscoverUnsupportedDirectory(t *testing.T) {
	t.Parallel()
	summary, err := NewProjectDetector(privacy.NewFilePolicy()).Discover(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	assertWarning(t, summary.Warnings, "project.no-supported-markers")
}

func TestDiscoverIsDeterministic(t *testing.T) {
	t.Parallel()
	detector := NewProjectDetector(privacy.NewFilePolicy())
	root := filepath.Join("..", "..", "tests", "fixtures", "node-npm")

	first, err := detector.Discover(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := detector.Discover(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated discovery differed:\nfirst: %#v\nsecond: %#v", first, second)
	}
}

func TestDiscoverHonorsCancelledContext(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewProjectDetector(privacy.NewFilePolicy()).Discover(ctx, t.TempDir())
	if err == nil {
		t.Fatal("Discover() error = nil, want cancellation")
	}
}

func TestReadRootEntriesUsesBoundedSortedSubset(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, name := range []string{"c.txt", "a.txt", "b.txt"} {
		writeTestFile(t, root, name, "")
	}

	entries, truncated, err := readRootEntries(context.Background(), root, 2)
	if err != nil {
		t.Fatalf("readRootEntries() error = %v", err)
	}
	if !truncated {
		t.Fatal("readRootEntries() truncated = false, want true")
	}
	if len(entries) != 2 || entries[0].Name() != "a.txt" || entries[1].Name() != "b.txt" {
		t.Fatalf("entries = %#v, want sorted two-entry subset", entries)
	}
}

func TestReadRootEntriesPrioritizesRelevantMetadata(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTestFile(t, root, "a-unrelated.txt", "")
	writeTestFile(t, root, "package.json", `{}`)

	entries, truncated, err := readRootEntries(context.Background(), root, 1)
	if err != nil {
		t.Fatalf("readRootEntries() error = %v", err)
	}
	if !truncated {
		t.Fatal("readRootEntries() truncated = false, want true")
	}
	if len(entries) != 1 || entries[0].Name() != "package.json" {
		t.Fatalf("entries = %#v, want package.json retained", entries)
	}
}

func writeTestFile(t *testing.T, root, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertDetection(t *testing.T, detections []model.Detection, id string) {
	t.Helper()
	for _, detection := range detections {
		if detection.ID == id {
			return
		}
	}
	t.Fatalf("detection %q not found in %#v", id, detections)
}

func assertRuntime(t *testing.T, detections []model.RuntimeDetection, id, requirement string) {
	t.Helper()
	for _, detection := range detections {
		if detection.ID == id {
			if detection.Requirement != requirement {
				t.Fatalf("runtime %q requirement = %q, want %q", id, detection.Requirement, requirement)
			}
			return
		}
	}
	t.Fatalf("runtime %q not found in %#v", id, detections)
}

func assertPackageManager(t *testing.T, detections []model.PackageManagerDetection, id, version string) {
	t.Helper()
	for _, detection := range detections {
		if detection.ID == id {
			if detection.DeclaredVersion != version {
				t.Fatalf("package manager %q version = %q, want %q", id, detection.DeclaredVersion, version)
			}
			return
		}
	}
	t.Fatalf("package manager %q not found in %#v", id, detections)
}

func assertWorkspace(t *testing.T, detections []model.WorkspaceDetection, id string, patterns []string) {
	t.Helper()
	for _, detection := range detections {
		if detection.ID == id {
			if !reflect.DeepEqual(detection.Patterns, patterns) {
				t.Fatalf("workspace %q patterns = %#v, want %#v", id, detection.Patterns, patterns)
			}
			return
		}
	}
	t.Fatalf("workspace %q not found in %#v", id, detections)
}

func assertWarning(t *testing.T, warnings []model.Warning, code string) {
	t.Helper()
	for _, warning := range warnings {
		if warning.Code == code {
			return
		}
	}
	t.Fatalf("warning %q not found in %#v", code, warnings)
}
