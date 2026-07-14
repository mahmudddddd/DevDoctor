package tui

import (
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

func appendPrintableInput(draft string, key tea.KeyMsg) (string, bool) {
	if key.Alt || (key.Type != tea.KeyRunes && key.Type != tea.KeySpace) {
		return draft, false
	}

	var input strings.Builder
	for _, value := range key.Runes {
		if unicode.IsPrint(value) {
			input.WriteRune(value)
		}
	}
	if key.Type == tea.KeySpace && input.Len() == 0 {
		input.WriteByte(' ')
	}
	if input.Len() == 0 {
		return draft, false
	}
	return draft + input.String(), true
}
