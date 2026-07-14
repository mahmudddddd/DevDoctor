package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	projectmodel "github.com/mahmudddddd/DevDoctor/internal/model"
)

func TestSlashCommandSelectionRunsDiscoveryAndShowsResults(t *testing.T) {
	t.Parallel()
	calls := 0
	model := NewModel(Config{
		ProjectPath: "/project",
		Discover: func(_ context.Context, path string) (projectmodel.ProjectReport, error) {
			calls++
			if path != "/project" {
				t.Fatalf("path = %q, want /project", path)
			}
			return sampleReport(), nil
		},
	})
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/dia")})
	if !model.paletteOpen || len(model.paletteMatches) != 1 {
		t.Fatalf("palette open/matches = %v/%d, want true/1", model.paletteOpen, len(model.paletteMatches))
	}

	var command tea.Cmd
	model, command = updateModelWithCommand(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if model.state != StateRunning || model.screen != ScreenDiagnose || command == nil {
		t.Fatalf("after selection state/screen/cmd = %s/%s/%v", model.state, model.screen, command != nil)
	}
	message := command()
	model = updateModel(t, model, message)
	if calls != 1 || model.state != StateOK || model.report == nil {
		t.Fatalf("calls/state/report = %d/%s/%v, want 1/OK/non-nil", calls, model.state, model.report != nil)
	}
}

func TestExactTypedCommandOverridesDescriptionMatch(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project"})
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/project")})
	if len(model.paletteMatches) < 2 || model.paletteMatches[0].ID != CommandProject {
		t.Fatalf("name match was not ranked before description match: %#v", model.paletteMatches)
	}
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if model.screen != ScreenProject {
		t.Fatalf("exact /project selected %s, want Project", model.screen)
	}
}

func TestRequiredCommandsSelectTheirViews(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		command CommandID
		screen  Screen
	}{
		{command: CommandProject, screen: ScreenProject},
		{command: CommandWarnings, screen: ScreenWarnings},
		{command: CommandExport, screen: ScreenExport},
		{command: CommandHelp, screen: ScreenHelp},
		{command: CommandClear, screen: ScreenHome},
	} {
		model := NewModel(Config{ProjectPath: "/project"})
		model.report = reportPointer(sampleReport())
		updated, _ := model.applyAction(Action{ID: test.command})
		result := updated.(Model)
		if result.screen != test.screen {
			t.Fatalf("command %s selected %s, want %s", test.command, result.screen, test.screen)
		}
	}
}

func TestArbitraryInputIsRejectedWithoutDiscovery(t *testing.T) {
	t.Parallel()
	calls := 0
	model := NewModel(Config{
		ProjectPath: "/project",
		Discover: func(context.Context, string) (projectmodel.ProjectReport, error) {
			calls++
			return sampleReport(), nil
		},
	})
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("go test ./...")})
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if calls != 0 {
		t.Fatalf("discovery calls = %d, want 0", calls)
	}
	if !strings.Contains(model.validation, "slash commands only") {
		t.Fatalf("validation = %q, want safe rejection", model.validation)
	}
}

func TestDiscoveryFailureAndCancellationAreExplicit(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		err   error
		state RunState
	}{
		{name: "failure", err: errors.New("bad\x1b[31mproject"), state: StateFailed},
		{name: "cancelled", err: context.Canceled, state: StateCancelled},
		{name: "timed out", err: context.DeadlineExceeded, state: StateTimedOut},
	} {
		t.Run(test.name, func(t *testing.T) {
			model := NewModel(Config{ProjectPath: "/project", Discover: func(context.Context, string) (projectmodel.ProjectReport, error) {
				return projectmodel.ProjectReport{}, test.err
			}})
			started, command := model.startDiscovery()
			model = started.(Model)
			model = updateModel(t, model, command())
			if model.state != test.state {
				t.Fatalf("state = %s, want %s", model.state, test.state)
			}
			if strings.Contains(model.View(), "\x1b[31mproject") {
				t.Fatalf("View() contains raw terminal control sequence")
			}
		})
	}
}

