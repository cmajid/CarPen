package scene

import (
	"image"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// at is the middle of a control, which is where a test puts a finger when it
// means to press that control and nothing else.
func at(c touchCircle) image.Point {
	return image.Pt(int(c.x), int(c.y))
}

// newTouchControls is a set of controls laid out for a screen, which is the
// state they are always in by the time anything reads them: the manager sizes a
// scene before it updates it.
func newTouchControls(width, height int) *touchControls {
	t := &touchControls{}
	t.resize(width, height)
	return t
}

// The stick is the reason steering is a stick and not two buttons: how far
// across the finger is, is how hard the car turns. Full travel asks for
// everything, half travel asks for about half, and the rest near the middle is
// the deadzone every other analogue control on the game has.
func TestTouchStickSteersByHowFarItIsPushed(t *testing.T) {
	// A finger lands on a whole pixel and the stick's middle falls between two,
	// so what comes back is within a pixel's worth of the lean asked for rather
	// than exactly it. A pixel of steering is not something a thumb can aim at
	// anyway; the tolerance is what one pixel is worth.
	const closeEnough = 1 / (0.16 * 480)

	for _, tc := range []struct {
		name   string
		across float64 // how far along the stick's radius the finger lands
		left   float64
		right  float64
	}{
		{"pushed past the ring, which is full lock", 1.4, 0, 1},
		{"pushed to the ring itself", 1, 0, 1},
		{"leant halfway right", 0.5, 0, (0.5 - analogDeadzone) / (1 - analogDeadzone)},
		{"pushed past the ring to the left", -1.4, 1, 0},
		{"leant halfway left", -0.5, (0.5 - analogDeadzone) / (1 - analogDeadzone), 0},
		{"resting in the middle", 0, 0, 0},
		{"barely off the middle, inside the deadzone", 0.1, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := newFakeInput()
			touch := newTouchControls(640, 480)
			stick := touch.layout.stick

			in.touchAt(image.Pt(int(stick.x+tc.across*stick.radius), int(stick.y)))
			touch.update(in)

			left := touch.analog(in, actionSteerLeft)
			right := touch.analog(in, actionSteerRight)

			if math.Abs(left-tc.left) > closeEnough || math.Abs(right-tc.right) > closeEnough {
				t.Errorf("steering left %.3f right %.3f, want left %.3f right %.3f", left, right, tc.left, tc.right)
			}
		})
	}
}

// A thumb that has taken hold of the stick keeps it, wherever it goes from
// there. Steering hard drags the thumb off the ring it started on, and a stick
// that let go at the edge would straighten the wheels halfway through the very
// turn that needed the most lock.
func TestTouchStickKeepsTheFingerThatTookIt(t *testing.T) {
	in := newFakeInput()
	touch := newTouchControls(640, 480)
	stick := touch.layout.stick

	in.touchAt(at(stick))
	touch.update(in)

	// Far outside the ring, and further out than the catchment that would have
	// taken hold of it in the first place.
	in.touchAt(image.Pt(int(stick.x+stick.radius*3), int(stick.y)))
	touch.update(in)

	if got := touch.analog(in, actionSteerRight); got != 1 {
		t.Errorf("steering right %.3f after dragging past the ring, want 1 — the finger still holds the stick", got)
	}

	in.lift()
	touch.update(in)

	if got := touch.analog(in, actionSteerRight); got != 0 {
		t.Errorf("steering right %.3f after lifting the finger, want 0", got)
	}
}

// Reaching for the stick is not the same as hitting it. The catchment is wider
// than the ring is drawn, so a thumb that lands near it steers rather than
// doing nothing at all.
func TestTouchStickIsCaughtFromBesideIt(t *testing.T) {
	in := newFakeInput()
	touch := newTouchControls(640, 480)
	stick := touch.layout.stick

	// Outside the ring, inside the catchment.
	in.touchAt(image.Pt(int(stick.x+stick.radius*1.3), int(stick.y)))
	touch.update(in)

	if got := touch.analog(in, actionSteerRight); got != 1 {
		t.Errorf("steering right %.3f for a thumb landing just outside the ring, want 1", got)
	}
}

