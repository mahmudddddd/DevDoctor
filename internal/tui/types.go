package tui

import (
	"context"

	"github.com/mahmudddddd/DevDoctor/internal/model"
)

const (
	minimumWidth  = 80
	minimumHeight = 24
)

// DiscoverFunc performs the existing read-only Phase 1 project discovery.
type DiscoverFunc func(context.Context, string) (model.ProjectReport, error)

// Screen identifies the active primary view.
type Screen int

// Supported primary screens.
const (
	ScreenHome Screen = iota
	ScreenDiagnose
	ScreenProject
	ScreenWarnings
	ScreenExport
	ScreenHelp
)

// RunState is the explicit lifecycle state shown by the shell.
type RunState string

// Supported interactive lifecycle states.
const (
	StateReady     RunState = "READY"
	StateRunning   RunState = "RUNNING"
	StateWaiting   RunState = "WAITING"
	StateOK        RunState = "OK"
	StateWarning   RunState = "WARNING"
	StateFailed    RunState = "FAILED"
	StateCancelled RunState = "CANCELLED"
	StateTimedOut  RunState = "TIMED OUT"
	StateSkipped   RunState = "SKIPPED"
)

// Focus identifies the single active keyboard surface.
type Focus int

// Supported keyboard focus owners.
const (
	FocusComposer Focus = iota
	FocusPalette
	FocusWarnings
	FocusWarningDetail
)

// Capabilities controls terminal presentation without changing information.
type Capabilities struct {
	Color bool
	ASCII bool
	Flat  bool
}

// Config supplies the TUI with presentation and discovery dependencies.
type Config struct {
	Context     context.Context
	ProjectPath string
	Discover    DiscoverFunc
	Color       bool
	ASCII       bool
	Flat        bool
}

func (screen Screen) String() string {
	switch screen {
	case ScreenHome:
		return "Home"
	case ScreenDiagnose:
		return "Diagnose"
	case ScreenProject:
		return "Project"
	case ScreenWarnings:
		return "Warnings"
	case ScreenExport:
		return "Export"
	case ScreenHelp:
		return "Help"
	default:
		return "Unknown"
	}
}
