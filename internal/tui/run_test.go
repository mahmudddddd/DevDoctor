package tui

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestConfigForEnvironmentHonorsNoColorAndDumbTerminal(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERM", "dumb")

	config := configForEnvironment(Config{Color: true})
	if config.Color || !config.ASCII || !config.Flat || !config.ReduceMotion {
		t.Fatalf("configForEnvironment() = %#v, want no-color ASCII flat reduced-motion mode", config)
	}
	view := NewModel(config).View()
	if strings.Contains(view, "\x1b[") || !strings.Contains(view, "[READY]") || !strings.Contains(view, "> Type / for commands") {
		t.Fatalf("dumb-terminal view is not readable flat text: %q", view)
	}
}

func TestConfigForEnvironmentUsesCanonicalReducedMotionVariable(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("DEBUGDOC_REDUCE_MOTION", "1")
	legacyName := "DEV" + "DOCTOR_REDUCE_MOTION"
	t.Setenv(legacyName, "1")

	config := configForEnvironment(Config{})
	if !config.ReduceMotion {
		t.Fatal("DEBUGDOC_REDUCE_MOTION=1 did not enable reduced motion")
	}

	t.Setenv("DEBUGDOC_REDUCE_MOTION", "")
	config = configForEnvironment(Config{})
	if config.ReduceMotion {
		t.Fatal("legacy reduced-motion variable was treated as canonical")
	}
}

func TestRunEntersAndRestoresAlternateScreen(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var output bytes.Buffer
	err := Run(RunOptions{
		Config: Config{Context: ctx, ProjectPath: "/project", ASCII: true, Flat: true},
		Input:  bytes.NewReader([]byte{3, 3}),
		Output: &output,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	rendered := output.String()
	if !strings.Contains(rendered, "\x1b[?1049h") || !strings.Contains(rendered, "\x1b[?1049l") {
		t.Fatalf("alternate-screen entry/restoration sequences missing: %q", rendered)
	}
}