// Only how far across the stick the finger is counts. A thumb travels in an arc
// and wanders up and down the stick without meaning anything by it, and the
// pedals are the other hand's job — the same split that keeps the pad's left
// stick off the accelerator.
func TestTouchStickIgnoresHowFarUpTheFingerIs(t *testing.T) {
	in := newFakeInput()
	touch := newTouchControls(640, 480)
	stick := touch.layout.stick

	// Far enough across to be full lock, and well up the stick from the middle —
	// but still inside the catchment, which is a circle and so is reached less
	// far on the diagonal.
	in.touchAt(image.Pt(int(stick.x+stick.radius*1.2), int(stick.y-stick.radius*0.5)))
	touch.update(in)

	if got := touch.analog(in, actionSteerRight); got != 1 {
		t.Errorf("steering right %.3f, want 1 whatever height the thumb is at", got)
	}
	if got := touch.analog(in, actionThrottle); got != 0 {
		t.Errorf("throttle %.3f, want 0 — the stick does not drive", got)
	}
}

// A drawn button is asked for while a finger is on it, and stops being asked
// for the moment the finger leaves.
func TestTouchButtonsAskForTheirAction(t *testing.T) {
	in := newFakeInput()
	touch := newTouchControls(640, 480)

	for _, button := range touchButtons {
		in.touchAt(at(touch.layout.slots[button.slot]))
		touch.update(in)

		if got := touch.value(button.action); got != 1 {
			t.Errorf("%s: asked for %.3f while held, want 1", button.label, got)
		}

		in.lift()
		touch.update(in)

		if got := touch.value(button.action); got != 0 {
			t.Errorf("%s: still asked for %.3f after the finger came off, want 0", button.label, got)
		}
	}
}

// A held button is one press, not one per tick. Pausing is on a drawn button,
// and a button that pressed itself sixty times a second would open the pause
// menu and close it again for as long as the thumb rested there.
func TestTouchButtonPressesOnceWhileHeld(t *testing.T) {
	in := newFakeInput()
	touch := newTouchControls(640, 480)
	pause := touch.layout.slots[slotBarRight]

	in.touchAt(at(pause))
	touch.update(in)
	if !touch.justPressed(in, actionCancel) {
		t.Fatal("the tick the finger arrives is not reported as a press")
	}

	in.touchAt(at(pause))
	touch.update(in)
	if touch.justPressed(in, actionCancel) {
		t.Error("a finger still resting on the button is reported as a second press")
	}

	in.lift()
	touch.update(in)
	if !touch.justReleased(in, actionCancel) {
		t.Error("lifting the finger is not reported as a release")
	}
}

// Both thumbs at once, which is the whole point of a screen laid out for two.
// Steering while accelerating is most of driving, and a control scheme that
// took only one finger at a time would make it impossible.
func TestTouchStickAndPedalWorkTogether(t *testing.T) {
	in := newFakeInput()
	touch := newTouchControls(640, 480)
	stick := touch.layout.stick

	in.touchAt(
		image.Pt(int(stick.x-stick.radius), int(stick.y)),
		at(touch.layout.slots[slotPedalNear]),
	)
	touch.update(in)

	if got := touch.analog(in, actionSteerLeft); got != 1 {
		t.Errorf("steering left %.3f with the other thumb on the throttle, want 1", got)
	}
	if got := touch.analog(in, actionThrottle); got != 1 {
		t.Errorf("throttle %.3f while steering, want 1", got)
	}
}

// The whole way through: a thumb on the drawn throttle drives the car the race
// is actually being played with, not a copy of the controls kept beside it.
func TestTouchDrivesTheRace(t *testing.T) {
	in := newFakeInput()
	g := newGameplay(in, testLevel(t))
	g.resize(640, 480)

	in.touchAt(at(g.touch.layout.slots[slotPedalNear]))
	if _, err := g.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if got := g.cars[g.activeCar].Throttle; got != 1 {
		t.Errorf("throttle %.3f with a thumb on the pedal, want 1", got)
	}

	in.lift()
	if _, err := g.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if got := g.cars[g.activeCar].Throttle; got != 0 {
		t.Errorf("throttle %.3f after the thumb came off, want 0", got)
	}
}

// The pause button on the screen is the Esc key, and reaches the same screen.
func TestTouchPauseButtonPauses(t *testing.T) {
	in := newFakeInput()
	g := newGameplay(in, testLevel(t))
	g.resize(640, 480)

	in.touchAt(at(g.touch.layout.slots[slotBarRight]))
	next, err := g.Update()
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if _, ok := next.(*Pause); !ok {
		t.Errorf("tapping the pause button went to %T, want *Pause", next)
	}
}

