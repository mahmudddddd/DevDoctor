package tui

import (
	"strings"
	"unicode"

	"github.com/mattn/go-runewidth"
)

// Viewport stores logical content and the user's scroll intent.
type Viewport struct {
	logical   []string
	rendered  []string
	width     int
	height    int
	offset    int
	following bool
}

type viewportAnchor struct {
	logical    int
	cellOffset int
}

func newViewport() Viewport {
	return Viewport{following: true}
}

// SetSize reflows content and preserves the nearest logical anchor.
func (v *Viewport) SetSize(width, height int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	oldAnchor := v.anchor()
	v.width = width
	v.height = height
	v.reflow()
	if v.following {
		v.JumpToLatest()
		return
	}
	v.offset = v.offsetForAnchor(oldAnchor)
	v.clamp()
}

// SetContent replaces logical rows while preserving scroll intent.
func (v *Viewport) SetContent(lines []string) {
	wasFollowing := v.following
	oldAnchor := v.anchor()
	v.logical = append(v.logical[:0], lines...)
	v.reflow()
	if wasFollowing {
		v.JumpToLatest()
		return
	}
	v.offset = v.offsetForAnchor(oldAnchor)
	v.clamp()
}

// ScrollUp moves away from live bottom-follow by the requested rows.
func (v *Viewport) ScrollUp(rows int) {
	if rows < 1 {
		rows = 1
	}
	v.offset -= rows
	v.clamp()
	v.following = v.offset >= v.maxOffset()
}

// ScrollDown moves toward the latest rendered rows.
func (v *Viewport) ScrollDown(rows int) {
	if rows < 1 {
		rows = 1
	}
	v.offset += rows
	v.clamp()
	v.following = v.offset >= v.maxOffset()
}

// PageUp moves one viewport page with one row of context.
func (v *Viewport) PageUp() {
	v.ScrollUp(max(1, v.height-1))
}

// PageDown moves one viewport page with one row of context.
func (v *Viewport) PageDown() {
	v.ScrollDown(max(1, v.height-1))
}

// JumpToStart moves to the first rendered row.
func (v *Viewport) JumpToStart() {
	v.offset = 0
	v.following = len(v.rendered) <= v.height
}

// JumpToLatest restores live bottom-follow.
func (v *Viewport) JumpToLatest() {
	v.offset = v.maxOffset()
	v.following = true
}

// Lines returns the currently visible rendered rows.
func (v Viewport) Lines() []string {
	end := min(len(v.rendered), v.offset+v.height)
	if v.offset >= end {
		return nil
	}
	lines := make([]string, end-v.offset)
	copy(lines, v.rendered[v.offset:end])
	return lines
}

// Position reports the visible one-based row range and total rows.
func (v Viewport) Position() (start, end, total int) {
	if len(v.rendered) == 0 {
		return 0, 0, 0
	}
	return v.offset + 1, min(len(v.rendered), v.offset+v.height), len(v.rendered)
}

// Following reports whether new content should stay pinned to the bottom.
func (v Viewport) Following() bool {
	return v.following
}

// Overflowing reports whether scrolling is possible.
func (v Viewport) Overflowing() bool {
	return len(v.rendered) > v.height
}

// RevealLogical scrolls only as much as needed to reveal a logical row.
func (v *Viewport) RevealLogical(index int) {
	if index < 0 || index >= len(v.logical) {
		return
	}
	start := v.offsetForLogical(index)
	height := len(wrapLine(v.logical[index], max(1, v.width)))
	if height < 1 {
		height = 1
	}
	end := start + height
	if start < v.offset {
		v.offset = start
	} else if end > v.offset+v.height {
		v.offset = end - v.height
	}
	v.clamp()
	v.following = v.offset >= v.maxOffset()
}

func (v *Viewport) reflow() {
	v.rendered = v.rendered[:0]
	for _, line := range v.logical {
		wrapped := wrapLine(line, v.width)
		if len(wrapped) == 0 {
			wrapped = []string{""}
		}
		v.rendered = append(v.rendered, wrapped...)
	}
	v.clamp()
}

