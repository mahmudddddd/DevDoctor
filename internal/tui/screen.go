package tui

import (
	"fmt"
	"strings"

	"github.com/mahmudddddd/DevDoctor/internal/model"
	"github.com/mahmudddddd/DevDoctor/internal/report"
)

func screenContent(screen Screen, state RunState, projectPath string, discoveryReport *model.ProjectReport, runError string, width, warningSelected int, warningDetail bool) []string {
	switch screen {
	case ScreenDiagnose:
		return diagnoseContent(state, projectPath, discoveryReport, runError, width)
	case ScreenProject:
		return projectContent(discoveryReport, width)
	case ScreenWarnings:
		return warningsContent(discoveryReport, warningSelected, warningDetail, width)
	case ScreenExport:
		return exportContent(discoveryReport, width)
	case ScreenHelp:
		return helpContent()
	case ScreenHome:
		return homeContent(discoveryReport)
	default:
		return homeContent(discoveryReport)
	}
}

func homeContent(discoveryReport *model.ProjectReport) []string {
	lines := []string{"Ready to inspect bounded local project metadata."}
	if discoveryReport == nil {
		lines = append(lines, "Discovery has not run yet.")
	} else {
		project := discoveryReport.Project
		lines = append(lines, fmt.Sprintf("Latest  %s · %d warning(s)", report.SafeText(project.Name), len(project.Warnings)))
	}
	return append(lines,
		"",
		"Next",
		"  /diagnose   Inspect the selected project",
		"  /project    View detected project details",
		"  /help       Review commands and privacy",
		"",
		"Local-only  No scripts, network, AI, installs, or file changes.",
	)
}

func diagnoseContent(state RunState, projectPath string, discoveryReport *model.ProjectReport, runError string, width int) []string {
	if state == StateRunning || state == StateWaiting {
		return []string{
			"Inspecting project metadata",
			"  " + middleElide(report.SafeText(projectPath), max(16, width-2)),
			"",
			"Scope",
			"  Allowlisted local metadata only; no project scripts are running.",
			"",
			"Ctrl+C cancels this discovery run safely.",
		}
	}
	if state == StateFailed || state == StateCancelled || state == StateTimedOut || state == StateSkipped {
		message := runError
		if message == "" {
			message = defaultRunMessage(state)
		}
		return []string{
			report.SafeText(message),
			"",
			"Next",
			"  /diagnose   Try safe discovery again",
			"  /help       Review discovery scope",
		}
	}
	if discoveryReport == nil {
		return []string{
			"No discovery result yet.",
			"",
			"Next",
			"  /diagnose   Inspect the selected project",
		}
	}

	project := discoveryReport.Project
	return []string{
		"Safe metadata inspection complete.",
		fmt.Sprintf("Detected  %d language(s) · %d runtime(s) · %d package manager(s)", len(project.Languages), len(project.Runtimes), len(project.PackageManagers)),
		fmt.Sprintf("          %d framework(s) · %d workspace(s) · %d warning(s)", len(project.Frameworks), len(project.Workspaces), len(project.Warnings)),
		"",
		"Next",
		"  /project    Review detected project details",
		"  /warnings   Review warnings and recommendations",
		"  /export     Preview the report",
	}
}

