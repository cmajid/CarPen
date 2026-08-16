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

	// car is the game's own art, parked on the menu. It is the quickest way to
	// say what this game is, and it costs one sprite that is already embedded.
	car *ebiten.Image
}

func NewMenu(in Input) *Menu {
	return &Menu{
		in:   in,
		list: newMenuList(56, 286, 224, "Start Race", "Quit"),
		car:  carpen.CarImage("yellow"),
	}
}

func (m *Menu) Update() (Scene, error) {
	m.fade.update()

	switch m.list.update(m.in) {
	case menuStart:
		return newGameplay(m.in), nil
	case menuQuit:
		return nil, ebiten.Termination
	}

	// Esc is the way out of every other screen, so it is the way out of the game
	// here.
	if m.in.IsKeyJustPressed(ebiten.KeyEscape) {
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

	drawPrompts(screen,
		prompt{key: "Up / Down", does: "Move"},
		prompt{key: "Enter", does: "Select"},
		prompt{key: "Esc", does: "Quit"},
	)
	m.fade.draw(screen)
}

// drawParkingBay parks the car in a marked-out bay on the right of the screen,
// which is the whole game in one picture.
func (m *Menu) drawParkingBay(screen *ebiten.Image) {
	const (
		left  = 366
		right = 574
		top   = 96
		foot  = 384
	)

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