func (v Viewport) logicalAnchor() int {
	return v.anchor().logical
}

func (v Viewport) anchor() viewportAnchor {
	if len(v.logical) == 0 {
		return viewportAnchor{}
	}
	remaining := v.offset
	for index, line := range v.logical {
		height := len(wrapLine(line, max(1, v.width)))
		if height == 0 {
			height = 1
		}
		if remaining < height {
			starts := wrappedStartCells(line, max(1, v.width))
			cellOffset := 0
			if remaining < len(starts) {
				cellOffset = starts[remaining]
			}
			return viewportAnchor{logical: index, cellOffset: cellOffset}
		}
		remaining -= height
	}
	return viewportAnchor{logical: len(v.logical) - 1}
}

func (v Viewport) offsetForAnchor(anchor viewportAnchor) int {
	offset := v.offsetForLogical(anchor.logical)
	if anchor.logical < 0 || anchor.logical >= len(v.logical) {
		return offset
	}
	starts := wrappedStartCells(v.logical[anchor.logical], max(1, v.width))
	wrappedOffset := 0
	for index, start := range starts {
		if start > anchor.cellOffset {
			break
		}
		wrappedOffset = index
	}
	return offset + wrappedOffset
}

func (v Viewport) offsetForLogical(anchor int) int {
	if anchor <= 0 {
		return 0
	}
	if anchor > len(v.logical) {
		anchor = len(v.logical)
	}
	offset := 0
	for _, line := range v.logical[:anchor] {
		height := len(wrapLine(line, max(1, v.width)))
		if height == 0 {
			height = 1
		}
		offset += height
	}
	return offset
}

func (v Viewport) maxOffset() int {
	return max(0, len(v.rendered)-v.height)
}

func (v *Viewport) clamp() {
	v.offset = min(max(0, v.offset), v.maxOffset())
}

func wrappedStartCells(value string, width int) []int {
	wrapped := wrapLine(value, width)
	starts := make([]int, len(wrapped))
	cursor := 0
	for index, segment := range wrapped {
		if segment == "" {
			starts[index] = runewidth.StringWidth(value[:cursor])
			continue
		}
		relative := strings.Index(value[cursor:], segment)
		if relative < 0 {
			starts[index] = min(runewidth.StringWidth(value), index*max(1, width))
			continue
		}
		start := cursor + relative
		starts[index] = runewidth.StringWidth(value[:start])
		cursor = start + len(segment)
	}
	return starts
}

func wrapLine(value string, width int) []string {
	if width < 1 {
		return []string{""}
	}
	if value == "" {
		return []string{""}
	}

	var lines []string
	var current strings.Builder
	currentWidth := 0
	lastSpaceByte := -1

	flush := func() {
		lines = append(lines, strings.TrimRightFunc(current.String(), unicode.IsSpace))
		current.Reset()
		currentWidth = 0
		lastSpaceByte = -1
	}

	for _, r := range value {
		runeWidth := runewidth.RuneWidth(r)
		if runeWidth < 1 {
			runeWidth = 1
		}
		if currentWidth+runeWidth > width && current.Len() > 0 {
			if lastSpaceByte >= 0 {
				whole := current.String()
				lines = append(lines, strings.TrimRightFunc(whole[:lastSpaceByte], unicode.IsSpace))
				remainder := strings.TrimLeftFunc(whole[lastSpaceByte:], unicode.IsSpace)
				current.Reset()
				current.WriteString(remainder)
				currentWidth = runewidth.StringWidth(remainder)
				lastSpaceByte = -1
			} else {
				flush()
			}
		}
		if unicode.IsSpace(r) {
			lastSpaceByte = current.Len()
		}
		current.WriteRune(r)
		currentWidth += runeWidth
	}
	if current.Len() > 0 || len(lines) == 0 {
		flush()
	}
	return lines
}
