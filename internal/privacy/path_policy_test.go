package privacy

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/mahmudddddd/DevDoctor/internal/model"
)

func TestPathPolicyAllowsNestedWorkingDirectoryAndAppliesDefaults(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	working := filepath.Join(root, "packages", "api")
	if err := os.MkdirAll(working, 0o755); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	prepared, err := NewPathPolicy().PrepareCommand(root, testCommandSpec(executable, filepath.Join("packages", "api")))
	if err != nil {
		t.Fatalf("PrepareCommand() error = %v", err)
	}
	spec := prepared.Spec()
	requireSamePath(t, spec.WorkingDirectory, working)
	if spec.Timeout != 2*time.Minute || spec.TerminationGrace != 2*time.Second || spec.OutputLimit != 256*1024 {
		t.Fatalf("unexpected defaults: %+v", spec)
	}
	if err := prepared.Revalidate(); err != nil {
		t.Fatalf("Revalidate() error = %v", err)
	}
}

func TestPathPolicyRejectsUnsafeWorkingDirectories(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	outside := t.TempDir()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(root, "file")
	if err := os.WriteFile(filePath, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		working string
		want    string
	}{
		{name: "outside", working: outside, want: "outside the project root"},
		{name: "traversal", working: filepath.Join("..", filepath.Base(outside)), want: "outside the project root"},
		{name: "missing", working: filepath.Join(root, "missing"), want: "resolve path"},
		{name: "file", working: filePath, want: "not a directory"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewPathPolicy().PrepareCommand(root, testCommandSpec(executable, test.working))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("PrepareCommand() error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestPathPolicyRejectsRelativeAndNonregularExecutables(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	_, err := NewPathPolicy().PrepareCommand(root, testCommandSpec("debugdoc", root))
	if err == nil || !strings.Contains(err.Error(), "absolute") {
		t.Fatalf("relative executable error = %v", err)
	}
	_, err = NewPathPolicy().PrepareCommand(root, testCommandSpec(root, root))
	if err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("directory executable error = %v", err)
	}
}

func TestPathPolicyRejectsSymlinkEscape(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink creation unavailable: %v", err)
		}
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewPathPolicy().PrepareCommand(root, testCommandSpec(executable, link))
	if err == nil || !strings.Contains(err.Error(), "outside the project root") {
		t.Fatalf("PrepareCommand() error = %v", err)
	}
}

func TestPreparedCommandDetectsWorkingDirectoryMove(t *testing.T) {
	root := t.TempDir()
	working := filepath.Join(root, "work")
	if err := os.Mkdir(working, 0o755); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := NewPathPolicy().PrepareCommand(root, testCommandSpec(executable, working))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(working, working+"-old"); err != nil {
		t.Fatal(err)
	}
	if err := prepared.Revalidate(); err == nil {
		t.Fatal("Revalidate() succeeded after the approved directory moved")
	}
}

func TestPreparedCommandReturnsDefensiveCopies(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	spec := testCommandSpec(executable, root)
	spec.Arguments = []string{"original"}
	spec.Environment.Set = map[string]string{"SAFE_VALUE": "original"}
	prepared, err := NewPathPolicy().PrepareCommand(root, spec)
	if err != nil {
		t.Fatal(err)
	}
	copySpec := prepared.Spec()
	copySpec.Arguments[0] = "changed"
	copySpec.Environment.Set["SAFE_VALUE"] = "changed"
	again := prepared.Spec()
	if again.Arguments[0] != "original" || again.Environment.Set["SAFE_VALUE"] != "original" {
		t.Fatalf("prepared command was mutated: %+v", again)
	}
}

func requireSamePath(t *testing.T, got, want string) {
	t.Helper()
	gotInfo, err := os.Stat(got)
	if err != nil {
		t.Fatalf("stat received path %q: %v", got, err)
	}
	wantInfo, err := os.Stat(want)
	if err != nil {
		t.Fatalf("stat expected path %q: %v", want, err)
	}
	if !os.SameFile(gotInfo, wantInfo) {
		t.Fatalf("received path %q and expected path %q identify different filesystem objects", got, want)
	}
}

func testCommandSpec(executable, working string) model.CommandSpec {
	return model.CommandSpec{
		OperationID:      "test.command",
		Purpose:          "Exercise the command safety boundary",
		Executable:       executable,
		WorkingDirectory: working,
		Mutation:         model.MutationNone,
		Network:          model.NetworkNone,
	}
}
