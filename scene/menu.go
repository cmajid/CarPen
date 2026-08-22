package scene

import (
	"math"

	"github.com/cmajid/carpen/carpen"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// The menu's choices, in the order they are offered. Starting a race is first,
// so the game is one key away from the screen it opens on.
const (
	menuStart = iota
	menuQuit
)

// Menu is the screen the game opens on.
type Menu struct {
	in   Input
	list *menuList
	fade fade

	// level is the puzzle Start Race deals out. The menu carries it rather than
	// loading it, so a level that will not load is reported when the game
	// starts up instead of when the player asks to play.
	level carpen.Level

	// car is the game's own art, parked on the menu. It is the quickest way to
	// say what this game is, and it costs one sprite that is already embedded.
	car *ebiten.Image

	// width is the screen the manager last reported. The title and the list are
	// set against the left edge and stay where they are put; the parked car is
	// hung off the right edge, so it needs the number.
	width int
}

func NewMenu(in Input, level carpen.Level) *Menu {
	return &Menu{
		in:    in,
		list:  newMenuList(56, 286, 224, "Start Race", "Quit"),
		level: level,
		car:   carpen.CarImage("yellow"),
		width: DesignWidth,
	}
}

func (m *Menu) resize(v viewport) { m.width = v.width }

func (m *Menu) Update() (Scene, error) {
	m.fade.update()

	switch m.list.update(m.in) {
	case menuStart:
		return newGameplay(m.in, m.level), nil
	case menuQuit:
		return nil, ebiten.Termination
	}

	// Backing out is the way out of every other screen, so it is the way out of
	// the game here.
	if justPressed(m.in, actionCancel) {
		return nil, ebiten.Termination
	}

	return nil, nil
}

func (m *Menu) Draw(screen *ebiten.Image) {
	screen.Fill(colourInk)
	m.drawParkingBay(screen)

	drawText(screen, "CarPen", fontTitle, 56, 150, colourText, text.AlignStart, text.AlignEnd)
	fillRect(screen, 56, 168, 58, 4, colourAccent)
	drawText(screen, "Park your car carefully", fontBody, 56, 200, colourTextMuted, text.AlignStart, text.AlignStart)

	m.list.draw(screen)

	drawMenuPrompts(screen, m.in, "Quit")
	m.fade.draw(screen)
}

// drawParkingBay parks the car in a marked-out bay on the right of the screen,
// which is the whole game in one picture. It is measured in from the right edge
// rather than written down, so that a wider screen puts the bay against the
// same margin instead of stranding it in the middle with the title.
func (m *Menu) drawParkingBay(screen *ebiten.Image) {
	const (
		fromEdge = 66  // from the right of the screen to the right of the bay
		bayWidth = 208 // between the two painted lines
		top      = 96
		foot     = 384
	)

	right := float64(m.width) - fromEdge
	left := right - bayWidth

	fillRect(screen, left, top, 4, foot-top, colourLine)
	fillRect(screen, right, top, 4, foot-top, colourLine)
	fillRect(screen, left, foot, right-left+4, 4, colourLine)

	const scale = 0.62
	width := float64(m.car.Bounds().Dx()) * scale
	height := float64(m.car.Bounds().Dy()) * scale

	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(-width/2, -height/2)
	op.GeoM.Rotate(4 * math.Pi / 180) // parked a touch crooked, as one does
	op.GeoM.Translate((left+right)/2+2, (top+foot)/2+8)
	screen.DrawImage(m.car, op)
}
