package tui

import (
	"strings"
	"testing"
)

func TestCommandsContainExactAllowlist(t *testing.T) {
	t.Parallel()
	commands := Commands()
	got := make([]string, len(commands))
	for index, command := range commands {
		got[index] = command.Name
	}
	want := []string{"/diagnose", "/project", "/warnings", "/export", "/help", "/clear", "/quit"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("Commands() = %v, want %v", got, want)
	}
}

func TestFilterCommandsMatchesNameAndDescriptionCaseInsensitively(t *testing.T) {
	t.Parallel()
	matches := FilterCommands("/DIA")
	if len(matches) != 1 || matches[0].ID != CommandDiagnose {
		t.Fatalf("FilterCommands(/DIA) = %#v, want diagnose", matches)
	}
	matches = FilterCommands("preview")
	if len(matches) != 1 || matches[0].ID != CommandExport {
		t.Fatalf("FilterCommands(deterministic) = %#v, want export", matches)
	}
}

func TestParseActionRejectsArbitraryInputAndArguments(t *testing.T) {
	t.Parallel()
	for _, input := range []string{"go test ./...", "! dir", "/diagnose .", "/unknown"} {
		if _, err := ParseAction(input); err == nil {
			t.Fatalf("ParseAction(%q) error = nil, want rejection", input)
		}
	}
	action, err := ParseAction("/PrOjEcT")
	if err != nil || action.ID != CommandProject {
		t.Fatalf("ParseAction(/PrOjEcT) = %#v, %v", action, err)
	}
}
