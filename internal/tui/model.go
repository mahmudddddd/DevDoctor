package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"

	projectmodel "github.com/mahmudddddd/DevDoctor/internal/model"
	"github.com/mahmudddddd/DevDoctor/internal/report"
)

type discoveryFinishedMsg struct {
	generation uint64
	report     projectmodel.ProjectReport
	err        error
}

// Model is the Bubble Tea state for the DebugDoc interactive shell.
type Model struct {
	parentContext  context.Context
	discover       DiscoverFunc
	projectPath    string
	report         *projectmodel.ProjectReport
	screen         Screen
	previousScreen Screen
	hasPrevious    bool
	state          RunState
	runError       string
	runCancel      context.CancelFunc
	generation     uint64
	quitPending    bool
	quitArmed      bool

	layout       Layout
	capabilities Capabilities
	styles       styles
	viewport     Viewport
	animation    animationState
	transient    transientFooter

	draft           string
	validation      string
	paletteOpen     bool
	paletteMatches  []Command
	paletteSelected int
	focus           Focus
	previousFocus   Focus
	warningSelected int
	warningDetail   bool
}

// NewModel constructs a deterministic initial Home screen.
func NewModel(config Config) Model {
	ctx := config.Context
	if ctx == nil {
		ctx = context.Background()
	}
	motionEnabled := !config.ReduceMotion && !config.Flat
	capabilities := Capabilities{Color: config.Color, ASCII: config.ASCII, Flat: config.Flat, Motion: motionEnabled}
	model := Model{
		parentContext: ctx,
		discover:      config.Discover,
		projectPath:   config.ProjectPath,
		screen:        ScreenHome,
		state:         StateReady,
		capabilities:  capabilities,
		styles:        newStyles(capabilities),
		viewport:      newViewport(),
		animation:     newAnimationState(motionEnabled, config.ASCII),
		focus:         FocusComposer,
	}
	model.resize(minimumWidth, minimumHeight)
	return model
}

// Init satisfies tea.Model. Discovery starts only after an explicit safe action.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update applies modal-first input routing and typed discovery messages.
func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.resize(message.Width, message.Height)
		return m, nil
	case discoveryFinishedMsg:
		return m.finishDiscovery(message)
	case animationTickMsg:
		if message.generation != m.generation || m.state != StateRunning || !m.animation.enabled {
			return m, nil
		}
		m.animation.advance()
		return m, m.animation.tick(m.generation)
	case transientExpiredMsg:
		m.transient.expire(message)
		return m, nil
	case tea.KeyMsg:
		return m.updateKey(message)
	default:
		return m, nil
	}
}

