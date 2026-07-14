package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	projectmodel "github.com/mahmudddddd/DevDoctor/internal/model"
)

func TestAnimationFramesAdvanceDeterministically(t *testing.T) {
	t.Parallel()
	unicodeAnimation := newAnimationState(true, false)
	if got := unicodeAnimation.indicator(); got != "◐" {
		t.Fatalf("initial Unicode frame = %q", got)
	}
	unicodeAnimation.advance()
	if got := unicodeAnimation.indicator(); got != "◓" {
		t.Fatalf("advanced Unicode frame = %q", got)
	}

	asciiAnimation := newAnimationState(true, true)
	if got := asciiAnimation.indicator(); got != "|" {
		t.Fatalf("initial ASCII frame = %q", got)
	}
	asciiAnimation.advance()
	if got := asciiAnimation.indicator(); got != "/" {
		t.Fatalf("advanced ASCII frame = %q", got)
	}
}

func TestReducedMotionUsesStaticRunningLabelAndNoTick(t *testing.T) {
	t.Parallel()
	animation := newAnimationState(false, false)
	if got := animation.indicator(); got != "RUNNING" {
		t.Fatalf("reduced-motion indicator = %q", got)
	}
	if command := animation.tick(1); command != nil {
		t.Fatal("reduced motion scheduled a cosmetic tick")
	}
	animation.advance()
	if animation.frame != 0 {
		t.Fatalf("reduced-motion frame advanced to %d", animation.frame)
	}
}

func TestIdleModelSchedulesNoAnimationTick(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project"})
	if command := model.Init(); command != nil {
		t.Fatal("idle model scheduled a command")
	}
}

func TestActiveDiagnosisSchedulesMotionOnlyWhenEnabled(t *testing.T) {
	t.Parallel()
	discover := func(context.Context, string) (projectmodel.ProjectReport, error) {
		return sampleReport(), nil
	}

	animated := NewModel(Config{ProjectPath: "/project", Discover: discover})
	started, command := animated.startDiscovery()
	animated = started.(Model)
	if animated.state != StateRunning || command == nil {
		t.Fatalf("animated start state/command = %s/%v", animated.state, command != nil)
	}
	batch, ok := command().(tea.BatchMsg)
	if !ok || len(batch) != 3 {
		t.Fatalf("animated start batch = %T/%d, want discovery, animation, transient", batch, len(batch))
	}

	reduced := NewModel(Config{ProjectPath: "/project", Discover: discover, ReduceMotion: true})
	started, command = reduced.startDiscovery()
	reduced = started.(Model)
	if reduced.animation.enabled {
		t.Fatal("reduced-motion model enabled animation")
	}
	batch, ok = command().(tea.BatchMsg)
	if !ok || len(batch) != 2 {
		t.Fatalf("reduced start batch = %T/%d, want discovery and transient only", batch, len(batch))
	}
}

func TestAnimationTickMutatesOnlyAnimationState(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project"})
	model.state = StateRunning
	model.generation = 7
	model.draft = "/dia"
	model.screen = ScreenDiagnose
	model.warningSelected = 2
	beforeContent := append([]string(nil), model.viewport.logical...)

	updated, command := model.Update(animationTickMsg{generation: 7})
	result := updated.(Model)
	if result.animation.frame != 1 || command == nil {
		t.Fatalf("frame/next command = %d/%v", result.animation.frame, command != nil)
	}
	if result.draft != model.draft || result.screen != model.screen || result.warningSelected != model.warningSelected {
		t.Fatalf("animation tick mutated unrelated model state: %#v", result)
	}
	if strings.Join(result.viewport.logical, "\n") != strings.Join(beforeContent, "\n") {
		t.Fatal("animation tick mutated viewport logical content")
	}
}

func TestTerminalStatesIgnoreStaleAnimationTicks(t *testing.T) {
	t.Parallel()
	for _, state := range []RunState{StateOK, StateWarning, StateFailed, StateCancelled, StateTimedOut} {
		model := NewModel(Config{ProjectPath: "/project"})
		model.state = state
		model.generation = 2
		updated, command := model.Update(animationTickMsg{generation: 2})
		result := updated.(Model)
		if result.animation.frame != 0 || command != nil {
			t.Fatalf("%s accepted stale animation tick: frame=%d command=%v", state, result.animation.frame, command != nil)
		}
	}
}

func TestRunningRenderUpdatesInPlaceWithoutDuplicateActivity(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project", ASCII: true})
	model.state = StateRunning
	model.generation = 3
	model.showScreen(ScreenDiagnose)
	before := model.View()
	updated, _ := model.Update(animationTickMsg{generation: 3})
	after := updated.(Model).View()
	if strings.Count(before, "Inspecting project metadata") != 1 || strings.Count(after, "Inspecting project metadata") != 1 {
		t.Fatalf("activity line was duplicated:\n%s\n---\n%s", before, after)
	}
	if before == after {
		t.Fatal("animated running frame did not change")
	}
}

func TestReducedMotionRunningRenderIsSemanticallyComplete(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project", ReduceMotion: true})
	model.state = StateRunning
	model.showScreen(ScreenDiagnose)
	view := model.View()
	if !strings.Contains(view, "RUNNING  Inspecting project metadata") {
		t.Fatalf("reduced-motion running state lacks static semantics: %q", view)
	}
}

func TestTransientFooterExpiresAndReplacementWins(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project", ASCII: true, Flat: true})
	firstCommand := model.transient.set("first", false)
	firstRevision := model.transient.revision
	secondCommand := model.transient.set("second", false)
	secondRevision := model.transient.revision
	if firstCommand == nil || secondCommand == nil || model.transient.text != "second" {
		t.Fatalf("transient replacement failed: %#v", model.transient)
	}

	updated, _ := model.Update(transientExpiredMsg{revision: firstRevision})
	model = updated.(Model)
	if model.transient.text != "second" {
		t.Fatal("stale expiry cleared replacement message")
	}
	updated, _ = model.Update(transientExpiredMsg{revision: secondRevision})
	model = updated.(Model)
	if model.transient.text != "" {
		t.Fatalf("current transient did not expire: %q", model.transient.text)
	}
}

func TestTransientFooterIsSanitizedAndDoesNotChangeHeight(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project", ASCII: true, Flat: true})
	height := model.layout.FooterHeight
	model.transient.set("done\x1b[31m", false)
	view := model.View()
	if strings.Contains(view, "\x1b[31m") || !strings.Contains(view, "done\\u001B[31m") {
		t.Fatalf("transient message was not sanitized: %q", view)
	}
	if model.layout.FooterHeight != height || height != 1 {
		t.Fatalf("footer height changed from %d to %d", height, model.layout.FooterHeight)
	}
}

func TestPersistentValidationOutranksTransientFooter(t *testing.T) {
	t.Parallel()
	model := NewModel(Config{ProjectPath: "/project", ASCII: true, Flat: true})
	model.transient.set("opened Help", false)
	model.validation = "Input rejected."
	footer := model.renderFooter()
	if len(footer) != 1 || !strings.Contains(footer[0], "Input rejected") || strings.Contains(footer[0], "opened Help") {
		t.Fatalf("footer priority = %#v", footer)
	}
}