func TestResponsiveRenderingDoesNotOverflow(t *testing.T) {
	t.Parallel()
	for _, size := range []struct {
		width  int
		height int
	}{
		{width: 80, height: 24},
		{width: 100, height: 30},
		{width: 120, height: 40},
		{width: 79, height: 23},
	} {
		model := NewModel(Config{ProjectPath: `C:\Users\a-very-long-user-name\projects\a-very-long-project-name\workspace`, ASCII: true})
		model = updateModel(t, model, tea.WindowSizeMsg{Width: size.width, Height: size.height})
		view := model.View()
		lines := strings.Split(view, "\n")
		if len(lines) > size.height {
			t.Fatalf("%dx%d rendered %d lines", size.width, size.height, len(lines))
		}
		for index, line := range lines {
			if width := lipgloss.Width(line); width > size.width {
				t.Fatalf("%dx%d line %d width = %d: %q", size.width, size.height, index, width, line)
			}
		}
	}
}

func TestRunningPaletteKeepsCancellationAndHidesUnavailableActions(t *testing.T) {
	t.Parallel()
	cancelled := false
	model := NewModel(Config{ProjectPath: "/project"})
	model.state = StateRunning
	model.runCancel = func() { cancelled = true }
	model.draft = "/"
	model.syncPalette()
	if len(model.paletteMatches) != 1 || model.paletteMatches[0].ID != CommandQuit {
		t.Fatalf("running palette = %#v, want only /quit", model.paletteMatches)
	}
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyCtrlC})
	if !cancelled || model.paletteOpen {
		t.Fatalf("Ctrl+C cancelled/open = %v/%v, want true/false", cancelled, model.paletteOpen)
	}
}

func TestEscapeNavigatesBackAndComposerSanitizesControls(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project"})
	model.navigateTo(ScreenHelp)
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEsc})
	if model.screen != ScreenHome {
		t.Fatalf("Esc screen = %s, want Home", model.screen)
	}
	model.draft = "\x1b[31munsafe"
	view := model.View()
	if strings.Contains(view, "\x1b[31munsafe") || !strings.Contains(view, "unsafe") {
		t.Fatalf("composer control text was not sanitized: %q", view)
	}
}

func TestMinimumLayoutAndPaletteWidthMatchContract(t *testing.T) {
	t.Parallel()
	layout := ComputeLayout(80, 24)
	if layout.HeaderHeight != 2 || layout.ComposerHeight != 1 || layout.FooterHeight != 1 || layout.DividerHeight != 1 || layout.ViewportHeight != 19 {
		t.Fatalf("80x24 layout = %#v", layout)
	}
	model := NewModel(Config{ProjectPath: "/project"})
	model = updateModel(t, model, tea.WindowSizeMsg{Width: 120, Height: 40})
	model.draft = "/"
	model.syncPalette()
	for _, line := range model.overlayPalette(nil) {
		if strings.Contains(line, "/") && lipgloss.Width(line) > 64 {
			t.Fatalf("palette line width = %d, want <= 64: %q", lipgloss.Width(line), line)
		}
	}
}

func TestHeaderShowsConfiguredProjectBeforeDiscovery(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: `C:\workspace\fixture`, ASCII: true})
	view := model.View()
	if strings.Contains(view, "No project selected") || !strings.Contains(view, "fixture") {
		t.Fatalf("initial header does not show configured project: %q", view)
	}
}

func TestNoColorAndASCIIFallback(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project", Color: false, ASCII: true, Flat: true})
	model.report = reportPointer(sampleReport())
	model.state = StateOK
	model.showScreen(ScreenDiagnose)
	view := model.View()
	if strings.Contains(view, "\x1b[") {
		t.Fatalf("View() contains ANSI escapes in no-color mode: %q", view)
	}
	if !strings.Contains(view, "[OK]") || !strings.Contains(view, "> ") {
		t.Fatalf("View() lacks ASCII status or prompt: %q", view)
	}
}

