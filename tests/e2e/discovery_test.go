package e2e

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mahmudddddd/DevDoctor/internal/app"
	"github.com/mahmudddddd/DevDoctor/internal/cli"
	"github.com/mahmudddddd/DevDoctor/internal/model"
)

func TestDiagnoseFixtureAsJSON(t *testing.T) {
	t.Parallel()
	fixture, err := filepath.Abs(filepath.Join("..", "fixtures", "node-npm"))
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	command := cli.NewRootCommand(cli.Dependencies{
		Input:         strings.NewReader(""),
		Output:        &output,
		ErrorOutput:   &output,
		IsInteractive: func() bool { return false },
		UseColor:      func() bool { return false },
		Discovery:     app.NewDiscoveryService(),
	})
	command.SetArgs([]string{"diagnose", "--path", fixture, "--format", "json"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	var report model.ProjectReport
	if err := json.Unmarshal(output.Bytes(), &report); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, output.String())
	}
	if report.Project.Name != "fixture-node-npm" {
		t.Fatalf("project name = %q, want fixture-node-npm", report.Project.Name)
	}
	if report.SchemaVersion != model.ReportSchemaVersion {
		t.Fatalf("schema version = %q, want %q", report.SchemaVersion, model.ReportSchemaVersion)
	}
}
