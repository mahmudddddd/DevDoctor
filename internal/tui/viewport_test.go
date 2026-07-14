package tui

import "testing"

func TestViewportScrollingAndBottomFollow(t *testing.T) {
	t.Parallel()
	viewport := newViewport()
	viewport.SetSize(20, 3)
	viewport.SetContent([]string{"one", "two", "three", "four", "five"})
	if !viewport.Following() || viewport.offset != 2 {
		t.Fatalf("initial follow/offset = %v/%d, want true/2", viewport.Following(), viewport.offset)
	}

	viewport.ScrollUp(1)
	if viewport.Following() || viewport.offset != 1 {
		t.Fatalf("after scroll follow/offset = %v/%d, want false/1", viewport.Following(), viewport.offset)
	}
	viewport.SetContent([]string{"one", "two", "three", "four", "five", "six"})
	if viewport.offset != 1 {
		t.Fatalf("manual offset after update = %d, want 1", viewport.offset)
	}
	viewport.JumpToLatest()
	if !viewport.Following() || viewport.offset != 3 {
		t.Fatalf("latest follow/offset = %v/%d, want true/3", viewport.Following(), viewport.offset)
	}
}

func TestViewportResizePreservesLogicalAnchor(t *testing.T) {
	t.Parallel()
	viewport := newViewport()
	viewport.SetSize(12, 2)
	viewport.SetContent([]string{"first row wraps here", "second", "third"})
	viewport.JumpToStart()
	viewport.ScrollDown(2)
	anchor := viewport.logicalAnchor()
	viewport.SetSize(8, 2)
	if got := viewport.logicalAnchor(); got != anchor {
		t.Fatalf("logical anchor after resize = %d, want %d", got, anchor)
	}
	if viewport.offset < 0 {
		t.Fatalf("offset = %d, want nonnegative", viewport.offset)
	}
}

func TestViewportResizePreservesPositionWithinWrappedLogicalRow(t *testing.T) {
	t.Parallel()
	viewport := newViewport()
	viewport.SetSize(12, 2)
	viewport.SetContent([]string{"aaaaa bbbbb ccccc ddddd", "second", "third", "fourth"})
	viewport.JumpToStart()
	viewport.ScrollDown(1)
	if got := viewport.Lines()[0]; got != "ccccc ddddd" {
		t.Fatalf("test setup visible row = %q, want ccccc ddddd", got)
	}

	viewport.SetSize(8, 2)
	if got := viewport.Lines()[0]; got != "ccccc" {
		t.Fatalf("visible row after resize = %q, want preserved ccccc anchor", got)
	}
}

func TestViewportHandlesEmptyAndOneLineContent(t *testing.T) {
	t.Parallel()
	viewport := newViewport()
	viewport.SetSize(10, 4)
	viewport.SetContent(nil)
	viewport.ScrollUp(1)
	if viewport.offset != 0 {
		t.Fatalf("empty offset = %d, want 0", viewport.offset)
	}
	viewport.SetContent([]string{"one"})
	viewport.PageDown()
	if viewport.offset != 0 || !viewport.Following() {
		t.Fatalf("one-line offset/follow = %d/%v, want 0/true", viewport.offset, viewport.Following())
	}
}
