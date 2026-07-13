package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mahmudddddd/DevDoctor/internal/model"
)

// WriteCommandResultJSON writes an internal command result without changing ProjectReport.
func WriteCommandResultJSON(output io.Writer, result model.CommandResult) error {
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// WriteCommandResultText writes a safe, explicit command lifecycle summary.
func WriteCommandResultText(output io.Writer, result model.CommandResult) error {
	if _, err := fmt.Fprintf(output, "Status: %s\n", result.Status); err != nil {
		return err
	}
	if result.Reason != "" {
		if _, err := fmt.Fprintf(output, "Reason: %s\n", result.Reason); err != nil {
			return err
		}
	}
	if result.ExitCode != nil {
		if _, err := fmt.Fprintf(output, "Exit code: %d\n", *result.ExitCode); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(output, "Cleanup complete: %t\n", result.CleanupComplete); err != nil {
		return err
	}
	if result.CleanupError != "" {
		if _, err := fmt.Fprintf(output, "Cleanup error: %s\n", SafeText(result.CleanupError)); err != nil {
			return err
		}
	}
	if err := writeStreamCapture(output, "stdout", result.Stdout); err != nil {
		return err
	}
	return writeStreamCapture(output, "stderr", result.Stderr)
}

func writeStreamCapture(output io.Writer, name string, capture model.StreamCapture) error {
	if _, err := fmt.Fprintf(output, "%s (%d captured bytes", name, capture.CapturedBytes); err != nil {
		return err
	}
	if capture.Truncated {
		if _, err := fmt.Fprintf(output, ", %d discarded bytes, truncated", capture.DiscardedBytes); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(output, "):"); err != nil {
		return err
	}
	_, err := fmt.Fprintln(output, SafeText(capture.Text))
	return err
}
