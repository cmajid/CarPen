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

	// The touch screen, which is the only thing a phone has: every finger down
	// this tick, and where each one is. Unlike a key or a button, a touch
	// carries its own position, so what it means depends on what is drawn under
	// it — which is why these come through raw and are read against the
	// on-screen controls in touch.go rather than against the binding table.
	//
	// The ids are appended to the slice given so that reading them every tick
	// costs no allocation; pass the buffer back in each time.
	AppendTouchIDs(ids []ebiten.TouchID) []ebiten.TouchID
	TouchPosition(id ebiten.TouchID) (x, y int)

	// TouchActive reports whether the player has touched the screen at all yet.
	// It is what the on-screen controls are shown on: a phone has no other way
	// in, and a desktop that has never been touched should not have thumb pads
	// drawn over the lot. It latches on the first touch and stays on, so the
	// controls do not blink away between one press and the next.
	TouchActive() bool
}

// Devices is the real keyboard, mouse, gamepad and touch screen — the Input the
// game itself runs on.
type Devices struct {
	pad     gamepad
	pointer touchPointer
}

// NewDevices opens the player's hardware. One of these is made at start-up and
// handed down through the scenes, because the pad keeps track of what the stick
// was doing last tick and a fresh one every scene would forget it — and the
// touch screen has to remember having been touched at all, which happens on the
// menu and has to still be known by the time the race starts.
func NewDevices() *Devices {
	return &Devices{pad: newGamepad(), pointer: newTouchPointer()}
}

func (*Devices) IsKeyJustPressed(key ebiten.Key) bool { return inpututil.IsKeyJustPressed(key) }

func (*Devices) IsKeyJustReleased(key ebiten.Key) bool { return inpututil.IsKeyJustReleased(key) }

func (*Devices) IsKeyPressed(key ebiten.Key) bool { return ebiten.IsKeyPressed(key) }

// CursorPosition is where the player is pointing. A finger is a pointer the
// menus already know how to read, so once the screen has been touched this
// answers with the finger instead of the mouse: the lists, the hit-testing and
// the focus rules all carry over to a phone without a line of theirs changing.
//
// It keeps answering with the last place touched after the finger lifts, rather
// than falling back to a mouse that is not there. A phone reports no touch far
// more often than it reports one, and a pointer that jumped back to the corner
// between taps would drag the menu's focus with it.
func (d *Devices) CursorPosition() (int, int) {
	if x, y, touched := d.pointer.position(); touched {
		return x, y
	}
	return ebiten.CursorPosition()
}

// IsMouseButtonJustPressed is the click, and a finger arriving on the glass is
// one: a tap and a left click mean the same thing to every screen in the game.
func (d *Devices) IsMouseButtonJustPressed(button ebiten.MouseButton) bool {
	if button == ebiten.MouseButtonLeft && d.pointer.justTouched() {
		return true
	}
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

func (*Devices) AppendTouchIDs(ids []ebiten.TouchID) []ebiten.TouchID {
	return ebiten.AppendTouchIDs(ids)
}

func (*Devices) TouchPosition(id ebiten.TouchID) (int, int) { return ebiten.TouchPosition(id) }

func (d *Devices) TouchActive() bool { return d.pointer.active() }
