package consent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mahmudddddd/DevDoctor/internal/model"
	"github.com/mahmudddddd/DevDoctor/internal/privacy"
)

func TestManagerAppliesApprovalScopesInMemory(t *testing.T) {
	t.Parallel()
	prepared := preparedConsentCommand(t, baseConsentSpec(t))
	tests := []struct {
		name      string
		decision  Decision
		wantCalls int
	}{
		{name: "once", decision: Decision{Outcome: Approved, Scope: ScopeOnce}, wantCalls: 2},
		{name: "check", decision: Decision{Outcome: Approved, Scope: ScopeThisCheck}, wantCalls: 1},
		{name: "run", decision: Decision{Outcome: Approved, Scope: ScopeThisRun}, wantCalls: 1},
		{name: "denied", decision: Decision{Outcome: Denied}, wantCalls: 2},
		{name: "unavailable", decision: Decision{Outcome: Unavailable}, wantCalls: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			approver := &recordingApprover{decision: test.decision}
			manager := NewManager(approver)
			for range 2 {
				if _, err := manager.Authorize(context.Background(), prepared, []string{"PATH", "SAFE_NAME"}); err != nil {
					t.Fatal(err)
				}
			}
			if approver.calls != test.wantCalls {
				t.Fatalf("approver calls = %d, want %d", approver.calls, test.wantCalls)
			}
		})
	}
}

func TestManagerFingerprintCoversEveryMaterialField(t *testing.T) {
	base := baseConsentSpec(t)
	approver := &recordingApprover{decision: Decision{Outcome: Approved, Scope: ScopeThisRun}}
	manager := NewManager(approver)
	authorize := func(spec model.CommandSpec, environmentNames []string) {
		t.Helper()
		prepared := preparedConsentCommand(t, spec)
		if _, err := manager.Authorize(context.Background(), prepared, environmentNames); err != nil {
			t.Fatal(err)
		}
	}
	authorize(base, []string{"PATH", "SAFE_NAME"})

	variations := []model.CommandSpec{
		withConsentSpec(base, func(spec *model.CommandSpec) { spec.OperationID = "changed.operation" }),
		withConsentSpec(base, func(spec *model.CommandSpec) { spec.Purpose = "Changed purpose" }),
		withConsentSpec(base, func(spec *model.CommandSpec) { spec.Arguments = []string{"changed"} }),
		withConsentSpec(base, func(spec *model.CommandSpec) { spec.WorkingDirectory = filepath.Join(spec.WorkingDirectory, "nested") }),
		withConsentSpec(base, func(spec *model.CommandSpec) { spec.Timeout += time.Second }),
		withConsentSpec(base, func(spec *model.CommandSpec) { spec.TerminationGrace += time.Second }),
		withConsentSpec(base, func(spec *model.CommandSpec) { spec.OutputLimit++ }),
		withConsentSpec(base, func(spec *model.CommandSpec) { spec.Mutation = model.MutationProject }),
		withConsentSpec(base, func(spec *model.CommandSpec) { spec.Network = model.NetworkRead }),
		withConsentSpec(base, func(spec *model.CommandSpec) { spec.StartsService = true }),
		withConsentSpec(base, func(spec *model.CommandSpec) { spec.Environment.Set = map[string]string{"SAFE_NAME": "changed"} }),
		withConsentSpec(base, func(spec *model.CommandSpec) { spec.DataBundles = []model.DataBundleDescriptor{{Name: "logs"}} }),
	}
	if err := os.Mkdir(filepath.Join(base.WorkingDirectory, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, variation := range variations {
		authorize(variation, []string{"PATH", "SAFE_NAME"})
	}
	authorize(base, []string{"PATH", "SAFE_NAME", "TMP"})

	wantCalls := 1 + len(variations) + 1
	if approver.calls != wantCalls {
		t.Fatalf("approver calls = %d, want %d", approver.calls, wantCalls)
	}
}

func TestRequestIsImmutableAndDoesNotExposeEnvironmentValues(t *testing.T) {
	t.Parallel()
	prepared := preparedConsentCommand(t, baseConsentSpec(t))
	request, err := newRequest(prepared, []string{"PATH", "SAFE_NAME"})
	if err != nil {
		t.Fatal(err)
	}
	arguments := request.Arguments()
	arguments[0] = "changed"
	names := request.EnvironmentNames()
	names[0] = "changed"
	bundles := request.DataBundles()
	bundles[0].Name = "changed"
	if request.Arguments()[0] != "base" || request.EnvironmentNames()[0] != "PATH" || request.DataBundles()[0].Name != "manifest" {
		t.Fatal("request getters exposed mutable storage")
	}
	for _, name := range request.EnvironmentNames() {
		if name == "private-value" {
			t.Fatal("environment value was exposed")
		}
	}
}

func baseConsentSpec(t *testing.T) model.CommandSpec {
	t.Helper()
	root := t.TempDir()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return model.CommandSpec{
		OperationID:      "test.consent",
		Purpose:          "Verify exact consent",
		Executable:       executable,
		Arguments:        []string{"base"},
		WorkingDirectory: root,
		Timeout:          3 * time.Second,
		TerminationGrace: time.Second,
		OutputLimit:      4096,
		Environment: model.EnvironmentSpec{Set: map[string]string{
			"SAFE_NAME": "private-value",
		}},
		Mutation:    model.MutationNone,
		Network:     model.NetworkNone,
		DataBundles: []model.DataBundleDescriptor{{Name: "manifest", Description: "package metadata"}},
	}
}

func preparedConsentCommand(t *testing.T, spec model.CommandSpec) privacy.PreparedCommand {
	t.Helper()
	prepared, err := privacy.NewPathPolicy().PrepareCommand(spec.WorkingDirectory, spec)
	if err != nil {
		t.Fatal(err)
	}
	return prepared
}

func withConsentSpec(base model.CommandSpec, change func(*model.CommandSpec)) model.CommandSpec {
	copySpec := base
	copySpec.Arguments = append([]string(nil), base.Arguments...)
	copySpec.Environment.Set = make(map[string]string, len(base.Environment.Set))
	for name, value := range base.Environment.Set {
		copySpec.Environment.Set[name] = value
	}
	copySpec.DataBundles = append([]model.DataBundleDescriptor(nil), base.DataBundles...)
	change(&copySpec)
	return copySpec
}

type recordingApprover struct {
	decision Decision
	calls    int
	request  Request
	err      error
}

func (approver *recordingApprover) Approve(_ context.Context, request Request) (Decision, error) {
	approver.calls++
	approver.request = request
	return approver.decision, approver.err
}