func projectContent(discoveryReport *model.ProjectReport, width int) []string {
	if discoveryReport == nil {
		return []string{
			"No project has been inspected yet.",
			"",
			"Next",
			"  /diagnose   Inspect safe project metadata",
		}
	}
	project := discoveryReport.Project
	lines := []string{
		"Identity",
		"  " + report.SafeText(project.Name),
	}
	lines = appendDetectionGroup(lines, "Languages", detectionNames(project.Languages))
	lines = appendDetectionGroup(lines, "Runtimes", runtimeNames(project.Runtimes))
	lines = appendDetectionGroup(lines, "Package managers", packageManagerNames(project.PackageManagers))
	lines = appendDetectionGroup(lines, "Frameworks", detectionNames(project.Frameworks))
	lines = appendDetectionGroup(lines, "Workspaces", workspaceNames(project.Workspaces))

	if missing := absentCategories(project); len(missing) > 0 {
		lines = append(lines, "", "Not detected  "+strings.Join(missing, ", "))
	}
	if len(project.RelevantFiles) > 0 {
		lines = append(lines, "", "Metadata")
		limit := min(4, len(project.RelevantFiles))
		for _, file := range project.RelevantFiles[:limit] {
			lines = append(lines, "  "+middleElide(report.SafeText(file), max(12, width-2)))
		}
		if len(project.RelevantFiles) > limit {
			lines = append(lines, fmt.Sprintf("  +%d more file(s)", len(project.RelevantFiles)-limit))
		}
	}
	return lines
}

func warningsContent(discoveryReport *model.ProjectReport, selected int, detail bool, width int) []string {
	if discoveryReport == nil {
		return []string{
			"No discovery result is available.",
			"",
			"Next",
			"  /diagnose   Inspect safe project metadata",
		}
	}
	warnings := discoveryReport.Project.Warnings
	if len(warnings) == 0 {
		return []string{
			"No warnings were found in the current discovery report.",
			"",
			"Next",
			"  /project    Review detected project details",
		}
	}
	selected = min(max(0, selected), len(warnings)-1)
	if detail {
		warning := warnings[selected]
		lines := []string{
			"Warning detail",
			"",
			fmt.Sprintf("%s  %s", report.SafeText(warning.Code), report.SafeText(warning.Message)),
			"",
			"Evidence",
		}
		if len(warning.Evidence) == 0 {
			lines = append(lines, "  No additional evidence was recorded.")
		} else {
			for _, evidence := range warning.Evidence {
				value := report.SafeText(evidence.Detail)
				if evidence.Path != "" {
					value = middleElide(report.SafeText(evidence.Path), max(12, width/2)) + ": " + value
				}
				lines = append(lines, "  "+value)
			}
		}
		return append(lines,
			"",
			"Recommendation",
			"  Review the referenced metadata before choosing a project action.",
			"  DebugDoc does not apply fixes in this phase.",
		)
	}

	lines := []string{fmt.Sprintf("%d item(s) need review. Select one for details.", len(warnings)), ""}
	rowWidth := min(74, max(1, width-2))
	for index, warning := range warnings {
		marker := "  "
		if index == selected {
			marker = "> "
		}
		label := fmt.Sprintf("%-20s %s", compactWarningCode(report.SafeText(warning.Code)), report.SafeText(warning.Message))
		lines = append(lines, marker+endElide(label, rowWidth))
	}
	return lines
}

func exportContent(discoveryReport *model.ProjectReport, width int) []string {
	if discoveryReport == nil {
		return []string{
			"No report is available to preview.",
			"",
			"Next",
			"  /diagnose   Inspect a project first",
			"",
			"For automation: debugdoc diagnose --path . --format text|json",
		}
	}

	project := discoveryReport.Project
	lines := []string{
		"Preview only — no file is written and the clipboard is not used.",
		fmt.Sprintf("Project  %s   Schema  %s", report.SafeText(project.Name), report.SafeText(discoveryReport.SchemaVersion)),
	}
	lines = appendPreviewGroup(lines, "Languages", detectionNames(project.Languages))
	lines = appendPreviewGroup(lines, "Runtimes", runtimeNames(project.Runtimes))
	lines = appendPreviewGroup(lines, "Package managers", packageManagerNames(project.PackageManagers))
	lines = appendPreviewGroup(lines, "Frameworks", detectionNames(project.Frameworks))
	lines = appendPreviewGroup(lines, "Workspaces", workspaceNames(project.Workspaces))
	if missing := absentCategories(project); len(missing) > 0 {
		lines = append(lines, "", "Not detected  "+strings.Join(missing, ", "))
	}
	if len(project.Warnings) > 0 {
		lines = append(lines, "", fmt.Sprintf("Warnings  %d", len(project.Warnings)))
		for _, warning := range project.Warnings {
			value := fmt.Sprintf("  %s  %s", compactWarningCode(report.SafeText(warning.Code)), report.SafeText(warning.Message))
			lines = append(lines, endElide(value, max(12, width)))
		}
	}
	if len(project.RelevantFiles) > 0 {
		lines = append(lines, "", "Metadata")
		for _, file := range project.RelevantFiles {
			lines = append(lines, "  "+middleElide(report.SafeText(file), max(12, width-2)))
		}
	}
	return lines
}