func TestHelpFitsAtMinimumSizeWithPersistentComposerAndOneRowFooter(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: `C:\DevDoctor`, ASCII: true})
	model.navigateTo(ScreenHelp)
	model = updateModel(t, model, tea.WindowSizeMsg{Width: 80, Height: 24})

	if model.viewport.Overflowing() {
		t.Fatal("Help overflows the 80x24 viewport")
	}
	view := model.View()
	lines := strings.Split(view, "\n")
	if len(lines) != 24 {
		t.Fatalf("rendered lines = %d, want 24", len(lines))
	}
	if !strings.Contains(view, "> Type / for commands") {
		t.Fatalf("minimum view does not retain the composer: %q", view)
	}
	if model.layout.FooterHeight != 1 || !strings.Contains(lines[len(lines)-1], "/ commands") {
		t.Fatalf("footer = %q, want one contextual row", lines[len(lines)-1])
	}
	if strings.Contains(view, "Rows ") {
		t.Fatalf("Help shows a scroll indicator even though it fits: %q", view)
	}
}

func TestComposerRemainsVisibleAtSupportedSizes(t *testing.T) {
	t.Parallel()
	for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 24}, {Width: 100, Height: 30}, {Width: 120, Height: 40}} {
		model := NewModel(Config{ProjectPath: "/project", ASCII: true})
		model = updateModel(t, model, size)
		lines := strings.Split(model.View(), "\n")
		composerIndex := size.Height - model.layout.FooterHeight - model.layout.ComposerHeight
		if composerIndex < 0 || !strings.Contains(lines[composerIndex], "> Type / for commands") {
			t.Fatalf("%dx%d composer line = %q", size.Width, size.Height, lines[composerIndex])
		}
	}
}

func TestHeaderAvoidsDuplicateBrandingAndReportRows(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: `C:\workspace\DevDoctor`, ASCII: true})
	view := model.View()
	if strings.Contains(view, "DevDoctor  DevDoctor") || strings.Contains(view, "View:") || strings.Contains(view, "Root:") {
		t.Fatalf("header retained duplicate/report-like rows: %q", view)
	}
	if strings.Count(strings.Split(view, "\n")[0], "DevDoctor") != 1 {
		t.Fatalf("product identity is duplicated: %q", strings.Split(view, "\n")[0])
	}
	separatorLines := 0
	for _, line := range strings.Split(view, "\n") {
		if strings.Trim(line, "-") == "" && line != "" {
			separatorLines++
		}
	}
	if separatorLines != 1 {
		t.Fatalf("separator lines = %d, want exactly one", separatorLines)
	}
}

func TestHeaderContainsEachViewNameOnlyOnce(t *testing.T) {
	t.Parallel()
	for _, screen := range []Screen{ScreenHome, ScreenDiagnose, ScreenProject, ScreenWarnings, ScreenExport, ScreenHelp} {
		model := NewModel(Config{ProjectPath: `C:\DevDoctor`, ASCII: true, Flat: true})
		model.report = reportPointer(sampleReport())
		model.showScreen(screen)
		view := model.View()
		if count := strings.Count(view, screen.String()); count != 1 {
			t.Fatalf("%s appears %d times, want once:\n%s", screen, count, view)
		}
	}
}

func TestReadableWidthAndTwoRowBottomArea(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: `C:\DevDoctor`, ASCII: true, Flat: true})
	model = updateModel(t, model, tea.WindowSizeMsg{Width: 120, Height: 40})
	if model.layout.ContentWidth != 88 {
		t.Fatalf("content width = %d, want 88", model.layout.ContentWidth)
	}
	if model.layout.ComposerHeight != 1 || model.layout.FooterHeight != 1 || model.layout.DividerHeight != 1 {
		t.Fatalf("bottom layout = %#v, want one divider plus two rows", model.layout)
	}
	view := model.View()
	if strings.Contains(view, "Registered DevDoctor actions only") {
		t.Fatalf("idle composer repeats safety guidance: %q", view)
	}
	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > 88 {
			t.Fatalf("line width = %d, want readable width <= 88: %q", lipgloss.Width(line), line)
		}
	}
}

