package tui

// Layout contains terminal-cell budgets for the persistent shell regions.
type Layout struct {
	Width          int
	Height         int
	ContentWidth   int
	HeaderHeight   int
	ViewportHeight int
	ComposerHeight int
	FooterHeight   int
	DividerHeight  int
	Limited        bool
	Compact        bool
}

// PaletteLayout describes the transient palette area above the composer.
type PaletteLayout struct {
	Width  int
	Height int
	Top    int
}

// ComputeLayout derives a stable responsive layout for a terminal size.
func ComputeLayout(width, height int) Layout {
	layout := Layout{
		Width:   max(0, width),
		Height:  max(0, height),
		Limited: width < minimumWidth || height < minimumHeight,
		Compact: width < 100 || height < 30,
	}
	if layout.Limited {
		layout.ContentWidth = max(1, width)
		layout.ViewportHeight = max(1, height)
		return layout
	}

	layout.ContentWidth = min(88, max(1, width))
	layout.HeaderHeight = 2
	layout.ComposerHeight = 1
	layout.FooterHeight = 1
	layout.DividerHeight = 1
	layout.ViewportHeight = max(1, height-layout.HeaderHeight-layout.ComposerHeight-layout.FooterHeight-layout.DividerHeight)
	return layout
}

// PaletteGeometry bounds and anchors the command palette inside the viewport.
func (l Layout) PaletteGeometry(rows int) PaletteLayout {
	if l.Limited || rows <= 0 {
		return PaletteLayout{}
	}
	height := min(rows, min(12, max(1, l.ViewportHeight-1)))
	width := min(64, max(1, l.ContentWidth-4))
	return PaletteLayout{
		Width:  width,
		Height: height,
		Top:    max(0, l.ViewportHeight-height),
	}
}
