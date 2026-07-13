package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/mahmudddddd/DevDoctor/internal/model"
)

func TestWriteJSONProducesVersionedReport(t *testing.T) {
	t.Parallel()
	report := testReport()
	var output bytes.Buffer

	if err := WriteJSON(&output, report); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}
	var decoded model.ProjectReport
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if decoded.SchemaVersion != model.ReportSchemaVersion {
		t.Fatalf("schema version = %q, want %q", decoded.SchemaVersion, model.ReportSchemaVersion)
	}
}

func TestWriteTextExplainsDiscoverySafety(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	if err := WriteText(&output, testReport(), TextOptions{}); err != nil {
		t.Fatalf("WriteText() error = %v", err)
	}
	for _, want := range []string{"DevDoctor project discovery", "JavaScript", "No project scripts were executed"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("output %q does not contain %q", output.String(), want)
		}
	}
	if strings.Contains(output.String(), "\x1b[") {
		t.Fatalf("plain output contains ANSI escape: %q", output.String())
	}
}

func TestWriteTextEscapesProjectControlledTerminalCharacters(t *testing.T) {
	t.Parallel()
	report := testReport()
	report.Project.Name = "safe\x1b[2J\nspoofed"
	report.Project.Languages[0].Evidence = []model.Evidence{{
		Path:   "package.json\r",
		Detail: "dependency\x1b]8;;https://example.com\a",
	}}

	var output bytes.Buffer
	if err := WriteText(&output, report, TextOptions{}); err != nil {
		t.Fatalf("WriteText() error = %v", err)
	}
	if strings.ContainsAny(output.String(), "\x1b\r\a") || strings.Contains(output.String(), "\nspoofed\n") {
		t.Fatalf("output contains raw terminal control characters: %q", output.String())
	}
	for _, character := range []rune{0x1b, '\n', '\r', 0x07} {
		want := fmt.Sprintf("%c%c%04X", 92, 'u', character)
		if !strings.Contains(output.String(), want) {
			t.Fatalf("output %q does not contain escaped control %q", output.String(), want)
		}
	}
}

func testReport() model.ProjectReport {
	return model.ProjectReport{
		SchemaVersion: model.ReportSchemaVersion,
		ToolVersion:   "test",
		Project: model.ProjectSummary{
			Root: "/tmp/project",
			Name: "project",
			Languages: []model.Detection{{
				ID: "javascript", Name: "JavaScript", Confidence: model.ConfidenceHigh,
			}},
		},
	}
}
