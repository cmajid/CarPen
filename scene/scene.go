// Package scene holds the screens the game moves between — the menu, the race,
// the pause overlay, the results — and the manager that switches between them.
// main only opens the window and hands it a manager; everything the player sees
// lives here.
package scene

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// A Scene is one screen of the game. Update returns the scene to hand over to,
// or nil to stay on this one, and returns an error to stop the game —
// ebiten.Termination to close the window without reporting a failure.
type Scene interface {
	Update() (next Scene, err error)
	Draw(screen *ebiten.Image)
}

// Input is the keyboard and mouse as a scene reads them. Ebiten reports both
// through globals that only a running game fills in, so scenes take them through
// this interface instead, and the tests can work the menus without a window.
type Input interface {
	IsKeyJustPressed(key ebiten.Key) bool
	IsKeyJustReleased(key ebiten.Key) bool
	CursorPosition() (x, y int)
	IsMouseButtonJustPressed(button ebiten.MouseButton) bool
}

// Keyboard is the real keyboard and mouse, the Input the game itself runs on.
type Keyboard struct{}

func (Keyboard) IsKeyJustPressed(key ebiten.Key) bool { return inpututil.IsKeyJustPressed(key) }

func (Keyboard) IsKeyJustReleased(key ebiten.Key) bool { return inpututil.IsKeyJustReleased(key) }

func (Keyboard) CursorPosition() (int, int) { return ebiten.CursorPosition() }

func (Keyboard) IsMouseButtonJustPressed(button ebiten.MouseButton) bool {
	return inpututil.IsMouseButtonJustPressed(button)
}
