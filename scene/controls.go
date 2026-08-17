package scene

import "github.com/hajimehoshi/ebiten/v2"

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
// The four d-pad buttons are also how the left stick arrives (see gamepad.go),
// so binding the d-pad binds the stick with it — which is why steering names
// only the d-pad and gets the stick for nothing.
type binding struct {
	keys    []ebiten.Key
	buttons []ebiten.StandardGamepadButton
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
	actionSteerLeft: {
		keys:    []ebiten.Key{ebiten.KeyLeft},
		buttons: []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftLeft},
	},
	actionSteerRight: {
		keys:    []ebiten.Key{ebiten.KeyRight},
		buttons: []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftRight},
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