func (m Model) updateKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyName := key.String()
	if keyName != keyCancel {
		m.quitArmed = false
	}

	if keyName == keyCancel && m.state == StateRunning && m.runCancel != nil {
		m.closePalette()
		m.runCancel()
		m.validation = ""
		return m, m.transient.set("Diagnosis cancellation requested", false)
	}
	if m.paletteOpen {
		return m.updatePalette(key)
	}
	if m.screen == ScreenWarnings && m.warningDetail && keyName == keyEscape {
		m.warningDetail = false
		m.focus = FocusWarnings
		m.refreshContent()
		m.viewport.JumpToStart()
		return m, m.transient.set("Returned to warnings", false)
	}
	if m.screen == ScreenWarnings && !m.warningDetail && m.warningCount() > 0 && m.draft == "" {
		switch keyName {
		case keyUp:
			m.moveWarningSelection(-1)
			return m, nil
		case keyDown:
			m.moveWarningSelection(1)
			return m, nil
		case keyEnter:
			m.warningDetail = true
			m.focus = FocusWarningDetail
			m.refreshContent()
			m.viewport.JumpToStart()
			return m, m.transient.set("Opened warning details", false)
		}
	}

	switch keyName {
	case keyCancel:
		if m.draft != "" {
			m.draft = ""
			m.validation = ""
			return m, m.transient.set("Input cleared", false)
		}
		if m.quitArmed {
			return m, tea.Quit
		}
		m.quitArmed = true
		m.validation = "Press Ctrl+C again or type /quit to leave DebugDoc."
		return m, nil
	case keyRedraw:
		return m, tea.ClearScreen
	case keyHelp:
		if m.draft == "" && m.state != StateRunning {
			m.navigateTo(ScreenHelp)
			return m, m.transient.set("Opened Help", false)
		}
	case keyUp:
		m.viewport.ScrollUp(1)
		return m, nil
	case keyDown:
		m.viewport.ScrollDown(1)
		return m, nil
	case keyPageUp:
		m.viewport.PageUp()
		return m, nil
	case keyPageDown:
		m.viewport.PageDown()
		return m, nil
	case keyHome:
		m.viewport.JumpToStart()
		return m, nil
	case keyEnd:
		m.viewport.JumpToLatest()
		return m, nil
	case keyEnter:
		return m.submitDraft()
	case keyBack, keyBackAlt:
		m.focus = FocusComposer
		m.draft = trimLastRune(m.draft)
		m.validation = ""
		m.transient.clear()
		m.syncPalette()
		return m, nil
	case keyEscape:
		if m.draft != "" {
			m.draft = ""
			m.validation = ""
			m.transient.clear()
			return m, nil
		}
		if m.hasPrevious {
			previous := m.previousScreen
			m.hasPrevious = false
			m.showScreen(previous)
		}
		return m, nil
	}

	if draft, appended := appendPrintableInput(m.draft, key); appended {
		m.focus = FocusComposer
		m.draft = draft
		m.validation = ""
		m.transient.clear()
		m.syncPalette()
	}
	return m, nil
}

func (m Model) updatePalette(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case keyEscape:
		m.closePalette()
		return m, nil
	case keyUp:
		if len(m.paletteMatches) > 0 {
			m.paletteSelected = (m.paletteSelected - 1 + len(m.paletteMatches)) % len(m.paletteMatches)
		}
		return m, nil
	case keyDown:
		if len(m.paletteMatches) > 0 {
			m.paletteSelected = (m.paletteSelected + 1) % len(m.paletteMatches)
		}
		return m, nil
	case keyEnter:
		if action, err := ParseAction(m.draft); err == nil {
			m.draft = ""
			m.validation = ""
			m.closePalette()
			return m.applyAction(action)
		}
		if len(m.paletteMatches) == 0 {
			m.validation = "No matching DebugDoc command. Press Esc to keep editing."
			return m, nil
		}
		m.draft = m.paletteMatches[m.paletteSelected].Name
		m.closePalette()
		return m.submitDraft()
	case keyBack, keyBackAlt:
		m.draft = trimLastRune(m.draft)
		m.validation = ""
		m.transient.clear()
		m.syncPalette()
		return m, nil
	}
	if draft, appended := appendPrintableInput(m.draft, key); appended {
		m.draft = draft
		m.validation = ""
		m.transient.clear()
		m.syncPalette()
	}
	return m, nil
}

func (m Model) submitDraft() (tea.Model, tea.Cmd) {
	action, err := ParseAction(m.draft)
	if err != nil {
		m.transient.clear()
		m.validation = err.Error() + "."
		return m, nil
	}
	m.draft = ""
	m.validation = ""
	m.closePalette()
	return m.applyAction(action)
}

