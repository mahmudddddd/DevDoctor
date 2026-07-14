package tui

import "strings"

const (
	keyCancel   = "ctrl+c"
	keyRedraw   = "ctrl+l"
	keyHelp     = "?"
	keyUp       = "up"
	keyDown     = "down"
	keyPageUp   = "pgup"
	keyPageDown = "pgdown"
	keyHome     = "home"
	keyEnd      = "end"
	keyEnter    = "enter"
	keyBack     = "backspace"
	keyBackAlt  = "ctrl+h"
	keyEscape   = "esc"
)

type keyHint struct {
	Key   string
	Label string
}

func keyLabel(key string) string {
	switch key {
	case keyCancel:
		return "Ctrl+C"
	case keyRedraw:
		return "Ctrl+L"
	case keyUp:
		return "Up"
	case keyDown:
		return "Down"
	case keyPageUp:
		return "PgUp"
	case keyPageDown:
		return "PgDn"
	case keyHome:
		return "Home"
	case keyEnd:
		return "End"
	case keyEnter:
		return "Enter"
	case keyEscape:
		return "Esc"
	case keyHelp:
		return "?"
	case keyBack, keyBackAlt:
		return "Backspace"
	default:
		return key
	}
}

func footerHints(model Model) []keyHint {
	if model.layout.Limited {
		if model.state == StateRunning {
			return []keyHint{{Key: keyLabel(keyCancel), Label: "cancel"}}
		}
		return []keyHint{{Key: keyLabel(keyCancel), Label: "quit"}}
	}
	if model.paletteOpen {
		return []keyHint{
			{Key: keyLabel(keyUp) + "/" + keyLabel(keyDown), Label: "navigate"},
			{Key: keyLabel(keyEnter), Label: "select"},
			{Key: keyLabel(keyEscape), Label: "close"},
		}
	}
	if model.state == StateRunning {
		return []keyHint{
			{Key: keyLabel(keyPageUp) + "/" + keyLabel(keyPageDown), Label: "scroll"},
			{Key: keyLabel(keyEnd), Label: "latest"},
			{Key: keyLabel(keyCancel), Label: "cancel"},
		}
	}
	if model.screen == ScreenWarnings && model.warningDetail {
		return []keyHint{
			{Key: keyLabel(keyPageUp) + "/" + keyLabel(keyPageDown), Label: "scroll"},
			{Key: keyLabel(keyEscape), Label: "back"},
			{Key: "/", Label: "commands"},
		}
	}
	if model.screen == ScreenWarnings && model.warningCount() > 0 {
		return []keyHint{
			{Key: keyLabel(keyUp) + "/" + keyLabel(keyDown), Label: "navigate"},
			{Key: keyLabel(keyEnter), Label: "details"},
			{Key: "/", Label: "commands"},
			{Key: keyLabel(keyEscape), Label: "back"},
		}
	}
	if model.viewport.Overflowing() {
		return []keyHint{
			{Key: keyLabel(keyPageUp) + "/" + keyLabel(keyPageDown), Label: "scroll"},
			{Key: keyLabel(keyEnd), Label: "latest"},
			{Key: "/", Label: "commands"},
			{Key: keyLabel(keyEscape), Label: "back"},
		}
	}
	if model.hasPrevious {
		return []keyHint{
			{Key: "/", Label: "commands"},
			{Key: keyLabel(keyEscape), Label: "back"},
			{Key: keyLabel(keyCancel), Label: "clear/quit"},
		}
	}
	return []keyHint{
		{Key: "/", Label: "commands"},
		{Key: keyLabel(keyHelp), Label: "help"},
		{Key: keyLabel(keyCancel), Label: "clear/quit"},
	}
}

func renderHints(hints []keyHint, compact bool) string {
	if compact && len(hints) > 4 {
		hints = hints[:4]
	}
	parts := make([]string, 0, len(hints))
	for _, hint := range hints {
		parts = append(parts, hint.Key+" "+hint.Label)
	}
	return strings.Join(parts, "   ")
}
