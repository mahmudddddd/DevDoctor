package tui

import "testing"

func TestKeymapUsesDistinctRoutedBindings(t *testing.T) {
	t.Parallel()
	keys := []string{keyCancel, keyRedraw, keyHelp, keyUp, keyDown, keyPageUp, keyPageDown, keyHome, keyEnd, keyEnter, keyBack, keyBackAlt, keyEscape}
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate routed key %q", key)
		}
		seen[key] = struct{}{}
		if keyLabel(key) == "" {
			t.Fatalf("key %q has no footer label", key)
		}
	}
}
