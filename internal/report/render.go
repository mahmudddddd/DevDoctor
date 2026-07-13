package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/mahmudddddd/DevDoctor/internal/model"
)

// TextOptions controls human-readable report styling.
type TextOptions struct {
	Color bool
}

// WriteJSON writes an indented, versioned report for automation.
func WriteJSON(writer io.Writer, report model.ProjectReport) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("write JSON report: %w", err)
	}
	return nil
}

// WriteText writes a beginner-friendly project discovery summary.
func WriteText(writer io.Writer, report model.ProjectReport, options TextOptions) error {
	var output strings.Builder

	title := "DevDoctor project discovery"
	section := func(value string) string { return value }
	warning := func(value string) string { return value }
	if options.Color {
		title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render(title)
		sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		section = func(value string) string { return sectionStyle.Render(value) }
		warning = func(value string) string { return warningStyle.Render(value) }
	}

	fmt.Fprintln(&output, title)
	fmt.Fprintf(&output, "Project: %s\n", safeText(report.Project.Name))
	fmt.Fprintf(&output, "Root: %s\n", safeText(report.Project.Root))
	fmt.Fprintf(&output, "Report schema: %s\n", safeText(report.SchemaVersion))

	writeDetections(&output, section("Languages"), report.Project.Languages)
	writeRuntimeDetections(&output, section("Runtimes"), report.Project.Runtimes)
	writePackageManagers(&output, section("Package managers"), report.Project.PackageManagers)
	writeDetections(&output, section("Frameworks"), report.Project.Frameworks)
	writeWorkspaces(&output, section("Workspaces"), report.Project.Workspaces)

	fmt.Fprintf(&output, "\n%s\n", section("Relevant metadata"))
	if len(report.Project.RelevantFiles) == 0 {
		fmt.Fprintln(&output, "  None found")
	} else {
		for _, file := range report.Project.RelevantFiles {
			fmt.Fprintf(&output, "  - %s\n", safeText(file))
		}
	}

	if len(report.Project.Warnings) > 0 {
		fmt.Fprintf(&output, "\n%s\n", warning("Warnings"))
		for _, item := range report.Project.Warnings {
			fmt.Fprintf(&output, "  - [%s] %s\n", safeText(item.Code), safeText(item.Message))
		}
	}

	fmt.Fprintln(&output, "\nNo project scripts were executed during this discovery.")
	if _, err := io.WriteString(writer, output.String()); err != nil {
		return fmt.Errorf("write text report: %w", err)
	}
	return nil
}

func writeDetections(output *strings.Builder, title string, detections []model.Detection) {
	fmt.Fprintf(output, "\n%s\n", title)
	if len(detections) == 0 {
		fmt.Fprintln(output, "  None detected")
		return
	}
	for _, detection := range detections {
		fmt.Fprintf(output, "  - %s (%s confidence)\n", safeText(detection.Name), safeText(string(detection.Confidence)))
		writeEvidence(output, detection.Evidence)
	}
}

func writeRuntimeDetections(output *strings.Builder, title string, detections []model.RuntimeDetection) {
	fmt.Fprintf(output, "\n%s\n", title)
	if len(detections) == 0 {
		fmt.Fprintln(output, "  None detected")
		return
	}
	for _, detection := range detections {
		requirement := ""
		if detection.Requirement != "" {
			requirement = "; required " + safeText(detection.Requirement)
		}
		fmt.Fprintf(output, "  - %s (%s confidence%s)\n", safeText(detection.Name), safeText(string(detection.Confidence)), requirement)
		writeEvidence(output, detection.Evidence)
	}
}

func writePackageManagers(output *strings.Builder, title string, detections []model.PackageManagerDetection) {
	fmt.Fprintf(output, "\n%s\n", title)
	if len(detections) == 0 {
		fmt.Fprintln(output, "  None detected")
		return
	}
	for _, detection := range detections {
		version := ""
		if detection.DeclaredVersion != "" {
			version = "; declared " + safeText(detection.DeclaredVersion)
		}
		fmt.Fprintf(output, "  - %s (%s confidence%s)\n", safeText(detection.Name), safeText(string(detection.Confidence)), version)
		writeEvidence(output, detection.Evidence)
	}
}

func writeWorkspaces(output *strings.Builder, title string, detections []model.WorkspaceDetection) {
	fmt.Fprintf(output, "\n%s\n", title)
	if len(detections) == 0 {
		fmt.Fprintln(output, "  None detected")
		return
	}
	for _, detection := range detections {
		fmt.Fprintf(output, "  - %s\n", safeText(detection.Name))
		if len(detection.Patterns) > 0 {
			patterns := make([]string, len(detection.Patterns))
			for i, pattern := range detection.Patterns {
				patterns[i] = safeText(pattern)
			}
			fmt.Fprintf(output, "    Patterns: %s\n", strings.Join(patterns, ", "))
		}
		writeEvidence(output, detection.Evidence)
	}
}

func writeEvidence(output *strings.Builder, evidence []model.Evidence) {
	for _, item := range evidence {
		location := ""
		if item.Path != "" {
			location = safeText(item.Path) + ": "
		}
		fmt.Fprintf(output, "    Evidence: %s%s\n", location, safeText(item.Detail))
	}
}
