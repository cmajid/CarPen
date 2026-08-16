package scene

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// menuList is a column of choices with one of them in focus. Every screen that
// asks the player to pick something is built from one, so picking works the same
// way throughout: Up and Down move, Enter or Space chooses, and the mouse can do
// both instead.
//
// The list wraps at both ends, which is what a linear menu is expected to do
// (Xbox Accessibility Guideline 112) and means the whole of it can be reached by
// holding one direction.
type menuList struct {
	items    []string
	selected int

	x, y, width float64

	// cursor is where the mouse was last seen. Hovering only takes the focus
	// when the mouse has actually moved, so a pointer left lying over the list
	// does not keep snatching the focus back from the arrow keys.
	cursor      image.Point
	cursorKnown bool
}

const (
	rowHeight = 38
	rowGap    = 4  // left between one row's highlight and the next
	rowPadX   = 18 // from the edge of the highlight to the start of the label
)

// nothingChosen is what update reports on the ticks where the player has not
// settled on anything yet, which is nearly all of them.
const nothingChosen = -1

func newMenuList(x, y, width float64, items ...string) *menuList {
	return &menuList{items: items, x: x, y: y, width: width}
}

// update moves the focus around and reports the item the player chose, or
// nothingChosen.
func (l *menuList) update(in Input) int {
	if in.IsKeyJustPressed(ebiten.KeyDown) || in.IsKeyJustPressed(ebiten.KeyTab) {
		l.move(1)
	}
	if in.IsKeyJustPressed(ebiten.KeyUp) {
		l.move(-1)
	}

	cursor := image.Pt(in.CursorPosition())
	if l.cursorKnown && cursor != l.cursor {
		if row := l.rowAt(cursor); row != nothingChosen {
			l.selected = row
		}
	}
	l.cursor, l.cursorKnown = cursor, true

	if in.IsKeyJustPressed(ebiten.KeyEnter) || in.IsKeyJustPressed(ebiten.KeySpace) {
		return l.selected
	}
	if in.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if row := l.rowAt(cursor); row != nothingChosen {
			l.selected = row
			return row
		}
	}

	return nothingChosen
}

// move walks the focus by delta, coming back round at either end.
func (l *menuList) move(delta int) {
	count := len(l.items)
	l.selected = ((l.selected+delta)%count + count) % count
}

// rowAt is the item under p, or nothingChosen if p is not on the list at all.
func (l *menuList) rowAt(p image.Point) int {
	for row := range l.items {
		if p.In(l.bounds(row)) {
			return row
		}
	}
	return nothingChosen
}

func (l *menuList) bounds(row int) image.Rectangle {
	top := int(l.y) + row*(rowHeight+rowGap)
	return image.Rect(int(l.x), top, int(l.x+l.width), top+rowHeight)
}

// draw paints the list, marking the focused item with a filled bar, a heavier
// weight and a darker label. The focus is deliberately not carried by colour
// alone: a player who cannot pick the amber out can still see which row has
// turned solid.
func (l *menuList) draw(dst *ebiten.Image) {
	for row, item := range l.items {
		box := l.bounds(row)
		face, colour := text.Face(fontItem), colourText

		if row == l.selected {
			fillRect(dst, float64(box.Min.X), float64(box.Min.Y), l.width, rowHeight, colourAccent)
			face, colour = fontItemHot, colourAccentInk
		} else {
			fillRect(dst, float64(box.Min.X), float64(box.Min.Y), 2, rowHeight, colourLine)
		}

		drawText(dst, item, face, l.x+rowPadX, float64(box.Min.Y)+rowHeight/2, colour, text.AlignStart, text.AlignCenter)
	}
}

// height is how much room the whole list takes, for placing whatever sits under
// it.
func (l *menuList) height() float64 {
	return float64(len(l.items)*(rowHeight+rowGap) - rowGap)
}
