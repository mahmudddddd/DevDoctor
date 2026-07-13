package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mahmudddddd/DevDoctor/internal/model"
)

type fakeDiscovery struct {
	report model.ProjectReport
	err    error
	calls  int
	path   string
}

func (f *fakeDiscovery) Diagnose(_ context.Context, path string) (model.ProjectReport, error) {
	f.calls++
	f.path = path
	return f.report, f.err
}

func TestRootWithoutArgumentsDoesNotPromptWhenNonInteractive(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	discovery := &fakeDiscovery{}
	command := NewRootCommand(Dependencies{
		Input:         strings.NewReader(""),
		Output:        &output,
		ErrorOutput:   &output,
		IsInteractive: func() bool { return false },
		UseColor:      func() bool { return false },
		Discovery:     discovery,
	})
	command.SetArgs(nil)

	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "interactive mode requires a terminal") {
		t.Fatalf("Execute() error = %v, want non-interactive guidance", err)
	}
	if discovery.calls != 0 {
		t.Fatalf("discovery calls = %d, want 0", discovery.calls)
	}
}

func TestDiagnoseWritesJSONWithoutPrompting(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	discovery := &fakeDiscovery{report: model.ProjectReport{
		SchemaVersion: model.ReportSchemaVersion,
		ToolVersion:   "test",
		Project: model.ProjectSummary{
			Root: "/project",
			Name: "fixture",
		},
	}}
	command := NewRootCommand(Dependencies{
		Input:         strings.NewReader("should not be read"),
		Output:        &output,
		ErrorOutput:   &output,
		IsInteractive: func() bool { return false },
		UseColor:      func() bool { return false },
		Discovery:     discovery,
	})
	command.SetArgs([]string{"diagnose", "--path", "./fixture", "--format", "json"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if discovery.calls != 1 || discovery.path != "./fixture" {
		t.Fatalf("discovery calls/path = %d/%q, want 1/%q", discovery.calls, discovery.path, "./fixture")
	}
	if !strings.Contains(output.String(), `"schemaVersion": "1.0"`) {
		t.Fatalf("output = %q, want JSON schema version", output.String())
	}
}

func TestDiagnoseRejectsUnknownFormatBeforeDiscovery(t *testing.T) {
	t.Parallel()
	discovery := &fakeDiscovery{}
	command := NewRootCommand(Dependencies{
		Input:         strings.NewReader(""),
		Output:        &bytes.Buffer{},
		ErrorOutput:   &bytes.Buffer{},
		IsInteractive: func() bool { return false },
		UseColor:      func() bool { return false },
		Discovery:     discovery,
	})
	command.SetArgs([]string{"diagnose", "--format", "xml"})

	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf("Execute() error = %v, want unsupported format", err)
	}
	if discovery.calls != 0 {
		t.Fatalf("discovery calls = %d, want 0", discovery.calls)
	}
}

func TestDiagnosePropagatesDiscoveryError(t *testing.T) {
	t.Parallel()
	discovery := &fakeDiscovery{err: errors.New("boom")}
	command := NewRootCommand(Dependencies{
		Input:         strings.NewReader(""),
		Output:        &bytes.Buffer{},
		ErrorOutput:   &bytes.Buffer{},
		IsInteractive: func() bool { return false },
		UseColor:      func() bool { return false },
		Discovery:     discovery,
	})
	command.SetArgs([]string{"diagnose"})

	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "discover project: boom") {
		t.Fatalf("Execute() error = %v, want wrapped discovery error", err)
	}
}
