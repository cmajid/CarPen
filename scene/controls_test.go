package scene

import (
	"testing"

	"github.com/cmajid/carpen/carpen"
	"github.com/hajimehoshi/ebiten/v2"
)

// The whole game can be played on a pad: the same things the keyboard reaches,
// reached with a pad instead. These are the pad's half of the tests that already
// cover the keys, written against what the player means rather than against the
// binding table, so a rebinding that broke the game would still be caught.

// The driving buttons reach the car the player is driving, and only that one.
func TestGameplayDrivesTheActiveCarOnThePad(t *testing.T) {
	for _, tc := range []struct {
		name   string
		button ebiten.StandardGamepadButton
		held   func(c carpen.Car) bool
	}{
		{name: "right trigger", button: ebiten.StandardGamepadButtonFrontBottomRight, held: func(c carpen.Car) bool { return c.Accelerate }},
		{name: "left trigger", button: ebiten.StandardGamepadButtonFrontBottomLeft, held: func(c carpen.Car) bool { return c.Decelerate }},
		{name: "stick or d-pad left", button: ebiten.StandardGamepadButtonLeftLeft, held: func(c carpen.Car) bool { return c.RotateLeft }},
		{name: "stick or d-pad right", button: ebiten.StandardGamepadButtonLeftRight, held: func(c carpen.Car) bool { return c.RotateRight }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := newFakeInput()
			game := newGameplay(in, testLevel(t))

			in.pressPad(tc.button)
			if _, err := game.Update(); err != nil {
				t.Fatalf("Update: %v", err)
			}

			if !tc.held(game.cars[game.activeCar]) {
				t.Errorf("%v did not reach the active car", tc.name)
			}
			if tc.held(game.cars[1]) {
				t.Errorf("%v reached the car the player is not driving", tc.name)
			}
		})
	}
}

// Going and stopping are the triggers' alone. The stick is for steering, and
// pushing it forward is not a request to pull away — that is the difference
// between driving a pad and driving a keyboard held sideways.
func TestGameplayStickDoesNotDriveOnThePad(t *testing.T) {
	for _, tc := range []struct {
		name   string
		button ebiten.StandardGamepadButton
	}{
		{name: "stick or d-pad up", button: ebiten.StandardGamepadButtonLeftTop},
		{name: "stick or d-pad down", button: ebiten.StandardGamepadButtonLeftBottom},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := newFakeInput()
			game := newGameplay(in, testLevel(t))

			in.pressPad(tc.button)
			if _, err := game.Update(); err != nil {
				t.Fatalf("Update: %v", err)
			}

			car := game.cars[game.activeCar]
			if car.Accelerate {
				t.Errorf("%s accelerated; only the right trigger should", tc.name)
			}
			if car.Decelerate {
				t.Errorf("%s braked; only the left trigger should", tc.name)
			}
		})
	}
}

// The same stick that must not drive still moves a menu, because a menu has
// nothing else for it to do.
func TestMenuListStillMovesOnTheStick(t *testing.T) {
	list := newTestList()
	in := newFakeInput()

	in.pressPad(ebiten.StandardGamepadButtonLeftBottom)
	step(list, in)
	if list.selected != 1 {
		t.Errorf("the stick did not move the menu: selected = %d, want 1", list.selected)
	}
}

// Letting go of a driving button lets go of the car, so a race left with the
// pad at rest does not drive on by itself.
func TestGameplayLetsGoOfTheCarOnThePad(t *testing.T) {
	in := newFakeInput()
	game := newGameplay(in, testLevel(t))

	in.pressPad(ebiten.StandardGamepadButtonFrontBottomRight)
	if _, err := game.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !game.cars[game.activeCar].Accelerate {
		t.Fatal("the trigger did not get the car moving")
	}

	in.releasePad(ebiten.StandardGamepadButtonFrontBottomRight)
	if _, err := game.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if game.cars[game.activeCar].Accelerate {
		t.Error("the car kept accelerating after the trigger came up")
	}
}