// Swapping happens when the button is let go of rather than when it goes down,
// the same as on a key and a pad — and the car being handed over is let go of
// first, which is the bug that left a car flat out with nothing to lift it.
func TestTouchSwapButtonSwapsOnRelease(t *testing.T) {
	in := newFakeInput()
	g := newGameplay(in, testLevel(t))
	g.resize(640, 480)

	in.touchAt(at(g.touch.layout.slots[slotBarLeft]))
	if _, err := g.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if g.activeCar != 0 {
		t.Fatalf("the car swapped on the way down, before the thumb came off")
	}

	in.lift()
	if _, err := g.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if g.activeCar != 1 {
		t.Errorf("driving car %d after the swap, want 1", g.activeCar)
	}
}

// The screen is a third device, not a replacement for the other two. A phone
// with a pad paired to it has both, and whichever is asking hardest wins — the
// rule the keys and the pad already settle their disagreements by.
func TestTouchDoesNotShutOutTheOtherDevices(t *testing.T) {
	in := newFakeInput()
	touch := newTouchControls(640, 480)

	in.holdPad(0.6, ebiten.StandardGamepadButtonFrontBottomRight)
	touch.update(in)

	if got := touch.analog(in, actionThrottle); math.Abs(got-beyondDeadzone(0.6)) > 0.001 {
		t.Errorf("throttle %.3f from a trigger held part way, want the trigger's own %.3f", got, beyondDeadzone(0.6))
	}
}

// The controls are placed against the screen rather than written down in
// pixels, so a device with more width than the game was drawn for puts the
// thumbs at its own edges and not at the edges of a 640-wide window.
func TestTouchControlsFollowTheScreen(t *testing.T) {
	narrow := newTouchLayout(640, 480)
	wide := newTouchLayout(1041, 480) // a phone in landscape

	if narrow.stick != wide.stick {
		t.Errorf("the stick moved from %+v to %+v; it is measured from the left edge, which did not move", narrow.stick, wide.stick)
	}

	gained := float64(1041 - 640)
	if moved := wide.slots[slotPedalNear].x - narrow.slots[slotPedalNear].x; math.Abs(moved-gained) > 0.001 {
		t.Errorf("the throttle moved %.1f for %.1f of extra width; it is measured from the right edge and should have moved with it", moved, gained)
	}
}

// What the extra width on a phone is for. The lot is centred in it and the
// controls sit out on the ground either side, so a player driving with their
// thumbs is not driving with them on top of the car they are parking.
func TestTouchControlsClearTheLotOnAPhone(t *testing.T) {
	const (
		screenWidth = 1041 // a handset in landscape, per Manager.Layout
		lotWidth    = 640
	)

	layout := newTouchLayout(screenWidth, 480)
	lotLeft := float64(screenWidth-lotWidth) / 2
	lotRight := lotLeft + lotWidth

	if right := layout.stick.x + layout.stick.radius; right > lotLeft {
		t.Errorf("the stick reaches %.1f, over a lot starting at %.1f", right, lotLeft)
	}
	for _, slot := range []touchSlot{slotPedalNear, slotPedalFar} {
		if left := layout.slots[slot].x - layout.slots[slot].radius; left < lotRight {
			t.Errorf("a pedal reaches back to %.1f, over a lot ending at %.1f", left, lotRight)
		}
	}
}

// The controls have to be big enough to hit. Apple asks for 44pt, and the game
// is scaled to the device by its height, so the check is made in the points a
// handset in landscape actually gives us.
func TestTouchControlsAreBigEnoughToHit(t *testing.T) {
	const (
		minimumPoints = 44
		deviceHeight  = 393.0 // an iPhone in landscape, in points
		gameHeight    = 480.0
	)

	scale := deviceHeight / gameHeight
	layout := newTouchLayout(1041, gameHeight)

	for _, button := range touchButtons {
		if points := 2 * layout.slots[button.slot].radius * scale; points < minimumPoints {
			t.Errorf("%s is %.1fpt across on a phone, under the %dpt a touch target has to be", button.label, points, minimumPoints)
		}
	}
}
