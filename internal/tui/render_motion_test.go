package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderReferenceFrames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		model    Model
		contains []string
	}{
		{
			name:  "ready",
			model: NewModel(Config{ProjectPath: `C:\DebugDoc`}),
			contains: []string{
				"DebugDoc",
				`C:\DebugDoc · Home`,
				"Ready to inspect bounded local project metadata.",
				"› Type / for commands",
			},
		},
		{
			name: "running animated",
			model: func() Model {
				model := NewModel(Config{ProjectPath: `C:\DebugDoc`})
				model.state = StateRunning
				model.showScreen(ScreenDiagnose)
				return model
			}(),
			contains: []string{
				"RUNNING",
				`C:\DebugDoc · Diagnose`,
				"◐  Inspecting project metadata",
				"Ctrl+C cancel",
			},
		},
		{
			name: "running reduced motion",
			model: func() Model {
				model := NewModel(Config{ProjectPath: `C:\DebugDoc`, ReduceMotion: true})
				model.state = StateRunning
				model.showScreen(ScreenDiagnose)
				return model
			}(),
			contains: []string{
				"RUNNING  Inspecting project metadata",
				"Ctrl+C cancel",
			},
		},
		{
			name: "warning result",
			model: func() Model {
				model := NewModel(Config{ProjectPath: `C:\DebugDoc`})
				model.state = StateWarning
				model.report = reportPointer(sampleReport())
				model.showScreen(ScreenDiagnose)
				return model
			}(),
			contains: []string{
				"WARNING",
				"Safe metadata inspection complete.",
				"/warnings   Review warnings and recommendations",
			},
		},
		{
			name: "transient footer",
			model: func() Model {
				model := NewModel(Config{ProjectPath: `C:\DebugDoc`})
				model.transient.set("Diagnosis complete", false)
				return model
			}(),
			contains: []string{"Diagnosis complete"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			view := test.model.View()
			for _, expected := range test.contains {
				if !strings.Contains(view, expected) {
					t.Errorf("render is missing %q:\n%s", expected, view)
				}
			}
			if strings.Contains(view, "Dev"+"Doctor") {
				t.Fatalf("legacy brand appears in render: %q", view)
			}
		})
	}
}

func TestMotionFramesPreserveSupportedLayoutBudgets(t *testing.T) {
	t.Parallel()
	for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 24}, {Width: 100, Height: 30}, {Width: 120, Height: 40}} {
		model := NewModel(Config{ProjectPath: `C:\DebugDoc`, ASCII: true})
		model.state = StateRunning
		model.showScreen(ScreenDiagnose)
		model = updateModel(t, model, size)
		for frame := 0; frame < len(asciiSpinnerFrames); frame++ {
			model.animation.frame = frame
			view := model.View()
			lines := strings.Split(view, "\n")
			if len(lines) != size.Height {
				t.Fatalf("%dx%d frame %d height = %d", size.Width, size.Height, frame, len(lines))
			}
			if model.layout.HeaderHeight != 2 || model.layout.FooterHeight != 1 || model.layout.ComposerHeight != 1 {
				t.Fatalf("%dx%d frame changed chrome budget: %#v", size.Width, size.Height, model.layout)
			}
			if !strings.Contains(lines[len(lines)-2], "> Type / for commands") {
				t.Fatalf("%dx%d frame %d moved composer: %q", size.Width, size.Height, frame, lines[len(lines)-2])
			}
		}
	}
}