// Start is the pause button on every other game, so it is the pause button here.
func TestGameplayPausesOnThePad(t *testing.T) {
	for _, tc := range []struct {
		name   string
		button ebiten.StandardGamepadButton
	}{
		{name: "Start", button: ebiten.StandardGamepadButtonCenterRight},
		{name: "B", button: ebiten.StandardGamepadButtonRightRight},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := newFakeInput()
			game := newGameplay(in, testLevel(t))

			in.pressPad(tc.button)
			next, err := game.Update()
			if err != nil {
				t.Fatalf("Update: %v", err)
			}
			if _, ok := next.(*Pause); !ok {
				t.Errorf("%s handed over to %T, want *Pause", tc.name, next)
			}
		})
	}
}

func TestGameplayFinishesAndSwapsCarsOnThePad(t *testing.T) {
	in := newFakeInput()
	game := newGameplay(in, testLevel(t))

	// X swaps on the way up, the way Tab does, so holding it does not run
	// through the cars.
	in.releasePad(ebiten.StandardGamepadButtonRightLeft)
	if _, err := game.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if game.activeCar != 1 {
		t.Errorf("after X, active car = %d, want 1", game.activeCar)
	}

	in.pressPad(ebiten.StandardGamepadButtonRightTop)
	next, err := game.Update()
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if _, ok := next.(*Results); !ok {
		t.Errorf("Y handed over to %T, want *Results", next)
	}
}

// The menus are worked with the pad the same way they are worked with the keys:
// the d-pad or stick moves the focus, A chooses.
func TestMenuListMovesAndChoosesOnThePad(t *testing.T) {
	list := newTestList()
	in := newFakeInput()

	in.pressPad(ebiten.StandardGamepadButtonLeftBottom)
	if chosen := step(list, in); chosen != nothingChosen {
		t.Fatalf("moving the focus chose %d, want nothing", chosen)
	}
	if list.selected != 1 {
		t.Fatalf("after d-pad down selected = %d, want 1", list.selected)
	}

	in.pressPad(ebiten.StandardGamepadButtonLeftTop)
	step(list, in)
	if list.selected != 0 {
		t.Fatalf("after d-pad up selected = %d, want 0", list.selected)
	}

	in.pressPad(ebiten.StandardGamepadButtonRightBottom)
	if chosen := step(list, in); chosen != 0 {
		t.Errorf("A chose %d, want 0", chosen)
	}
}

// Backing out is B on the pad wherever it is Esc on the keyboard.
func TestMenuQuitsOnB(t *testing.T) {
	in := newFakeInput()
	menu := NewMenu(in, testLevel(t))

	in.pressPad(ebiten.StandardGamepadButtonRightRight)
	if _, err := menu.Update(); err != ebiten.Termination {
		t.Errorf("B on the menu gave err = %v, want ebiten.Termination", err)
	}
}

// The prompts name what the player is actually holding. A pad in their hands and
// "Esc" on the screen is a prompt that helps nobody.
func TestPromptsFollowTheDeviceInTheirHands(t *testing.T) {
	in := newFakeInput()

	if got := hint(in, "Esc", "B"); got != "Esc" {
		t.Errorf("with no pad, hint = %q, want %q", got, "Esc")
	}
	if got := promptFor(in, "Enter", "A", "Select"); got != (prompt{key: "Enter", does: "Select"}) {
		t.Errorf("with no pad, promptFor = %+v", got)
	}

	in.padOn = true

	if got := hint(in, "Esc", "B"); got != "B" {
		t.Errorf("with a pad, hint = %q, want %q", got, "B")
	}
	if got := promptFor(in, "Enter", "A", "Select"); got != (prompt{key: "A", does: "Select"}) {
		t.Errorf("with a pad, promptFor = %+v", got)
	}
}
