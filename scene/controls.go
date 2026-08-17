package scene

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// An action is something the player means to do, named for the meaning rather
// than for the thing pressed. Screens ask for actions, so what drives the car is
// written down once here instead of at every place that reads it — and adding
// the pad meant adding a column to this table rather than an "or" to twenty
// conditions.
//
// Actions are named for the screen they belong to where the same key means two
// things: Tab moves down a menu but swaps cars in a race, so those are two
// actions and not one.
type action int

const (
	actionMenuNext action = iota
	actionMenuPrev
	actionConfirm
	actionCancel
	actionThrottle
	actionBrake
	actionSteerLeft
	actionSteerRight
	actionSwapCar
	actionFinish
	actionDebugBoxes
)

// binding is everything that means one action. The pad's buttons are given in
// Ebiten's standard layout, whose names say where a button sits rather than what
// any one maker prints on it: RightBottom is A on an Xbox pad, cross on a
// DualSense, B on a Switch Pro.
//
// The four d-pad buttons are also how the left stick arrives as a press (see
// gamepad.go), so binding the d-pad binds the stick with it. That is enough for
// a menu, which only wants to know that the stick moved; steering wants to know
// how far, and names the axis as well.
type binding struct {
	keys    []ebiten.Key
	buttons []ebiten.StandardGamepadButton
	axis    axisBinding
}

// axisBinding is one way of an analogue axis: the axis itself, and which way
// along it counts. A stick has one axis for two actions, so steering left takes
// the horizontal axis negated and steering right takes it as it comes.
type axisBinding struct {
	axis  ebiten.StandardGamepadAxis
	sign  float64
	bound bool
}

func towards(axis ebiten.StandardGamepadAxis, sign float64) axisBinding {
	return axisBinding{axis: axis, sign: sign, bound: true}
}

var bindings = [...]binding{
	actionMenuNext: {
		keys:    []ebiten.Key{ebiten.KeyDown, ebiten.KeyTab},
		buttons: []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftBottom},
	},
	actionMenuPrev: {
		keys:    []ebiten.Key{ebiten.KeyUp},
		buttons: []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftTop},
	},
	actionConfirm: {
		keys:    []ebiten.Key{ebiten.KeyEnter, ebiten.KeySpace},
		buttons: []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightBottom},
	},
	actionCancel: {
		keys: []ebiten.Key{ebiten.KeyEscape},
		buttons: []ebiten.StandardGamepadButton{
			ebiten.StandardGamepadButtonRightRight,  // B: back a step
			ebiten.StandardGamepadButtonCenterRight, // Start: the pause button everywhere else
		},
	},
	// On a pad the car is driven the way every racing game drives one: the
	// triggers do the going and the stopping, and the stick is left doing
	// nothing but steering. Pushing the stick forward deliberately does not
	// pull away — a pad splits driving across both hands, and having the same
	// thumb both aim and accelerate is what makes a pad feel like a keyboard
	// with worse arrow keys.
	//
	// The keyboard keeps Up and Down, which are all it has.
	actionThrottle: {
		keys:    []ebiten.Key{ebiten.KeyUp},
		buttons: []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonFrontBottomRight}, // RT
	},
	actionBrake: {
		keys:    []ebiten.Key{ebiten.KeyDown},
		buttons: []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonFrontBottomLeft}, // LT
	},
	// Steering is where the stick's travel earns its keep, so it is named
	// outright rather than left to the d-pad it stands in for elsewhere: a
	// stick leant on gently asks for a gentle angle, where the d-pad beside it
	// has only full lock to ask for.
	actionSteerLeft: {
		keys:    []ebiten.Key{ebiten.KeyLeft},
		buttons: []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftLeft},
		axis:    towards(ebiten.StandardGamepadAxisLeftStickHorizontal, -1),
	},
	actionSteerRight: {
		keys:    []ebiten.Key{ebiten.KeyRight},
		buttons: []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftRight},
		axis:    towards(ebiten.StandardGamepadAxisLeftStickHorizontal, 1),
	},
	actionSwapCar: {
		keys:    []ebiten.Key{ebiten.KeyTab},
		buttons: []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightLeft}, // X
	},
	actionFinish: {
		keys:    []ebiten.Key{ebiten.KeyEnter},
		buttons: []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightTop}, // Y
	},
	// The box overlay is a development tool and stays off the pad: it is not
	// worth a button a player might land on by accident.
	actionDebugBoxes: {
		keys: []ebiten.Key{ebiten.KeyF3},
	},
}

// analogDeadzone is how far a trigger or a stick has to leave its rest before
// it is taken to mean anything, and what is left is stretched back over the
// whole 0..1 so that full travel still asks for everything. A pad at rest does
// not sit at a clean zero — a worn stick wanders, and a trigger's spring stops
// a shade short — and without this the car would creep off on a pad nobody is
// touching.
//
// It is larger than it needs to be for a new pad on purpose: the cost of too
// much is a small dead patch at the start of the travel, and the cost of too
// little is a car that will not stand still.
const analogDeadzone = 0.15

// analog reports how far the player is asking for a, from 0 to 1, on whichever
// device they are asking with. A held key is the whole way: a keyboard has only
// the two ends, and pretending otherwise would make the arrows worse than they
// were. Where more than one device asks at once, the one asking hardest wins,
// so a hand resting on a stick cannot hold back the keys.
func analog(in Input, a action) float64 {
	for _, key := range bindings[a].keys {
		if in.IsKeyPressed(key) {
			return 1
		}
	}

	value := 0.0
	for _, button := range bindings[a].buttons {
		value = math.Max(value, beyondDeadzone(in.GamepadButtonValue(button)))
	}
	if axis := bindings[a].axis; axis.bound {
		value = math.Max(value, beyondDeadzone(axis.sign*in.GamepadAxisValue(axis.axis)))
	}
	return value
}

// beyondDeadzone is how much of an analogue reading counts, once the dead patch
// around rest is taken off and the rest is stretched back out to 0..1.
func beyondDeadzone(value float64) float64 {
	if value <= analogDeadzone {
		return 0
	}
	return math.Min((value-analogDeadzone)/(1-analogDeadzone), 1)
}

// justPressed reports the tick the player asks for a, on whichever device they
// asked with.
func justPressed(in Input, a action) bool {
	for _, key := range bindings[a].keys {
		if in.IsKeyJustPressed(key) {
			return true
		}
	}
	for _, button := range bindings[a].buttons {
		if in.IsGamepadButtonJustPressed(button) {
			return true
		}
	}
	return false
}

// justReleased reports the tick the player stops asking for a. Where an action
// has more than one thing bound to it, letting go of any of them is letting go:
// nobody drives with the stick and the arrow keys at once, and taking the last
// one released to be the one that counts would mean remembering which were down.
func justReleased(in Input, a action) bool {
	for _, key := range bindings[a].keys {
		if in.IsKeyJustReleased(key) {
			return true
		}
	}
	for _, button := range bindings[a].buttons {
		if in.IsGamepadButtonJustReleased(button) {
			return true
		}
	}
	return false
}
