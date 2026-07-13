package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mahmudddddd/DevDoctor/internal/consent"
	"github.com/mahmudddddd/DevDoctor/internal/model"
	"github.com/mahmudddddd/DevDoctor/internal/privacy"
)

func TestExecutionServiceFailsClosedWithoutStartingRunner(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		decision   consent.Decision
		wantReason model.CommandReason
	}{
		{name: "denied", decision: consent.Decision{Outcome: consent.Denied}, wantReason: model.ReasonApprovalDenied},
		{name: "unavailable", decision: consent.Decision{Outcome: consent.Unavailable}, wantReason: model.ReasonApprovalUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := &fakeCommandRunner{}
			service := ExecutionService{
				preparer:   privacy.NewPathPolicy(),
				authorizer: fakeCommandAuthorizer{decision: test.decision},
				runner:     runner,
			}
			root, spec := executionTestSpec(t)
			result, err := service.Execute(context.Background(), root, spec)
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != model.CommandSkipped || result.Reason != test.wantReason || runner.calls != 0 {
				t.Fatalf("result = %+v, runner calls = %d", result, runner.calls)
			}
		})
	}
}

func TestExecutionServiceRunsOnlyApprovedPreparedCommand(t *testing.T) {
	t.Parallel()
	runner := &fakeCommandRunner{result: model.CommandResult{Status: model.CommandCompleted, CleanupComplete: true}}
	service := ExecutionService{
		preparer: privacy.NewPathPolicy(),
		authorizer: fakeCommandAuthorizer{decision: consent.Decision{
			Outcome: consent.Approved,
			Scope:   consent.ScopeOnce,
		}},
		runner: runner,
	}
	root, spec := executionTestSpec(t)
	result, err := service.Execute(context.Background(), root, spec)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != model.CommandCompleted || runner.calls != 1 {
		t.Fatalf("result = %+v, runner calls = %d", result, runner.calls)
	}
	if runner.prepared.Spec().Executable != spec.Executable {
		t.Fatalf("runner received executable %q, want %q", runner.prepared.Spec().Executable, spec.Executable)
	}
}

func TestExecutionServiceRevalidatesAfterApproval(t *testing.T) {
	root, spec := executionTestSpec(t)
	working := filepath.Join(root, "nested")
	if err := os.Mkdir(working, 0o755); err != nil {
		t.Fatal(err)
	}
	spec.WorkingDirectory = working
	runner := &fakeCommandRunner{}
	service := ExecutionService{
		preparer: privacy.NewPathPolicy(),
		authorizer: mutatingAuthorizer{mutate: func() {
			_ = os.Rename(working, working+"-moved")
		}},
		runner: runner,
	}
	_, err := service.Execute(context.Background(), root, spec)
	if err == nil || !strings.Contains(err.Error(), "revalidate approved command") {
		t.Fatalf("Execute() error = %v", err)
	}
	if runner.calls != 0 {
		t.Fatalf("runner calls = %d, want 0", runner.calls)
	}
}

func executionTestSpec(t *testing.T) (string, model.CommandSpec) {
	t.Helper()
	root := t.TempDir()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return root, model.CommandSpec{
		OperationID:      "test.execute",
		Purpose:          "Verify application execution orchestration",
		Executable:       executable,
		Arguments:        []string{"argument"},
		WorkingDirectory: root,
		Mutation:         model.MutationNone,
		Network:          model.NetworkNone,
	}
}

type fakeCommandAuthorizer struct {
	decision consent.Decision
	err      error
}

func (authorizer fakeCommandAuthorizer) Authorize(context.Context, privacy.PreparedCommand, []string) (consent.Decision, error) {
	return authorizer.decision, authorizer.err
}

type mutatingAuthorizer struct {
	mutate func()
}

func (authorizer mutatingAuthorizer) Authorize(context.Context, privacy.PreparedCommand, []string) (consent.Decision, error) {
	authorizer.mutate()
	return consent.Decision{Outcome: consent.Approved, Scope: consent.ScopeOnce}, nil
}

type fakeCommandRunner struct {
	calls    int
	prepared privacy.PreparedCommand
	result   model.CommandResult
	err      error
}

func (runner *fakeCommandRunner) Run(_ context.Context, prepared privacy.PreparedCommand) (model.CommandResult, error) {
	runner.calls++
	runner.prepared = prepared
	return runner.result, runner.err
}
