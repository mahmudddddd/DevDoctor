package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mahmudddddd/DevDoctor/internal/model"
)

func TestBuildEnvironmentUsesMinimalBaselineAndExplicitValues(t *testing.T) {
	t.Setenv("DEBUGDOC_UNRELATED", "must-not-pass")
	t.Setenv("DEBUGDOC_SAFE_INPUT", "passed")
	environment, err := buildEnvironment(model.EnvironmentSpec{
		Pass: []string{"DEBUGDOC_SAFE_INPUT"},
		Set:  map[string]string{"DEBUGDOC_SAFE_OVERRIDE": "set"},
	})
	if err != nil {
		t.Fatalf("buildEnvironment() error = %v", err)
	}
	joined := strings.Join(environment, "\n")
	if strings.Contains(joined, "DEBUGDOC_UNRELATED=") {
		t.Fatal("unrelated host environment was inherited")
	}
	if !strings.Contains(joined, "DEBUGDOC_SAFE_INPUT=passed") || !strings.Contains(joined, "DEBUGDOC_SAFE_OVERRIDE=set") {
		t.Fatalf("explicit environment missing from %q", joined)
	}
}

func TestBuildEnvironmentDoesNotLoadDotEnv(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, ".env"), []byte("DEBUGDOC_DOTENV=loaded\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWorkingDirectory) })

	environment, err := buildEnvironment(model.EnvironmentSpec{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(environment, "\n"), "DEBUGDOC_DOTENV=") {
		t.Fatal(".env content was loaded")
	}
}

func TestEnvironmentRejectsSecretsDuplicatesAndMalformedNames(t *testing.T) {
	t.Parallel()
	tests := []model.EnvironmentSpec{
		{Set: map[string]string{"API_TOKEN": "do-not-leak"}},
		{Pass: []string{"SAFE_NAME", "SAFE_NAME"}},
		{Pass: []string{"BAD=NAME"}},
	}
	for _, spec := range tests {
		_, err := buildEnvironment(spec)
		if err == nil {
			t.Fatalf("buildEnvironment(%+v) succeeded", spec)
		}
		if strings.Contains(err.Error(), "do-not-leak") {
			t.Fatalf("error leaked environment value: %v", err)
		}
	}
}

func TestEnvironmentNamesNeverContainValues(t *testing.T) {
	t.Parallel()
	names, err := EnvironmentNames(model.EnvironmentSpec{Set: map[string]string{"SAFE_NAME": "private-value"}})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(names, "\n")
	if strings.Contains(joined, "private-value") || !strings.Contains(joined, "SAFE_NAME") {
		t.Fatalf("EnvironmentNames() = %q", joined)
	}
}
