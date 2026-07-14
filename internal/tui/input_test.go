package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModifierAndControlInputDoesNotMutateDraft(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{name: "bare ctrl nul rune", key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{0}}},
		{name: "bare alt", key: tea.KeyMsg{Type: tea.KeyRunes, Alt: true}},
		{name: "alt printable", key: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}, Alt: true}},
		{name: "unknown ctrl", key: tea.KeyMsg{Type: tea.KeyCtrlA}},
		{name: "shift only", key: tea.KeyMsg{Type: tea.KeyShiftLeft}},
		{name: "unknown special", key: tea.KeyMsg{Type: tea.KeyType(999)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := NewModel(Config{ProjectPath: "/project"})
			model.draft = "safe"
			model = updateModel(t, model, test.key)
			if model.draft != "safe" {
				t.Fatalf("draft = %q, want unchanged", model.draft)
			}
		})
	}
}

func TestPrintableTextAndPasteAreAccepted(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project"})
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("hello")},
		{Type: tea.KeySpace, Runes: []rune{' '}},
		{Type: tea.KeyRunes, Runes: []rune("世界🙂")},
		{Type: tea.KeyRunes, Runes: []rune(" café界"), Paste: true},
	} {
		model = updateModel(t, model, key)
	}
	if model.draft != "hello 世界🙂 café界" {
		t.Fatalf("draft = %q", model.draft)
	}
}

func TestControlCharactersAreFilteredFromPaste(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project", ASCII: true, Flat: true})
	model = updateModel(t, model, tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'a', 0, '\n', '\t', '界', '\x1b', 'b'},
		Paste: true,
	})
	if model.draft != "a界b" {
		t.Fatalf("draft = %q, want printable paste only", model.draft)
	}
	view := model.View()
	if strings.Contains(view, "\\u0000") || strings.ContainsRune(view, 0) {
		t.Fatalf("composer surfaced NUL text: %q", view)
	}
}

func TestPaletteInputUsesSamePrintableFilter(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project"})
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{0}})
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}, Alt: true})
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("pro")})
	if model.draft != "/pro" {
		t.Fatalf("palette draft = %q, want /pro", model.draft)
	}
}
