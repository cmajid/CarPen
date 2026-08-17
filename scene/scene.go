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

// Input is the keyboard, the mouse and the gamepad as a scene reads them.
// Ebiten reports all three through globals that only a running game fills in, so
// scenes take them through this interface instead, and the tests can work the
// menus without a window.
//
// This is the devices as they are, not what a press means: which button drives
// and which one pauses is decided once in controls.go, so a scene asks for an
// action rather than for a key.
type Input interface {
	IsKeyJustPressed(key ebiten.Key) bool
	IsKeyJustReleased(key ebiten.Key) bool
	CursorPosition() (x, y int)
	IsMouseButtonJustPressed(button ebiten.MouseButton) bool

	// IsKeyPressed is the key as it stands rather than the moment it moved. A
	// menu cares about the press; the accelerator cares about being held, and
	// working that out from the two edges is how a key released on another
	// screen used to leave the car driving itself.
	IsKeyPressed(key ebiten.Key) bool

	// The pad is read in Ebiten's standard layout, which is the one place the
	// difference between an Xbox pad, a DualSense and a Switch Pro controller
	// is dealt with: all three arrive here as the same buttons.
	IsGamepadButtonJustPressed(button ebiten.StandardGamepadButton) bool
	IsGamepadButtonJustReleased(button ebiten.StandardGamepadButton) bool

	// GamepadButtonValue and GamepadAxisValue are the pad's analogue half: how
	// far a trigger is pulled, in 0..1, and how far a stick is pushed, in
	// −1..1. A button that is only ever down or up answers 0 or 1, so the same
	// reading serves the d-pad.
	GamepadButtonValue(button ebiten.StandardGamepadButton) float64
	GamepadAxisValue(axis ebiten.StandardGamepadAxis) float64

	// GamepadConnected reports whether there is a pad to read at all, which is
	// what the screens write their prompts for.
	GamepadConnected() bool
}

// Devices is the real keyboard, mouse and gamepad — the Input the game itself
// runs on.
type Devices struct {
	pad gamepad
}

// NewDevices opens the player's hardware. One of these is made at start-up and
// handed down through the scenes, because the pad keeps track of what the stick
// was doing last tick and a fresh one every scene would forget it.
func NewDevices() *Devices {
	return &Devices{pad: newGamepad()}
}

func (*Devices) IsKeyJustPressed(key ebiten.Key) bool { return inpututil.IsKeyJustPressed(key) }

func (*Devices) IsKeyJustReleased(key ebiten.Key) bool { return inpututil.IsKeyJustReleased(key) }

func (*Devices) IsKeyPressed(key ebiten.Key) bool { return ebiten.IsKeyPressed(key) }

func (*Devices) CursorPosition() (int, int) { return ebiten.CursorPosition() }

func (*Devices) IsMouseButtonJustPressed(button ebiten.MouseButton) bool {
	return inpututil.IsMouseButtonJustPressed(button)
}

func (d *Devices) IsGamepadButtonJustPressed(button ebiten.StandardGamepadButton) bool {
	return d.pad.justPressed(button)
}

func (d *Devices) IsGamepadButtonJustReleased(button ebiten.StandardGamepadButton) bool {
	return d.pad.justReleased(button)
}

func (d *Devices) GamepadButtonValue(button ebiten.StandardGamepadButton) float64 {
	return d.pad.buttonValue(button)
}

func (d *Devices) GamepadAxisValue(axis ebiten.StandardGamepadAxis) float64 {
	return d.pad.axisValue(axis)
}

func (d *Devices) GamepadConnected() bool { return d.pad.connected() }
