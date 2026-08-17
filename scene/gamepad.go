package scene

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// gamepad is the pad the game is being played on, read through Ebiten's
// standard layout. That layout is what makes one set of bindings enough for
// every pad: Ebiten carries SDL's controller database and maps whatever is
// plugged in onto the same buttons, so an Xbox pad's A, a DualSense's cross and
// a Switch Pro's B all arrive as StandardGamepadButtonRightBottom.
//
// Only a pad the database knows is picked up. An unmapped pad's buttons are in
// no particular order, and guessing at them would be worse than the keyboard
// the player still has.
type gamepad struct {
	id    ebiten.GamepadID
	found bool

	// ids is the scan buffer, kept so that looking for the pad every tick does
	// not allocate.
	ids []ebiten.GamepadID

	// now and was are the left stick read as a d-pad, this tick and last. The
	// stick is an axis, and an axis has no press to report, so the edge the
	// screens are written against is worked out here.
	now, was stick

	// tick is the tick these were last worked out on. Ebiten's own tick count
	// is what marks the boundary, because the stick has to move on once per
	// tick however many times the scenes ask about it — a scene reading both
	// the press and the release of one direction asks twice.
	tick int64
}

func newGamepad() gamepad {
	// Ebiten's tick count starts at 0, so the state has to be staler than that
	// to be refreshed on the very first one.
	return gamepad{tick: -1}
}

// How far the stick has to be pushed to count as a press, and how far it has to
// come back to count as a release. The two differ so that a stick held near the
// line does not chatter a menu up and down.
const (
	stickPress   = 0.5
	stickRelease = 0.35
)

// stick is the left stick reported as the four directions of a d-pad.
type stick struct{ up, down, left, right bool }

// next is where the stick stands after a tick that read it at x, y. Ebiten
// reports both axes in −1..1 with up and left negative, following the web
// gamepad standard.
func (s stick) next(x, y float64) stick {
	return stick{
		up:    pushed(-y, s.up),
		down:  pushed(y, s.down),
		left:  pushed(-x, s.left),
		right: pushed(x, s.right),
	}
}

// direction reports how the stick stands in for one of the d-pad's buttons, and
// whether the button is one of the four it stands in for at all.
func (s stick) direction(button ebiten.StandardGamepadButton) (held, isDirection bool) {
	switch button {
	case ebiten.StandardGamepadButtonLeftTop:
		return s.up, true
	case ebiten.StandardGamepadButtonLeftBottom:
		return s.down, true
	case ebiten.StandardGamepadButtonLeftLeft:
		return s.left, true
	case ebiten.StandardGamepadButtonLeftRight:
		return s.right, true
	}
	return false, false
}

// pushed reports whether an axis is pushed far enough this way to count, giving
// it a little further to come back than it took to get there.
func pushed(value float64, wasPushed bool) bool {
	if wasPushed {
		return value >= stickRelease
	}
	return value >= stickPress
}

func (g *gamepad) connected() bool {
	g.refresh()
	return g.found
}

// justPressed reports the tick a button goes down. The four d-pad buttons
// answer for the left stick as well, so a binding written for the d-pad is
// driven by the stick without being written twice.
func (g *gamepad) justPressed(button ebiten.StandardGamepadButton) bool {
	g.refresh()
	if !g.found {
		return false
	}
	if inpututil.IsStandardGamepadButtonJustPressed(g.id, button) {
		return true
	}

	now, isDirection := g.now.direction(button)
	was, _ := g.was.direction(button)
	return isDirection && now && !was
}

// justReleased reports the tick a button comes back up.
func (g *gamepad) justReleased(button ebiten.StandardGamepadButton) bool {
	g.refresh()
	if !g.found {
		return false
	}
	if inpututil.IsStandardGamepadButtonJustReleased(g.id, button) {
		return true
	}

	now, isDirection := g.now.direction(button)
	was, _ := g.was.direction(button)
	return isDirection && was && !now
}

// refresh looks for the pad and moves the stick on, once per tick however often
// it is called.
func (g *gamepad) refresh() {
	tick := ebiten.Tick()
	if tick == g.tick {
		return
	}
	g.tick = tick

	g.find()
	if !g.found {
		g.now, g.was = stick{}, stick{}
		return
	}

	g.was = g.now
	g.now = g.now.next(
		ebiten.StandardGamepadAxisValue(g.id, ebiten.StandardGamepadAxisLeftStickHorizontal),
		ebiten.StandardGamepadAxisValue(g.id, ebiten.StandardGamepadAxisLeftStickVertical),
	)
}

// find settles on the first pad Ebiten can read in the standard layout. It runs
// every tick rather than once at start-up, because a pad is plugged in and
// unplugged while the game is running; the ids come back in order, so the pad
// the game is on only changes when that pad itself goes away.
func (g *gamepad) find() {
	g.ids = ebiten.AppendGamepadIDs(g.ids[:0])

	var first ebiten.GamepadID
	any := false

	for _, id := range g.ids {
		if !ebiten.IsStandardGamepadLayoutAvailable(id) {
			continue
		}
		if g.found && id == g.id {
			return // still holding the one already being played on
		}
		if !any {
			first, any = id, true
		}
	}

	g.id, g.found = first, any
}
