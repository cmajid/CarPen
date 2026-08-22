package scene

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// DesignWidth and DesignHeight are the size the game is laid out at: the
// coordinate space the levels, the menus and the sprites are all written in.
// It is not the size of anything on a screen — Layout takes it to whatever
// shape the device turns out to be, and Ebitengine scales the result to fill
// the glass — so it is written down here once and nowhere else.
const (
	DesignWidth  = 640
	DesignHeight = 480
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

	// pointsPerPixel is what one game pixel turned out to be on the device,
	// worked out in Layout and handed to the scenes with the size. It is zero
	// until Layout has been called for the first time.
	pointsPerPixel float64
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
	resize(v viewport)
}

// viewport is the screen a scene lays itself out on: how big it is in the
// game's own pixels, and how big one of those pixels is on the glass.
//
// The second half is what a phone and a tablet disagree about. The game is
// always 480 pixels tall whatever it is running on, so on a handset one game
// pixel is about eight tenths of a point and on an iPad it is over two — which
// means a control written down as so many game pixels comes out half as wide
// again on the tablet, and a stick sized to a thumb on a phone ends up a
// quarter of the screen on an iPad. Anything a finger has to land on is sized
// through this rather than in game pixels alone.
type viewport struct {
	width, height int

	// pointsPerPixel is how much of the device's own measure — points on a
	// phone or a tablet, logical pixels in a window — one game pixel is drawn
	// across. Zero means nothing is known about the device yet, which is where
	// the game stands until Ebiten first calls Layout.
	pointsPerPixel float64
}

// size works out how big one part of a touch control should be, in game pixels.
// It is the smaller of two answers: points, the size the thing has to be on the
// glass, which is what a thumb measures in and is the same on every device; and
// fraction of the screen's height, the size it was drawn at, which stops a
// control swallowing a screen that has few points to give.
//
// A device that has said nothing about itself gets the drawing's own size,
// which is what every screen got before there was a device in the picture.
func (v viewport) size(points, fraction float64) float64 {
	drawn := fraction * float64(v.height)
	if v.pointsPerPixel <= 0 {
		return drawn
	}
	return math.Min(points/v.pointsPerPixel, drawn)
}

// NewManager starts the game on first, in a screen of the design size. That
// size is the game's own — what the levels are laid out in and what the menus
// are drawn at — and stays the smallest the screen can be, whatever device it
// ends up on.
func NewManager(first Scene) *Manager {
	m := &Manager{current: first, width: DesignWidth, height: DesignHeight, screenWidth: DesignWidth, screenHeight: DesignHeight}
	m.resizeCurrent()
	return m
}

// Update ticks the running scene and hands over to whatever scene it returns.
// The scene taking over is first updated on the next tick, so the key press that
// ended one scene is never read a second time by the one that follows it.
func (m *Manager) Update() error {
	// Fullscreen is the window's business rather than any scene's, so the
	// toggle lives here with the rest of what Ebiten drives directly. On a
	// device with no window to grow — a phone, a tablet — the key never comes
	// and the call would do nothing anyway.
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

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
		scene.resize(viewport{
			width:          m.screenWidth,
			height:         m.screenHeight,
			pointsPerPixel: m.pointsPerPixel,
		})
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
// or square window letterboxes rather than cutting the lot off the sides. A
// tablet is close enough to the lot's own 4:3 that it comes out at very nearly
// the 640x480 the game is drawn in, filling the screen with no bars and no
// stretching — and with no ground to spare either side, which is what the
// controls in touch.go have to answer for.
//
// It also records what one game pixel is on the device (see viewport), which is
// the only place that can be known: outsideHeight is the one number the
// platform gives us in the player's own units rather than the game's.
func (m *Manager) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if outsideWidth <= 0 || outsideHeight <= 0 {
		return m.screenWidth, m.screenHeight // nothing to go on; keep what we had
	}

	width := int(math.Round(float64(m.height) * float64(outsideWidth) / float64(outsideHeight)))
	if width < m.width {
		width = m.width
	}

	m.screenWidth, m.screenHeight = width, m.height

	// The outside size is in the device's own measure — points on a phone or a
	// tablet — and the height is the axis the game is pinned to, so the two
	// together are the scale everything a thumb has to hit is sized by.
	m.pointsPerPixel = float64(outsideHeight) / float64(m.height)

	return width, m.height
}
