package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mahmudddddd/DevDoctor/internal/report"
)

const (
	animationInterval = 160 * time.Millisecond
	transientDuration = 1800 * time.Millisecond
)

var (
	unicodeSpinnerFrames = [...]string{"◐", "◓", "◑", "◒"}
	asciiSpinnerFrames   = [...]string{"|", "/", "-", "\\"}
)

type animationTickMsg struct {
	generation uint64
}

type transientExpiredMsg struct {
	revision uint64
}

type animationState struct {
	enabled bool
	ascii   bool
	frame   int
}

func newAnimationState(enabled, ascii bool) animationState {
	return animationState{enabled: enabled, ascii: ascii}
}

func (a animationState) indicator() string {
	if !a.enabled {
		return "RUNNING"
	}
	if a.ascii {
		return asciiSpinnerFrames[a.frame%len(asciiSpinnerFrames)]
	}
	return unicodeSpinnerFrames[a.frame%len(unicodeSpinnerFrames)]
}

func (a *animationState) advance() {
	if !a.enabled {
		return
	}
	frameCount := len(unicodeSpinnerFrames)
	if a.ascii {
		frameCount = len(asciiSpinnerFrames)
	}
	a.frame = (a.frame + 1) % frameCount
}

func (a animationState) tick(generation uint64) tea.Cmd {
	if !a.enabled {
		return nil
	}
	return tea.Tick(animationInterval, func(time.Time) tea.Msg {
		return animationTickMsg{generation: generation}
	})
}

type transientFooter struct {
	text       string
	revision   uint64
	persistent bool
}

func (f *transientFooter) set(text string, persistent bool) tea.Cmd {
	f.revision++
	f.text = report.SafeText(text)
	f.persistent = persistent
	if persistent || f.text == "" {
		return nil
	}
	revision := f.revision
	return tea.Tick(transientDuration, func(time.Time) tea.Msg {
		return transientExpiredMsg{revision: revision}
	})
}

func (f *transientFooter) expire(message transientExpiredMsg) bool {
	if f.persistent || message.revision != f.revision || f.text == "" {
		return false
	}
	f.text = ""
	return true
}

func (f *transientFooter) clear() {
	if f.text == "" && !f.persistent {
		return
	}
	f.revision++
	f.text = ""
	f.persistent = false
}
