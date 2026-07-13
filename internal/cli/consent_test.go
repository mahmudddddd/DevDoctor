package cli

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mahmudddddd/DevDoctor/internal/consent"
	"github.com/mahmudddddd/DevDoctor/internal/model"
	"github.com/mahmudddddd/DevDoctor/internal/privacy"
)

func TestApprovalRenderingShowsExactRequestWithoutEnvironmentValues(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := privacy.NewPathPolicy().PrepareCommand(root, model.CommandSpec{
		OperationID:      "test.render",
		Purpose:          "Render\x1b[31m safely",
		Executable:       executable,
		Arguments:        []string{"a b", "|", ""},
		WorkingDirectory: root,
		Timeout:          3 * time.Second,
		TerminationGrace: time.Second,
		OutputLimit:      4096,
		Environment: model.EnvironmentSpec{Set: map[string]string{
			"SAFE_NAME": "private-value",
		}},
		Mutation:      model.MutationProject,
		Network:       model.NetworkRead,
		StartsService: true,
		DataBundles: []model.DataBundleDescriptor{{
			Name:        "manifest",
			Description: "approved metadata",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	capture := &captureRequestApprover{}
	manager := consent.NewManager(capture)
	if _, err := manager.Authorize(context.Background(), prepared, []string{"PATH", "SAFE_NAME"}); err != nil {
		t.Fatal(err)
	}
	rendered := formatApprovalRequest(capture.request)
	for _, expected := range []string{
		"Operation: test.render",
		"Purpose: Render\\u001B[31m safely",
		"Executable: " + executable,
		`1: "a b"`,
		`2: "|"`,
		`3: ""`,
		"Working directory: " + root,
		"Mutation: project",
		"Network: read",
		"Starts service: true",
		"Timeout: 3s",
		"Termination grace: 1s",
		"Output limit: 4096 bytes per stream",
		"SAFE_NAME",
		"manifest: approved metadata",
	} {
		if !strings.Contains(rendered, expected) {
			t.Errorf("rendered request missing %q:\n%s", expected, rendered)
		}
	}
	if strings.Contains(rendered, "private-value") || strings.Contains(rendered, "\x1b") {
		t.Fatalf("rendered request exposed a value or terminal escape: %q", rendered)
	}
}

func TestNonInteractiveApproverFailsClosedWithoutInput(t *testing.T) {
	t.Parallel()
	decision, err := (NonInteractiveApprover{}).Approve(context.Background(), consent.Request{})
	if err != nil {
		t.Fatal(err)
	}
	if decision.Outcome != consent.Unavailable {
		t.Fatalf("outcome = %q", decision.Outcome)
	}
}

type captureRequestApprover struct {
	request consent.Request
}

func (approver *captureRequestApprover) Approve(_ context.Context, request consent.Request) (consent.Decision, error) {
	approver.request = request
	return consent.Decision{Outcome: consent.Approved, Scope: consent.ScopeOnce}, nil
}