func (m Model) applyAction(action Action) (tea.Model, tea.Cmd) {
	if !IsActionAvailable(action, m.commandContext()) {
		m.transient.clear()
		m.validation = "That DebugDoc action is unavailable while discovery is running."
		return m, nil
	}
	m.validation = ""
	switch action.ID {
	case CommandDiagnose:
		return m.startDiscovery()
	case CommandProject:
		m.navigateTo(ScreenProject)
		return m, m.transient.set("Opened Project", false)
	case CommandWarnings:
		m.navigateTo(ScreenWarnings)
		return m, m.transient.set("Opened Warnings", false)
	case CommandExport:
		m.navigateTo(ScreenExport)
		return m, m.transient.set("Opened Export", false)
	case CommandHelp:
		m.navigateTo(ScreenHelp)
		return m, m.transient.set("Opened Help", false)
	case CommandClear:
		m.runError = ""
		m.hasPrevious = false
		m.showScreen(ScreenHome)
		return m, m.transient.set("View cleared", false)
	case CommandQuit:
		if m.state == StateRunning && m.runCancel != nil {
			m.quitPending = true
			m.runCancel()
			return m, m.transient.set("Cancelling diagnosis before exit", false)
		}
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m Model) startDiscovery() (tea.Model, tea.Cmd) {
	if m.state == StateRunning {
		m.transient.clear()
		m.validation = "A discovery run is already active."
		return m, nil
	}
	if m.discover == nil {
		m.state = StateFailed
		m.runError = "Discovery is unavailable."
		m.hasPrevious = false
		m.showScreen(ScreenDiagnose)
		return m, m.transient.set("Diagnosis unavailable", true)
	}
	ctx, cancel := context.WithCancel(m.parentContext)
	m.runCancel = cancel
	m.generation++
	generation := m.generation
	discover := m.discover
	projectPath := m.projectPath
	m.state = StateRunning
	m.runError = ""
	m.validation = ""
	m.hasPrevious = false
	m.animation.frame = 0
	m.showScreen(ScreenDiagnose)
	discoveryCommand := func() tea.Msg {
		discoveryReport, err := discover(ctx, projectPath)
		return discoveryFinishedMsg{generation: generation, report: discoveryReport, err: err}
	}
	return m, tea.Batch(
		discoveryCommand,
		m.animation.tick(generation),
		m.transient.set("Diagnosis started", false),
	)
}

func (m Model) finishDiscovery(message discoveryFinishedMsg) (tea.Model, tea.Cmd) {
	if message.generation != m.generation {
		return m, nil
	}
	if m.runCancel != nil {
		m.runCancel()
		m.runCancel = nil
	}
	if message.err != nil {
		switch {
		case errors.Is(message.err, context.Canceled):
			m.state = StateCancelled
			m.runError = "Discovery was cancelled."
		case errors.Is(message.err, context.DeadlineExceeded):
			m.state = StateTimedOut
			m.runError = "Discovery timed out."
		default:
			m.state = StateFailed
			m.runError = report.SafeText(message.err.Error())
		}
	} else {
		discoveryReport := message.report
		m.report = &discoveryReport
		m.warningSelected = 0
		if len(discoveryReport.Project.Warnings) > 0 {
			m.state = StateWarning
		} else {
			m.state = StateOK
		}
	}
	m.hasPrevious = false
	m.showScreen(ScreenDiagnose)
	if m.quitPending {
		return m, tea.Quit
	}

	var messageText string
	persistent := false
	switch m.state {
	case StateOK:
		messageText = "Diagnosis complete"
	case StateWarning:
		messageText = "Diagnosis complete with warnings"
	case StateCancelled:
		messageText = "Diagnosis cancelled"
	case StateTimedOut:
		messageText = "Diagnosis timed out"
		persistent = true
	case StateFailed:
		messageText = "Diagnosis failed"
		persistent = true
	case StateReady, StateRunning, StateWaiting, StateSkipped:
		messageText = "Diagnosis finished"
	default:
		messageText = "Diagnosis finished"
	}
	return m, m.transient.set(messageText, persistent)
}

func (m *Model) showScreen(screen Screen) {
	m.screen = screen
	m.warningDetail = false
	if screen == ScreenWarnings && m.warningCount() > 0 {
		m.focus = FocusWarnings
	} else {
		m.focus = FocusComposer
	}
	m.refreshContent()
	m.viewport.JumpToStart()
}

func (m *Model) navigateTo(screen Screen) {
	if screen != m.screen {
		m.previousScreen = m.screen
		m.hasPrevious = true
	}
	m.showScreen(screen)
}

func (m *Model) refreshContent() {
	m.viewport.SetContent(screenContent(
		m.screen,
		m.state,
		m.projectPath,
		m.report,
		m.runError,
		max(1, m.layout.ContentWidth),
		m.warningSelected,
		m.warningDetail,
	))
}

func (m Model) commandContext() CommandContext {
	return CommandContext{State: m.state, HasReport: m.report != nil}
}

func (m *Model) resize(width, height int) {
	m.layout = ComputeLayout(width, height)
	if m.layout.Limited {
		return
	}
	m.viewport.SetSize(m.layout.ContentWidth, m.layout.ViewportHeight)
	m.refreshContent()
}

func (m *Model) syncPalette() {
	trimmed := strings.TrimLeft(m.draft, " ")
	shouldOpen := strings.HasPrefix(trimmed, "/")
	if !shouldOpen {
		m.closePalette()
		m.paletteMatches = nil
		m.paletteSelected = 0
		return
	}
	if !m.paletteOpen {
		m.previousFocus = m.focus
		m.focus = FocusPalette
		m.paletteOpen = true
	}
	m.paletteMatches = FilterAvailableCommands(trimmed, m.commandContext())
	if m.paletteSelected >= len(m.paletteMatches) {
		m.paletteSelected = max(0, len(m.paletteMatches)-1)
	}
}

func (m *Model) closePalette() {
	if !m.paletteOpen {
		return
	}
	m.paletteOpen = false
	m.focus = m.previousFocus
	if m.focus == FocusPalette {
		m.focus = FocusComposer
	}
}

func (m Model) warningCount() int {
	if m.report == nil {
		return 0
	}
	return len(m.report.Project.Warnings)
}

func (m *Model) moveWarningSelection(delta int) {
	count := m.warningCount()
	if count == 0 {
		return
	}
	m.warningSelected = (m.warningSelected + delta + count) % count
	m.focus = FocusWarnings
	m.refreshContent()
	m.viewport.RevealLogical(2 + m.warningSelected)
}

// View renders one bounded full-screen frame.
func (m Model) View() string {
	if m.layout.Limited {
		return m.renderLimited()
	}

	lines := make([]string, 0, m.layout.Height)
	lines = append(lines, m.renderHeader()...)
	viewportLines := m.renderViewport()
	if m.paletteOpen {
		viewportLines = m.overlayPalette(viewportLines)
	}
	lines = append(lines, fitLines(viewportLines, m.layout.ViewportHeight)...)
	lines = append(lines, m.renderDivider())
	lines = append(lines, m.renderComposer()...)
	lines = append(lines, m.renderFooter()...)
	lines = fitLines(lines, m.layout.Height)
	for index, line := range lines {
		lines[index] = truncateCells(line, max(1, m.layout.Width))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderHeader() []string {
	status := m.styles.state(m.state)
	if m.capabilities.ASCII || m.capabilities.Flat {
		status = m.stateMarker()
	}
	headerWidth := min(64, m.layout.ContentWidth)
	first := alignEdges(m.styles.product.Render("DebugDoc"), status, headerWidth)

	view := m.screen.String()
	separator := " · "
	pathWidth := max(12, m.layout.ContentWidth-runewidth.StringWidth(separator)-runewidth.StringWidth(view))
	path := middleElide(report.SafeText(m.projectPath), pathWidth)
	second := path + separator + view
	return fitLines([]string{first, m.styles.muted.Render(second)}, m.layout.HeaderHeight)
}

func (m Model) renderViewport() []string {
	lines := m.viewport.Lines()
	for index, line := range lines {
		if m.screen == ScreenDiagnose && m.state == StateRunning && m.viewport.offset+index == 0 {
			indicator := m.styles.accent.Render(m.animation.indicator())
			lines[index] = indicator + "  " + line
			continue
		}
		switch {
		case strings.HasPrefix(line, "> "):
			lines[index] = m.styles.warningSelected.Render(padCells(line, min(76, m.layout.ContentWidth)))
		case line == "Warning detail":
			lines[index] = m.styles.title.Render(line)
		case strings.HasPrefix(line, "Latest  "), strings.HasPrefix(line, "Local-only  "), strings.HasPrefix(line, "Not detected  "):
			lines[index] = m.styles.muted.Render(line)
		case isSectionHeading(line):
			lines[index] = m.styles.section.Render(line)
		case strings.HasPrefix(line, "WARNING  "):
			lines[index] = m.styles.warning.Render(line)
		case strings.HasPrefix(line, "FAILED  "):
			lines[index] = m.styles.error.Render(line)
		}
	}
	return lines
}

func (m Model) renderDivider() string {
	character := "─"
	if m.capabilities.ASCII || m.capabilities.Flat {
		character = "-"
	}
	return m.styles.divider.Render(strings.Repeat(character, m.layout.ContentWidth))
}

func (m Model) renderComposer() []string {
	focused := m.focus == FocusComposer || m.focus == FocusPalette
	prompt := "›"
	if m.capabilities.ASCII || m.capabilities.Flat {
		prompt = ">"
	}
	promptStyle := m.styles.composerFocused
	if !focused {
		prompt = "·"
		if m.capabilities.ASCII || m.capabilities.Flat {
			prompt = ":"
		}
		promptStyle = m.styles.composerIdle
	}

	available := max(1, m.layout.ContentWidth-runewidth.StringWidth(prompt)-1)
	draft := tailCells(report.SafeText(m.draft), available)
	if draft == "" {
		draft = m.styles.muted.Render("Type / for commands")
	}
	return fitLines([]string{promptStyle.Render(prompt) + " " + draft}, m.layout.ComposerHeight)
}

func (m Model) renderFooter() []string {
	if m.validation != "" {
		return fitLines([]string{m.styles.error.Render(truncateCells(m.validation, m.layout.ContentWidth))}, m.layout.FooterHeight)
	}
	if m.transient.text != "" {
		return fitLines([]string{m.styles.muted.Render(truncateCells(m.transient.text, m.layout.ContentWidth))}, m.layout.FooterHeight)
	}
	if m.viewport.Overflowing() {
		start, end, total := m.viewport.Position()
		message := fmt.Sprintf("Rows %d–%d of %d   PgUp/PgDn scroll   End latest", start, end, total)
		return fitLines([]string{m.styles.muted.Render(truncateCells(message, m.layout.ContentWidth))}, m.layout.FooterHeight)
	}
	hints := footerHints(m)
	for len(hints) > 0 && lipgloss.Width(renderHints(hints, false)) > m.layout.ContentWidth {
		hints = hints[:len(hints)-1]
	}
	return fitLines([]string{m.styles.muted.Render(renderHints(hints, false))}, m.layout.FooterHeight)
}

func (m Model) overlayPalette(viewportLines []string) []string {
	rows := 2
	if len(m.paletteMatches) > 0 {
		rows = 1 + len(m.paletteMatches)
	}
	geometry := m.layout.PaletteGeometry(rows)
	if geometry.Height == 0 {
		return fitLines(viewportLines, m.layout.ViewportHeight)
	}

	palette := []string{m.styles.paletteTitle.Render("Commands")}
	if len(m.paletteMatches) == 0 {
		palette = append(palette, "  No matching DebugDoc command")
	} else {
		visible := min(len(m.paletteMatches), geometry.Height-1)
		start := 0
		if m.paletteSelected >= visible {
			start = m.paletteSelected - visible + 1
		}
		for index := start; index < start+visible; index++ {
			command := m.paletteMatches[index]
			marker := "  "
			if index == m.paletteSelected {
				marker = "> "
			}
			line := fmt.Sprintf("%s%-10s %s", marker, command.Name, command.Description)
			line = padCells(truncateCells(line, geometry.Width), geometry.Width)
			if index == m.paletteSelected {
				line = m.styles.paletteSelected.Render(line)
			} else {
				line = m.styles.muted.Render(line)
			}
			palette = append(palette, line)
		}
	}
	palette = fitLines(palette, geometry.Height)

	base := fitLines(viewportLines, m.layout.ViewportHeight)
	copy(base[geometry.Top:geometry.Top+geometry.Height], palette)
	return base
}

func alignEdges(left, right string, width int) string {
	if right == "" {
		return truncateCells(left, width)
	}
	space := width - lipgloss.Width(left) - lipgloss.Width(right)
	if space < 1 {
		left = truncateCells(left, max(1, width-lipgloss.Width(right)-1))
		space = max(1, width-lipgloss.Width(left)-lipgloss.Width(right))
	}
	return left + strings.Repeat(" ", space) + right
}

func padCells(value string, width int) string {
	padding := width - lipgloss.Width(value)
	if padding <= 0 {
		return value
	}
	return value + strings.Repeat(" ", padding)
}

func isSectionHeading(value string) bool {
	switch value {
	case "Project", "Status", "Next", "Privacy", "Scope", "Detected", "Identity", "Languages", "Runtimes", "Package managers", "Frameworks", "Workspaces", "Not detected", "Metadata", "Evidence", "Recommendation", "Report", "Commands", "Keyboard":
		return true
	default:
		return false
	}
}

func (m Model) renderLimited() string {
	lines := []string{
		"DebugDoc",
		"",
		"Interactive mode needs at least 80 columns by 24 rows.",
		fmt.Sprintf("Current size: %dx%d.", m.layout.Width, m.layout.Height),
		"Resize the terminal to continue.",
	}
	if m.state == StateRunning {
		lines = append(lines, "Discovery is still running. Press Ctrl+C to cancel.")
	} else {
		lines = append(lines, "Type /quit after resizing, or press Ctrl+C twice to quit.")
	}
	lines = fitLines(lines, max(1, m.layout.Height))
	for index, line := range lines {
		lines[index] = truncateCells(line, max(1, m.layout.Width))
	}
	return strings.Join(lines, "\n")
}

func (m Model) stateMarker() string {
	if m.capabilities.ASCII || m.capabilities.Flat {
		switch m.state {
		case StateOK:
			return "[OK]"
		case StateWarning:
			return "[WARN]"
		case StateFailed:
			return "[FAIL]"
		case StateRunning:
			return "[RUNNING]"
		case StateWaiting:
			return "[WAITING]"
		case StateCancelled:
			return "[CANCELLED]"
		case StateTimedOut:
			return "[TIMED OUT]"
		case StateSkipped:
			return "[SKIPPED]"
		case StateReady:
			return "[READY]"
		default:
			return "[READY]"
		}
	}
	switch m.state {
	case StateOK:
		return "✓"
	case StateWarning:
		return "!"
	case StateFailed:
		return "×"
	case StateRunning:
		return "•"
	case StateWaiting:
		return "○"
	case StateCancelled:
		return "–"
	case StateTimedOut:
		return "×"
	case StateSkipped:
		return "–"
	case StateReady:
		return "○"
	default:
		return "○"
	}
}

func fitLines(lines []string, height int) []string {
	if height < 0 {
		height = 0
	}
	if len(lines) > height {
		result := make([]string, height)
		copy(result, lines[:height])
		return result
	}
	result := make([]string, height)
	copy(result, lines)
	return result
}

func trimLastRune(value string) string {
	if value == "" {
		return value
	}
	_, size := utf8.DecodeLastRuneInString(value)
	return value[:len(value)-size]
}

func truncateCells(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= width {
		return value
	}
	return ansi.Truncate(value, width, "")
}

func endElide(value string, width int) string {
	if runewidth.StringWidth(value) <= width {
		return value
	}
	if width <= 3 {
		return truncateCells(value, width)
	}
	return truncateCells(value, width-3) + "..."
}

func tailCells(value string, width int) string {
	if runewidth.StringWidth(value) <= width {
		return value
	}
	runes := []rune(value)
	used := 0
	start := len(runes)
	for start > 0 {
		candidate := runewidth.RuneWidth(runes[start-1])
		if candidate < 1 {
			candidate = 1
		}
		if used+candidate > width {
			break
		}
		start--
		used += candidate
	}
	return string(runes[start:])
}

func middleElide(value string, width int) string {
	if runewidth.StringWidth(value) <= width || width < 5 {
		return truncateCells(value, width)
	}
	leftWidth := (width - 3) / 2
	rightWidth := width - 3 - leftWidth
	left := truncateCells(value, leftWidth)
	right := tailCells(value, rightWidth)
	return left + "..." + right
}
