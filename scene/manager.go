package scene

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Manager runs one scene at a time. It is the ebiten.Game the window drives, and
// the only part of the game Ebiten knows about.
type Manager struct {
	current Scene
	width,
	height int

	// screenWidth and screenHeight are the size Layout last settled on, which
	// is what the scenes lay themselves out against. They are not the same as
	// width and height: those are the size the game was written at, and this is
	// the shape of the device it turned out to be running on.
	screenWidth,
	screenHeight int
}

var _ ebiten.Game = (*Manager)(nil)

// A resizer is a scene that arranges itself around the screen it is given
// rather than sitting at one fixed size. The manager tells it how big the
// screen is before every update, and again the moment it takes over, so a
// screen is never drawn or hit-tested against a size it has not seen.
//
// It is optional: a scene that looks the same on every device says nothing and
// gets asked nothing.
type resizer interface {
	resize(width, height int)
}

// NewManager starts the game on first, in a screen width by height pixels. That
// size is the game's own — what the levels are laid out in and what the menus
// are drawn at — and stays the smallest the screen can be, whatever device it
// ends up on.
func NewManager(width, height int, first Scene) *Manager {
	m := &Manager{current: first, width: width, height: height, screenWidth: width, screenHeight: height}
	m.resizeCurrent()
	return m
}

// Update ticks the running scene and hands over to whatever scene it returns.
// The scene taking over is first updated on the next tick, so the key press that
// ended one scene is never read a second time by the one that follows it.
func (m *Manager) Update() error {
	m.resizeCurrent()

	next, err := m.current.Update()
	if err != nil {
		return err
	}
	if next != nil {
		m.current = next
		// The scene taking over is drawn this same frame, before its own first
		// update, so it is told the size now rather than a frame from now.
		m.resizeCurrent()
	}

	return nil
}

func (m *Manager) resizeCurrent() {
	if scene, ok := m.current.(resizer); ok {
		scene.resize(m.screenWidth, m.screenHeight)
	}
}

func (m *Manager) Draw(screen *ebiten.Image) {
	m.current.Draw(screen)
}

// Layout gives the game the shape of the screen it is actually on. The height is
// held at the size the game was drawn for, so everything keeps the proportions
// it was laid out in; the width follows the device's own aspect ratio.
//
// This is what makes a phone work. A handset in landscape is roughly twice as
// wide as it is tall, against the 4:3 the lot is drawn at, and holding the game
// to 4:3 there would either letterbox away a third of the screen or stretch the
// cars. Following the width instead leaves that extra as open ground either
// side of the lot, which is exactly where the thumbs are and exactly where the
// on-screen controls go — so the controls cost the player none of the race.
//
// It never returns less than the width the game was written at, so a portrait
// or square window letterboxes rather than cutting the lot off the sides.
func (m *Manager) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if outsideWidth <= 0 || outsideHeight <= 0 {
		return m.screenWidth, m.screenHeight // nothing to go on; keep what we had
	}

	width := int(math.Round(float64(m.height) * float64(outsideWidth) / float64(outsideHeight)))
	if width < m.width {
		width = m.width
	}

	m.screenWidth, m.screenHeight = width, m.height
	return width, m.height
}
