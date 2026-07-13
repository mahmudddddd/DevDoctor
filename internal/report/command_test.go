package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mahmudddddd/DevDoctor/internal/model"
)

func TestWriteCommandResultTextEscapesControlsAndShowsTruncation(t *testing.T) {
	t.Parallel()
	result := model.CommandResult{
		Status: model.CommandCompleted,
		Stdout: model.StreamCapture{
			Text:           "safe\x1b[31m",
			CapturedBytes:  8,
			DiscardedBytes: 100,
			Truncated:      true,
		},
		Stderr:          model.StreamCapture{},
		CleanupComplete: true,
	}
	var output bytes.Buffer
	if err := WriteCommandResultText(&output, result); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	if strings.Contains(text, "\x1b") || !strings.Contains(text, "safe\\u001B[31m") || !strings.Contains(text, "100 discarded bytes, truncated") {
		t.Fatalf("output = %q", text)
	}
}

func TestWriteCommandResultJSONIncludesCaptureMetadata(t *testing.T) {
	t.Parallel()
	result := model.CommandResult{
		Status:          model.CommandTimedOut,
		Reason:          model.ReasonTimedOut,
		Stdout:          model.StreamCapture{CapturedBytes: 4, DiscardedBytes: 9, Truncated: true},
		CleanupComplete: true,
	}
	var output bytes.Buffer
	if err := WriteCommandResultJSON(&output, result); err != nil {
		t.Fatal(err)
	}
	var decoded model.CommandResult
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Stdout.CapturedBytes != 4 || decoded.Stdout.DiscardedBytes != 9 || !decoded.Stdout.Truncated {
		t.Fatalf("decoded = %+v", decoded)
	}
}