func TestWarningSelectionIsBoundedToContentColumn(t *testing.T) {
	t.Parallel()
	reportValue := sampleReport()
	reportValue.Project.Warnings = []projectmodel.Warning{{Code: "very-long-warning-code", Message: strings.Repeat("warning detail ", 12)}}
	model := NewModel(Config{ProjectPath: "/project", Color: false, ASCII: true, Flat: true})
	model.report = &reportValue
	model.showScreen(ScreenWarnings)
	for _, line := range strings.Split(model.View(), "\n") {
		if strings.HasPrefix(line, "> ") {
			if width := lipgloss.Width(line); width > 76 {
				t.Fatalf("selected warning width = %d, want <= 76: %q", width, line)
			}
			return
		}
	}
	t.Fatal("selected warning row not found")
}

func TestPaletteIsTransientAnchoredAndSelectionIsTextual(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project", Color: false, ASCII: true, Flat: true})
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	view := model.View()
	lines := strings.Split(view, "\n")
	commandsLine := -1
	selectedLine := -1
	for index, line := range lines {
		if strings.TrimSpace(line) == "Commands" {
			commandsLine = index
		}
		if strings.HasPrefix(line, "> /diagnose") {
			selectedLine = index
		}
	}
	composerIndex := len(lines) - model.layout.FooterHeight - model.layout.ComposerHeight
	if commandsLine < model.layout.HeaderHeight || selectedLine < 0 || selectedLine >= composerIndex {
		t.Fatalf("palette not anchored above composer: commands=%d selected=%d composer=%d\n%s", commandsLine, selectedLine, composerIndex, view)
	}
	if !strings.Contains(view, "  /project") {
		t.Fatalf("unselected palette row missing: %q", view)
	}

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEsc})
	if model.paletteOpen || model.draft != "/" || strings.Contains(model.View(), "\nCommands\n") {
		t.Fatalf("Esc did not transiently close palette while preserving draft: open=%v draft=%q", model.paletteOpen, model.draft)
	}
}

func TestPaletteSelectionRemainsVisuallyDistinctWithColorEnabled(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project", Color: true})
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	view := model.View()
	if !strings.Contains(view, "> /diagnose") || !strings.Contains(view, "  /project") {
		t.Fatalf("selected and unselected rows are not textually distinct: %q", view)
	}
}

func TestScrollPositionAppearsOnlyForOverflow(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project", ASCII: true})
	if strings.Contains(model.View(), "Rows ") {
		t.Fatal("short Home content shows a position indicator")
	}
	content := make([]string, 40)
	for index := range content {
		content[index] = fmt.Sprintf("row %d", index)
	}
	model.viewport.SetContent(content)
	if !strings.Contains(model.View(), "Rows ") {
		t.Fatal("overflowing content does not show a position indicator")
	}
}

func TestLongProjectPathIsMiddleElided(t *testing.T) {
	t.Parallel()
	path := `C:\Users\a-very-long-user-name\projects\a-very-long-workspace-name\packages\a-very-long-project-name`
	model := NewModel(Config{ProjectPath: path, ASCII: true})
	model = updateModel(t, model, tea.WindowSizeMsg{Width: 80, Height: 24})
	header := strings.Split(model.View(), "\n")[1]
	if lipgloss.Width(header) > 80 || !strings.Contains(header, "...") || strings.Contains(header, path) {
		t.Fatalf("long path was not cleanly elided: %q", header)
	}
}

