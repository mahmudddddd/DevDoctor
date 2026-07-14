package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mahmudddddd/DevDoctor/internal/model"
	"github.com/mahmudddddd/DevDoctor/internal/privacy"
)

const helperEnvironmentName = "GO_WANT_DEBUGDOC_HELPER"

func TestRunnerPreservesArgumentsWithoutShellInterpretation(t *testing.T) {
	t.Parallel()
	arguments := []string{"a b", "", "|", "$(not-run)", ">file", "quote\"value", "雪"}
	result, err := runHelper(context.Background(), t, helperSpec(t, "echo", arguments...))
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != model.CommandCompleted {
		t.Fatalf("status = %q, stderr = %q", result.Status, result.Stderr.Text)
	}
	var echoed []string
	if err := json.Unmarshal([]byte(result.Stdout.Text), &echoed); err != nil {
		t.Fatalf("decode stdout %q: %v", result.Stdout.Text, err)
	}
	if fmt.Sprint(echoed) != fmt.Sprint(arguments) {
		t.Fatalf("echoed = %#v, want %#v", echoed, arguments)
	}
	if result.Stderr.Text != "stderr-marker" {
		t.Fatalf("stderr = %q", result.Stderr.Text)
	}
}

func TestRunnerClassifiesNonzeroExit(t *testing.T) {
	t.Parallel()
	result, err := runHelper(context.Background(), t, helperSpec(t, "exit", "7"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != model.CommandFailed || result.Reason != model.ReasonNonzeroExit || result.ExitCode == nil || *result.ExitCode != 7 {
		t.Fatalf("result = %+v", result)
	}
}

func TestRunnerReportsStartFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	nonExecutable := filepath.Join(root, "not-an-executable.txt")
	if err := os.WriteFile(nonExecutable, []byte("plain text"), 0o600); err != nil {
		t.Fatal(err)
	}
	prepared, err := privacy.NewPathPolicy().PrepareCommand(root, model.CommandSpec{
		OperationID:      "test.start-failure",
		Purpose:          "Verify start failure classification",
		Executable:       nonExecutable,
		WorkingDirectory: root,
		Mutation:         model.MutationNone,
		Network:          model.NetworkNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := New().Run(context.Background(), prepared)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != model.CommandFailed || result.Reason != model.ReasonStartFailed {
		t.Fatalf("result = %+v", result)
	}
}

func TestRunnerTimesOutAndCancels(t *testing.T) {
	t.Parallel()
	t.Run("timeout", func(t *testing.T) {
		spec := helperSpec(t, "sleep", "10s")
		spec.Timeout = 100 * time.Millisecond
		spec.TerminationGrace = 50 * time.Millisecond
		result, err := runHelper(context.Background(), t, spec)
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if result.Status != model.CommandTimedOut || result.Reason != model.ReasonTimedOut || !result.CleanupComplete {
			t.Fatalf("result = %+v", result)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(100*time.Millisecond, cancel)
		spec := helperSpec(t, "sleep", "10s")
		spec.Timeout = 5 * time.Second
		spec.TerminationGrace = 50 * time.Millisecond
		result, err := runHelper(ctx, t, spec)
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if result.Status != model.CommandCancelled || result.Reason != model.ReasonCancelled || !result.CleanupComplete {
			t.Fatalf("result = %+v", result)
		}
	})
}

func TestRunnerBoundsStdoutAndStderrIndependently(t *testing.T) {
	t.Parallel()
	spec := helperSpec(t, "flood", strconv.Itoa(2*1024*1024))
	spec.OutputLimit = 4096
	result, err := runHelper(context.Background(), t, spec)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != model.CommandCompleted {
		t.Fatalf("result = %+v", result)
	}
	for name, capture := range map[string]model.StreamCapture{"stdout": result.Stdout, "stderr": result.Stderr} {
		if capture.CapturedBytes != 4096 || capture.DiscardedBytes != 2*1024*1024-4096 || !capture.Truncated {
			t.Fatalf("%s capture = %+v", name, capture)
		}
	}
}

func TestRunnerCleansDescendantTreeOnTimeout(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "pids.txt")
	spec := helperSpec(t, "tree-parent", pidFile)
	spec.Timeout = 500 * time.Millisecond
	spec.TerminationGrace = 50 * time.Millisecond
	result, err := runHelper(context.Background(), t, spec)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != model.CommandTimedOut || !result.CleanupComplete {
		t.Fatalf("result = %+v", result)
	}
	pids := readHelperPIDs(t, pidFile)
	if len(pids) < 3 {
		t.Fatalf("recorded pids = %v, want parent, child, grandchild", pids)
	}
	for _, pid := range pids {
		if !waitForProcessExit(pid, 3*time.Second) {
			t.Errorf("process %d remained alive after cleanup", pid)
		}
	}
}

func TestRunnerRepeatedFloodCancellation(t *testing.T) {
	for iteration := 0; iteration < 3; iteration++ {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(300*time.Millisecond, cancel)
		spec := helperSpec(t, "flood-sleep", strconv.Itoa(512*1024))
		spec.Timeout = 5 * time.Second
		spec.TerminationGrace = 25 * time.Millisecond
		spec.OutputLimit = 1024
		result, err := runHelper(ctx, t, spec)
		if err != nil {
			t.Fatalf("iteration %d: Run() error = %v", iteration, err)
		}
		if result.Status != model.CommandCancelled || !result.Stdout.Truncated || !result.Stderr.Truncated {
			t.Fatalf("iteration %d: result = %+v", iteration, result)
		}
	}
}

func runHelper(ctx context.Context, t *testing.T, spec model.CommandSpec) (model.CommandResult, error) {
	t.Helper()
	root := spec.WorkingDirectory
	prepared, err := privacy.NewPathPolicy().PrepareCommand(root, spec)
	if err != nil {
		t.Fatalf("PrepareCommand() error = %v", err)
	}
	return New().Run(ctx, prepared)
}

func helperSpec(t *testing.T, mode string, arguments ...string) model.CommandSpec {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	commandArguments := make([]string, 0, 3+len(arguments))
	commandArguments = append(commandArguments, "-test.run=^TestRunnerHelperProcess$", "--", mode)
	commandArguments = append(commandArguments, arguments...)
	return model.CommandSpec{
		OperationID:      "test.helper." + mode,
		Purpose:          "Exercise the shell-free runner",
		Executable:       executable,
		Arguments:        commandArguments,
		WorkingDirectory: root,
		Timeout:          5 * time.Second,
		TerminationGrace: 100 * time.Millisecond,
		OutputLimit:      256 * 1024,
		Environment: model.EnvironmentSpec{Set: map[string]string{
			helperEnvironmentName: "1",
		}},
		Mutation: model.MutationNone,
		Network:  model.NetworkNone,
	}
}

func TestRunnerHelperProcess(_ *testing.T) {
	if os.Getenv(helperEnvironmentName) != "1" {
		return
	}
	arguments := helperArguments()
	if len(arguments) == 0 {
		os.Exit(2)
	}
	switch arguments[0] {
	case "echo":
		_ = json.NewEncoder(os.Stdout).Encode(arguments[1:])
		_, _ = fmt.Fprint(os.Stderr, "stderr-marker")
	case "exit":
		code, _ := strconv.Atoi(arguments[1])
		os.Exit(code)
	case "sleep":
		duration, _ := time.ParseDuration(arguments[1])
		time.Sleep(duration)
	case "ignore-term":
		ignoreTerminationSignal()
		time.Sleep(10 * time.Minute)
	case "flood", "flood-sleep":
		count, _ := strconv.Atoi(arguments[1])
		var wait sync.WaitGroup
		wait.Add(2)
		go func() {
			defer wait.Done()
			_ = ioWriteRepeated(os.Stdout, 'o', count)
		}()
		go func() {
			defer wait.Done()
			_ = ioWriteRepeated(os.Stderr, 'e', count)
		}()
		wait.Wait()
		if arguments[0] == "flood-sleep" {
			time.Sleep(10 * time.Second)
		}
	case "tree-parent":
		recordHelperPID(arguments[1])
		startHelperDescendant("tree-child", arguments[1])
		time.Sleep(10 * time.Minute)
	case "tree-child":
		recordHelperPID(arguments[1])
		startHelperDescendant("tree-grandchild", arguments[1])
		time.Sleep(10 * time.Minute)
	case "tree-grandchild":
		recordHelperPID(arguments[1])
		time.Sleep(10 * time.Minute)
	default:
		os.Exit(3)
	}
	os.Exit(0)
}

func helperArguments() []string {
	for index, argument := range os.Args {
		if argument == "--" {
			return os.Args[index+1:]
		}
	}
	return nil
}

func ioWriteRepeated(file *os.File, value byte, count int) error {
	buffer := []byte(strings.Repeat(string(value), 32*1024))
	written := 0
	for written < count {
		chunk := count - written
		if chunk > len(buffer) {
			chunk = len(buffer)
		}
		current, err := file.Write(buffer[:chunk])
		written += current
		if err != nil {
			return err
		}
	}
	return nil
}

func startHelperDescendant(mode, pidFile string) {
	executable, _ := os.Executable()
	command := exec.Command(executable, "-test.run=^TestRunnerHelperProcess$", "--", mode, pidFile)
	if err := command.Start(); err != nil {
		os.Exit(4)
	}
}

func recordHelperPID(path string) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		os.Exit(5)
	}
	_, _ = fmt.Fprintln(file, os.Getpid())
	_ = file.Close()
}

func readHelperPIDs(t *testing.T, path string) []int {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close PID file: %v", err)
		}
	}()
	var pids []int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pid, err := strconv.Atoi(scanner.Text())
		if err != nil {
			t.Fatal(err)
		}
		pids = append(pids, pid)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return pids
}
