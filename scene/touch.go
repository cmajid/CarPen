package scene

import (
	"math"

	"github.com/cmajid/carpen/carpen"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// The controls a player gets when the screen is the only thing they have. A
// keyboard and a pad are already in the player's hands, so binding them is a
// table (controls.go); a touch screen has nothing on it until the game draws
// something there, so these controls are a place on the glass and a shape on
// top of it, and the two have to agree. That is the whole of what this file is
// for: one set of circles, hit-tested in update and drawn in draw, so a button
// cannot end up somewhere other than where it looks.
//
// Nothing here is a new way of driving. Each control answers for an action the
// game already had, and the values it produces are the ones a trigger and a
// stick produce — 0..1 for the pedals, −1..1 for the steering — so gameplay
// reads a finger through the same lines that read a pad.

// touchPointer is the touch screen standing in for the mouse, which is what
// lets the menus work on a phone without knowing that a phone is what they are
// on. It is the device half of touch, and belongs to Devices; the on-screen
// driving controls below are a screen's own business.
type touchPointer struct {
	// ids and began are scan buffers, kept so that reading the screen every
	// tick does not allocate.
	ids   []ebiten.TouchID
	began []ebiten.TouchID

	// x and y are where the screen was last touched, which outlives the touch
	// itself — see Devices.CursorPosition.
	x, y int

	// touched latches the first time the player touches the screen at all, and
	// is what the on-screen controls are shown on.
	touched bool

	// live is whether a finger is down this very tick.
	live bool

	// tick is the tick this was last worked out on, so that the screens asking
	// several times over in one tick read it once. Ebiten's own count marks the
	// boundary, as it does for the pad.
	tick int64
}

func newTouchPointer() touchPointer {
	// Ebiten's tick count starts at 0, so the state has to be staler than that
	// to be refreshed on the very first one.
	return touchPointer{tick: -1}
}

func (t *touchPointer) refresh() {
	tick := ebiten.Tick()
	if tick == t.tick {
		return
	}
	t.tick = tick

	t.ids = ebiten.AppendTouchIDs(t.ids[:0])
	t.began = inpututil.AppendJustPressedTouchIDs(t.began[:0])
	t.live = len(t.ids) > 0

	// A finger that has just arrived is the one being pointed with, even if
	// others are already down: the tap that lands on a menu row is the newest
	// one, not whichever happens to be first in the list.
	switch {
	case len(t.began) > 0:
		t.x, t.y = ebiten.TouchPosition(t.began[0])
		t.touched = true
	case t.live:
		t.x, t.y = ebiten.TouchPosition(t.ids[0])
		t.touched = true
	}
}

// position is where the player is pointing, and whether touch is what they are
// pointing with at all.
func (t *touchPointer) position() (x, y int, touched bool) {
	t.refresh()
	return t.x, t.y, t.touched
}

// justTouched reports the tick a finger arrives on the glass — the tap, as
// opposed to the holding down that follows it.
func (t *touchPointer) justTouched() bool {
	t.refresh()
	return len(t.began) > 0
}

func (t *touchPointer) active() bool {
	t.refresh()
	return t.touched
}

// touchCircle is one control's place on the screen. Circles rather than
// rectangles because a thumb lands in a round patch and because the distance
// from the middle is what the stick is read by anyway.
type touchCircle struct {
	x, y, radius float64
}

func (c touchCircle) contains(x, y float64) bool {
	dx, dy := x-c.x, y-c.y
	return dx*dx+dy*dy <= c.radius*c.radius
}

// The on-screen controls are sized and placed as fractions of the screen's
// height, never in pixels: the game is laid out at a fixed height and a width
// that follows the device (see Manager.Layout), so the height is the one number
// that means the same thing on every screen the game runs on. A phone, a tablet
// and a desktop window each get controls the same size relative to the game,
// and the wider the device the further apart the two thumbs sit.
//
// The sizes are floors set by the thumb rather than by the drawing: the pedals
// come out around 110 game pixels across, which on a phone in landscape is
// comfortably past the 44pt Apple asks a touch target to be, with the smaller
// row along the top still clearing it.
const (
	touchMargin    = 0.05  // from the screen's edge to the nearest control
	touchGap       = 0.035 // between one control and the next
	touchStickSize = 0.16  // the steering stick's base
	touchPedalSize = 0.115 // the two pedals under the right thumb
	touchBarSize   = 0.06  // the row of smaller buttons along the top

	// touchStickReach is how much wider than it looks the stick's catchment is.
	// A thumb going for the stick lands near it rather than on it, and a stick
	// that has to be hit exactly is a stick that is missed; the ring is drawn
	// where the steering is measured from, and the area that takes hold of it
	// is larger than the ring.
	touchStickReach = 1.6

	// touchKnobSize is the knob's radius as a fraction of the base's.
	touchKnobSize = 0.42
)

// touchLayout is where every on-screen control sits, worked out once each time
// the screen changes size rather than per tick.
type touchLayout struct {
	stick touchCircle
	slots [numTouchSlots]touchCircle
}

// newTouchLayout places the controls on a screen width by height. The stick
// goes bottom left and the pedals bottom right, which is where the thumbs
// already are when a phone is held in two hands; the buttons that are pressed
// between manoeuvres rather than during them go along the top, out of the way
// of both.
func newTouchLayout(width, height int) touchLayout {
	w, h := float64(width), float64(height)

	margin := h * touchMargin
	gap := h * touchGap
	stick := h * touchStickSize
	pedal := h * touchPedalSize
	bar := h * touchBarSize

	l := touchLayout{
		stick: touchCircle{x: margin + stick, y: h - margin - stick, radius: stick},
	}

	// The pedals are stacked rather than set side by side. Two of them abreast
	// are wider than the ground a phone leaves beside the lot, which would put
	// the second one over the very car being parked; stacked, they take one
	// column and both stay out on the margin. The corner is the throttle, being
	// the one held for most of a race; backing up is the reach above it.
	near := touchCircle{x: w - margin - pedal, y: h - margin - pedal, radius: pedal}
	l.slots[slotPedalNear] = near
	l.slots[slotPedalFar] = touchCircle{x: near.x, y: near.y - 2*pedal - gap, radius: pedal}

	// The top row is laid out from the right edge inwards, so the order the
	// slots are named in is the order they appear.
	top := hudHeight + margin*0.6 + bar
	for i, slot := range [...]touchSlot{slotBarRight, slotBarMiddle, slotBarLeft} {
		l.slots[slot] = touchCircle{x: w - margin - bar - float64(i)*(2*bar+gap), y: top, radius: bar}
	}

	return l
}

// touchControls is the screen being driven with: which of the drawn controls
// are under a finger, and how far the stick has been pushed. One of these
// belongs to the race, because the race is the only screen that needs more from
// a touch than where it landed — everywhere else, a tap is a click and the
// menus were already built for that.
type touchControls struct {
	layout touchLayout

	// now and was are which actions the screen is being asked for, this tick
	// and last. A drawn button has no press of its own to report, so the edges
	// the scenes are written against are worked out here, the same way the
	// stick's are in gamepad.go.
	now, was [numActions]bool

	// steering is the stick's lean, −1..1 with left negative, in the same terms
	// a real stick's axis arrives in.
	steering float64

	// stickID is the finger that took hold of the stick, and stickHeld whether
	// one has. The finger is remembered rather than looked for again each tick,
	// so that steering hard enough to drag it outside the ring keeps steering
	// instead of letting go halfway through the turn.
	stickID   ebiten.TouchID
	stickHeld bool

	// ids is the scan buffer, kept so that reading the screen every tick does
	// not allocate.
	ids []ebiten.TouchID
}

// resize places the controls on a screen of a new size. It is called from the
// scene's own resize rather than worked out while drawing, because where the
// controls are is what update hit-tests against, and a control that is drawn in
// one place and pressed in another is worse than no control at all.
func (t *touchControls) resize(width, height int) {
	t.layout = newTouchLayout(width, height)
}

// update reads every finger on the screen onto the controls, once per tick. It
// has to run before anything asks what the player is doing, because everything
// below answers out of what it works out here.
func (t *touchControls) update(in Input) {
	t.was, t.now = t.now, [numActions]bool{}
	t.ids = in.AppendTouchIDs(t.ids[:0])

	t.updateStick(in)

	for _, id := range t.ids {
		if t.stickHeld && id == t.stickID {
			continue // the steering finger is not also pressing whatever it passes over
		}

		x, y := t.positionOf(in, id)
		for _, button := range touchButtons {
			if t.layout.slots[button.slot].contains(x, y) {
				t.now[button.action] = true
			}
		}
	}
}

// updateStick works out how far the steering is being asked over, and takes
// hold of a finger to ask with if it has not got one.
func (t *touchControls) updateStick(in Input) {
	if t.stickHeld && !t.holds(t.stickID) {
		t.stickHeld = false // the finger came off the glass
	}

	if !t.stickHeld {
		catchment := touchCircle{x: t.layout.stick.x, y: t.layout.stick.y, radius: t.layout.stick.radius * touchStickReach}
		for _, id := range t.ids {
			if x, y := t.positionOf(in, id); catchment.contains(x, y) {
				t.stickID, t.stickHeld = id, true
				break
			}
		}
	}

	if !t.stickHeld {
		// No self-centring is needed here the way it is not needed on a stick:
		// letting go asks for nothing, and Car.Steer leaves the wheels where
		// they were pointed rather than straightening them.
		t.steering = 0
		return
	}

	// Only how far across the finger is counts. The stick steers and does
	// nothing else — the same split that keeps the pad's left stick off the
	// accelerator — so how far up or down it has wandered is not asking for
	// anything.
	x, _ := t.positionOf(in, t.stickID)
	lean := clamp((x-t.layout.stick.x)/t.layout.stick.radius, -1, 1)
	t.steering = math.Copysign(beyondDeadzone(math.Abs(lean)), lean)
}

// holds reports whether id is still one of the fingers on the screen.
func (t *touchControls) holds(id ebiten.TouchID) bool {
	for _, other := range t.ids {
		if other == id {
			return true
		}
	}
	return false
}

func (t *touchControls) positionOf(in Input, id ebiten.TouchID) (x, y float64) {
	px, py := in.TouchPosition(id)
	return float64(px), float64(py)
}

// value is how far the screen is being asked for a, in the 0..1 every other
// device answers in. The stick's lean is split back into the two steering
// actions, because an action is one way along an axis and the bindings table
// names them separately.
func (t *touchControls) value(a action) float64 {
	switch a {
	case actionSteerLeft:
		return math.Max(0, -t.steering)
	case actionSteerRight:
		return math.Max(0, t.steering)
	}

	if t.now[a] {
		return 1 // a drawn button is down or it is not, like a key
	}
	return 0
}

// analog, justPressed and justReleased are the package's own three, with the
// screen folded in. A scene that can be played by touch reads through these
// instead, and gets the keyboard and the pad along with them: whichever device
// is asking hardest wins, so nothing is lost by a phone that also has a pad
// paired to it.
func (t *touchControls) analog(in Input, a action) float64 {
	return math.Max(analog(in, a), t.value(a))
}

func (t *touchControls) justPressed(in Input, a action) bool {
	return justPressed(in, a) || (t.now[a] && !t.was[a])
}

func (t *touchControls) justReleased(in Input, a action) bool {
	return justReleased(in, a) || (t.was[a] && !t.now[a])
}

// draw paints the controls. They are drawn rather than drawn from images, which
// keeps them sharp at whatever size the device asks for and keeps them looking
// like the rest of the game: the same palette, the same accent for the thing
// being pressed.
func (t *touchControls) draw(dst *ebiten.Image) {
	base := t.layout.stick
	fillCircle(dst, base.x, base.y, base.radius, colourTouch)
	strokeCircle(dst, carpen.Vector{X: base.x, Y: base.y}, base.radius, 2, colourTouchLine)

	// The knob sits where the steering is, which is what makes the control
	// readable: how far it is off centre is how hard the car is turning.
	knob, colour := base.radius*touchKnobSize, colourTouchKnob
	if t.stickHeld {
		colour = colourAccent
	}
	fillCircle(dst, base.x+t.steering*base.radius, base.y, knob, colour)

	for _, button := range touchButtons {
		circle := t.layout.slots[button.slot]

		fill, face, ink := colourTouch, text.Face(fontPrompt), colourText
		if t.now[button.action] {
			// Pressed is shown by filling with the accent rather than by
			// growing or moving: a thumb is over the control at the moment it
			// matters, and what it has not covered still has to change.
			fill, face, ink = colourAccent, fontPromptOn, colourAccentInk
		}

		fillCircle(dst, circle.x, circle.y, circle.radius, fill)
		strokeCircle(dst, carpen.Vector{X: circle.x, Y: circle.y}, circle.radius, 2, colourTouchLine)
		drawText(dst, button.label, face, circle.x, circle.y, ink, text.AlignCenter, text.AlignCenter)
	}
}

func clamp(value, low, high float64) float64 {
	return math.Min(math.Max(value, low), high)
}