func TestWarningsSelectionAndDetailKeepFocusVisible(t *testing.T) {
	t.Parallel()
	reportValue := sampleReport()
	reportValue.Project.Warnings = []projectmodel.Warning{
		{Code: "ambiguous-runtime", Message: "More than one runtime marker was found."},
		{Code: "metadata-skipped", Message: "A denied metadata path was skipped."},
	}
	model := NewModel(Config{ProjectPath: "/project", ASCII: true, Flat: true})
	model.report = &reportValue
	model.navigateTo(ScreenWarnings)
	if model.focus != FocusWarnings || !strings.Contains(model.View(), "> ambiguous-runtime") {
		t.Fatalf("initial warning focus/selection is not visible: %q", model.View())
	}
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyDown})
	if model.warningSelected != 1 || !strings.Contains(model.View(), "> metadata-skipped") {
		t.Fatalf("warning selection did not move visibly: %q", model.View())
	}
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	if !model.warningDetail || model.focus != FocusWarningDetail || !strings.Contains(model.View(), "Warning detail") {
		t.Fatalf("Enter did not open warning detail: %q", model.View())
	}
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEsc})
	if model.warningDetail || model.focus != FocusWarnings || !strings.Contains(model.View(), "> metadata-skipped") {
		t.Fatalf("Esc did not restore warning-list focus: %q", model.View())
	}
}

func TestEveryPrimaryViewStaysWithinSupportedFrames(t *testing.T) {
	t.Parallel()
	reportValue := sampleReport()
	reportValue.Project.Warnings = []projectmodel.Warning{{Code: "warning", Message: strings.Repeat("long warning ", 12)}}
	for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 24}, {Width: 100, Height: 30}, {Width: 120, Height: 40}} {
		for _, screen := range []Screen{ScreenHome, ScreenDiagnose, ScreenProject, ScreenWarnings, ScreenExport, ScreenHelp} {
			model := NewModel(Config{ProjectPath: `C:\a\long\project\path`, ASCII: true})
			model.report = &reportValue
			model.state = StateWarning
			model = updateModel(t, model, size)
			model.showScreen(screen)
			lines := strings.Split(model.View(), "\n")
			if len(lines) > size.Height {
				t.Fatalf("%s at %dx%d rendered %d lines", screen, size.Width, size.Height, len(lines))
			}
			for index, line := range lines {
				if width := lipgloss.Width(line); width > size.Width {
					t.Fatalf("%s at %dx%d line %d width=%d: %q", screen, size.Width, size.Height, index, width, line)
				}
			}
		}
	}
}

func TestCtrlLStillRequestsRedraw(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project"})
	_, command := updateModelWithCommand(t, model, tea.KeyMsg{Type: tea.KeyCtrlL})
	if command == nil {
		t.Fatal("Ctrl+L did not return a redraw command")
	}
}

func updateModel(t *testing.T, model Model, message tea.Msg) Model {
	t.Helper()
	updated, _ := model.Update(message)
	result, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update() returned %T, want tui.Model", updated)
	}
	return result
}

func updateModelWithCommand(t *testing.T, model Model, message tea.Msg) (Model, tea.Cmd) {
	t.Helper()
	updated, command := model.Update(message)
	result, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update() returned %T, want tui.Model", updated)
	}
	return result, command
}

func sampleReport() projectmodel.ProjectReport {
	return projectmodel.ProjectReport{
		SchemaVersion: projectmodel.ReportSchemaVersion,
		ToolVersion:   "test",
		Project: projectmodel.ProjectSummary{
			Root: "/project",
			Name: "fixture",
			Languages: []projectmodel.Detection{{
				ID: "javascript", Name: "JavaScript", Confidence: projectmodel.ConfidenceHigh,
				Evidence: []projectmodel.Evidence{{Kind: "manifest", Path: "package.json", Detail: "package manifest"}},
			}},
		},
	}
}

func reportPointer(value projectmodel.ProjectReport) *projectmodel.ProjectReport {
	return &value
}
