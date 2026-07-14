package tui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	product         lipgloss.Style
	title           lipgloss.Style
	section         lipgloss.Style
	muted           lipgloss.Style
	accent          lipgloss.Style
	success         lipgloss.Style
	warning         lipgloss.Style
	error           lipgloss.Style
	divider         lipgloss.Style
	composerFocused lipgloss.Style
	composerIdle    lipgloss.Style
	paletteTitle    lipgloss.Style
	paletteSelected lipgloss.Style
	warningSelected lipgloss.Style
}

func newStyles(capabilities Capabilities) styles {
	plain := lipgloss.NewStyle()
	styles := styles{
		product:         plain,
		title:           plain,
		section:         plain,
		muted:           plain,
		accent:          plain,
		success:         plain,
		warning:         plain,
		error:           plain,
		divider:         plain,
		composerFocused: plain,
		composerIdle:    plain,
		paletteTitle:    plain,
		paletteSelected: plain,
		warningSelected: plain,
	}
	if !capabilities.Color {
		return styles
	}

	const (
		accentColor    = "#4FB6A6"
		mutedColor     = "#87939F"
		successColor   = "#70B886"
		warningColor   = "#D7A94D"
		errorColor     = "#D87878"
		dividerColor   = "#3B5658"
		selectionColor = "#173D3A"
	)

	styles.product = styles.product.Bold(!capabilities.Flat).Foreground(lipgloss.Color(accentColor))
	styles.title = styles.title.Bold(!capabilities.Flat)
	styles.section = styles.section.Bold(!capabilities.Flat).Foreground(lipgloss.Color(accentColor))
	styles.muted = styles.muted.Foreground(lipgloss.Color(mutedColor))
	styles.accent = styles.accent.Foreground(lipgloss.Color(accentColor))
	styles.success = styles.success.Foreground(lipgloss.Color(successColor))
	styles.warning = styles.warning.Foreground(lipgloss.Color(warningColor))
	styles.error = styles.error.Foreground(lipgloss.Color(errorColor))
	styles.divider = styles.divider.Foreground(lipgloss.Color(dividerColor))
	styles.composerFocused = styles.composerFocused.Bold(!capabilities.Flat).Foreground(lipgloss.Color(accentColor))
	styles.composerIdle = styles.composerIdle.Foreground(lipgloss.Color(mutedColor))
	styles.paletteTitle = styles.paletteTitle.Bold(!capabilities.Flat).Foreground(lipgloss.Color(accentColor))
	styles.paletteSelected = styles.paletteSelected.Bold(!capabilities.Flat).Foreground(lipgloss.Color("#DFF7F2")).Background(lipgloss.Color(selectionColor))
	styles.warningSelected = styles.warningSelected.Bold(!capabilities.Flat).Foreground(lipgloss.Color("#FFF2D2")).Background(lipgloss.Color("#4A3617"))
	return styles
}

func (s styles) state(value RunState) string {
	label := string(value)
	switch value {
	case StateOK:
		return s.success.Render(label)
	case StateWarning:
		return s.warning.Render(label)
	case StateFailed, StateCancelled, StateTimedOut:
		return s.error.Render(label)
	case StateRunning, StateWaiting:
		return s.accent.Render(label)
	case StateReady, StateSkipped:
		return label
	default:
		return label
	}
}
