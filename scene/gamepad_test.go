package scene

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// The stick has to be pushed properly before it counts, so resting a thumb on it
// does not steer.
func TestStickNeedsAProperPush(t *testing.T) {
	for _, tc := range []struct {
		name string
		x, y float64
		want stick
	}{
		{name: "at rest", x: 0, y: 0},
		{name: "barely moved", x: 0.2, y: -0.2},
		{name: "short of the line", x: 0.49, y: -0.49},
		{name: "pushed left", x: -1, y: 0, want: stick{left: true}},
		{name: "pushed right", x: 1, y: 0, want: stick{right: true}},
		{name: "pushed up", x: 0, y: -1, want: stick{up: true}},
		{name: "pushed down", x: 0, y: 1, want: stick{down: true}},
		{name: "pushed into a corner", x: -1, y: -1, want: stick{up: true, left: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (stick{}).next(tc.x, tc.y); got != tc.want {
				t.Errorf("next(%v, %v) = %+v, want %+v", tc.x, tc.y, got, tc.want)
			}
		})
	}
}

// A stick held near the line keeps whatever it was doing rather than chattering
// between pushed and not, which on a menu would run the focus up and down on its
// own.
func TestStickHoldsOnPastTheLineItCrossed(t *testing.T) {
	pushed := (stick{}).next(1, 0)
	if !pushed.right {
		t.Fatalf("a full push did not register: %+v", pushed)
	}

	if easedOff := pushed.next(0.4, 0); !easedOff.right {
		t.Errorf("easing back to 0.4 let go of a push that started at 1: %+v", easedOff)
	}
	if released := pushed.next(0.3, 0); released.right {
		t.Errorf("coming back to 0.3 stayed pushed: %+v", released)
	}

	// The same 0.4 that holds a push on is not enough to start one.
	if fromRest := (stick{}).next(0.4, 0); fromRest.right {
		t.Errorf("0.4 from rest counted as a push: %+v", fromRest)
	}
}

// The stick answers for the d-pad, so a binding written for the d-pad is driven
// by either. Nothing else on the pad is the stick's to answer for.
func TestStickAnswersForTheDPadOnly(t *testing.T) {
	pushed := stick{up: true, down: true, left: true, right: true}

	for _, tc := range []struct {
		name   string
		button ebiten.StandardGamepadButton
	}{
		{name: "up", button: ebiten.StandardGamepadButtonLeftTop},
		{name: "down", button: ebiten.StandardGamepadButtonLeftBottom},
		{name: "left", button: ebiten.StandardGamepadButtonLeftLeft},
		{name: "right", button: ebiten.StandardGamepadButtonLeftRight},
	} {
		t.Run(tc.name, func(t *testing.T) {
			held, isDirection := pushed.direction(tc.button)
			if !isDirection {
				t.Fatalf("%v is not one of the directions the stick stands in for", tc.name)
			}
			if !held {
				t.Errorf("%v was pushed but did not report held", tc.name)
			}
		})
	}

	for _, button := range []ebiten.StandardGamepadButton{
		ebiten.StandardGamepadButtonRightBottom,
		ebiten.StandardGamepadButtonCenterRight,
		ebiten.StandardGamepadButtonFrontBottomRight,
	} {
		if _, isDirection := pushed.direction(button); isDirection {
			t.Errorf("button %v was taken for a stick direction", button)
		}
	}
}