func helpContent() []string {
	return []string{
		"Commands",
		"  /diagnose   Inspect project metadata",
		"  /project    View detected project details",
		"  /warnings   Review warnings",
		"  /export     Preview the report",
		"  /help       Show help",
		"  /quit       Exit DebugDoc",
		"",
		"Keyboard",
		"  Up/Down navigate   Enter select   Esc back",
		"  PgUp/PgDn page     End latest     Ctrl+L redraw",
		"  Ctrl+C cancel, clear, or quit",
		"",
		"Privacy",
		"  DebugDoc reads bounded local metadata only.",
		"  Arbitrary shell input is never executed.",
	}
}

func defaultRunMessage(state RunState) string {
	switch state {
	case StateCancelled:
		return "Discovery was cancelled."
	case StateTimedOut:
		return "Discovery timed out."
	case StateSkipped:
		return "Discovery was skipped."
	case StateFailed:
		return "Discovery failed."
	case StateReady, StateRunning, StateWaiting, StateOK, StateWarning:
		return "Discovery did not complete."
	default:
		return "Discovery did not complete."
	}
}

func appendDetectionGroup(lines []string, title string, values []string) []string {
	if len(values) == 0 {
		return lines
	}
	return append(lines, "", title, "  "+strings.Join(values, ", "))
}

func appendPreviewGroup(lines []string, title string, values []string) []string {
	if len(values) == 0 {
		return lines
	}
	return append(lines, "", title+"  "+strings.Join(values, ", "))
}

func compactWarningCode(value string) string {
	const limit = 20
	if len([]rune(value)) <= limit {
		return value
	}
	return endElide(value, limit)
}

func detectionNames(detections []model.Detection) []string {
	values := make([]string, 0, len(detections))
	for _, detection := range detections {
		values = append(values, fmt.Sprintf("%s (%s)", report.SafeText(detection.Name), report.SafeText(string(detection.Confidence))))
	}
	return values
}

func runtimeNames(detections []model.RuntimeDetection) []string {
	values := make([]string, 0, len(detections))
	for _, detection := range detections {
		value := report.SafeText(detection.Name)
		if detection.Requirement != "" {
			value += " " + report.SafeText(detection.Requirement)
		}
		values = append(values, value)
	}
	return values
}

func packageManagerNames(detections []model.PackageManagerDetection) []string {
	values := make([]string, 0, len(detections))
	for _, detection := range detections {
		value := report.SafeText(detection.Name)
		if detection.DeclaredVersion != "" {
			value += " " + report.SafeText(detection.DeclaredVersion)
		}
		values = append(values, value)
	}
	return values
}

func workspaceNames(detections []model.WorkspaceDetection) []string {
	values := make([]string, 0, len(detections))
	for _, detection := range detections {
		values = append(values, report.SafeText(detection.Name))
	}
	return values
}

func absentCategories(project model.ProjectSummary) []string {
	var missing []string
	if len(project.Languages) == 0 {
		missing = append(missing, "languages")
	}
	if len(project.Runtimes) == 0 {
		missing = append(missing, "runtimes")
	}
	if len(project.PackageManagers) == 0 {
		missing = append(missing, "package managers")
	}
	if len(project.Frameworks) == 0 {
		missing = append(missing, "frameworks")
	}
	if len(project.Workspaces) == 0 {
		missing = append(missing, "workspaces")
	}
	return missing
}
