package scene

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func newTestList() *menuList {
	return newMenuList(100, 200, 220, "Start Race", "Options", "Quit")
}

// step runs one tick of the list and reports what the player chose on it.
// Whatever was set up for the tick is spent by it, the way a real keyboard
// reports a key as just pressed for one tick and no longer.
func step(list *menuList, in *fakeInput) int {
	chosen := list.update(in)
	in.clear()
	return chosen
}

// The focus starts on the first item, so the thing a player most likely came for
// is one key away.
func TestMenuListStartsOnTheFirstItem(t *testing.T) {
	if got := newTestList().selected; got != 0 {
		t.Errorf("selected = %d, want 0", got)
	}
}

func TestMenuListMovesWithTheArrowKeys(t *testing.T) {
	list := newTestList()
	in := newFakeInput()

	in.press(ebiten.KeyDown)
	step(list, in)
	if list.selected != 1 {
		t.Fatalf("after Down selected = %d, want 1", list.selected)
	}

	in.press(ebiten.KeyUp)
	step(list, in)
	if list.selected != 0 {
		t.Errorf("after Up selected = %d, want 0", list.selected)
	}
}

// A linear menu is expected to come back round at either end, so the whole of it
// can be reached by holding one direction (Xbox Accessibility Guideline 112).
func TestMenuListWrapsAtBothEnds(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  ebiten.Key
		from int
		want int
	}{
		{name: "down past the last item", key: ebiten.KeyDown, from: 2, want: 0},
		{name: "up past the first item", key: ebiten.KeyUp, from: 0, want: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			list := newTestList()
			list.selected = tc.from
			in := newFakeInput()

			in.press(tc.key)
			step(list, in)

			if list.selected != tc.want {
				t.Errorf("selected = %d, want %d", list.selected, tc.want)
			}
		})
	}
}

func TestMenuListTabMovesToTheNextItem(t *testing.T) {
	list := newTestList()
	in := newFakeInput()

	in.press(ebiten.KeyTab)
	step(list, in)

	if list.selected != 1 {
		t.Errorf("after Tab selected = %d, want 1", list.selected)
	}
}

func TestMenuListChoosesWithEnterOrSpace(t *testing.T) {
	for _, key := range []ebiten.Key{ebiten.KeyEnter, ebiten.KeySpace} {
		t.Run(key.String(), func(t *testing.T) {
			list := newTestList()
			list.selected = 2
			in := newFakeInput()

			in.press(key)

			if got := step(list, in); got != 2 {
				t.Errorf("chose %d, want 2", got)
			}
		})
	}
}

func TestMenuListChoosesNothingWithoutAKey(t *testing.T) {
	list := newTestList()
	in := newFakeInput()

	if got := step(list, in); got != nothingChosen {
		t.Errorf("chose %d, want nothing chosen", got)
	}
}

// The mouse is the other way through every menu, so a player never has to work
// out which of their hands the game wants.
func TestMenuListFollowsTheMouse(t *testing.T) {
	list := newTestList()
	in := newFakeInput()
	step(list, in) // the first tick only teaches the list where the mouse is

	row := list.bounds(2)
	in.moveTo(row.Min.X+10, row.Min.Y+10)
	step(list, in)
	if list.selected != 2 {
		t.Fatalf("hovering the third row selected %d, want 2", list.selected)
	}

	in.click()
	if got := step(list, in); got != 2 {
		t.Errorf("clicking the third row chose %d, want 2", got)
	}
}

func TestMenuListIgnoresAClickBesideIt(t *testing.T) {
	list := newTestList()
	in := newFakeInput()
	step(list, in)

	in.moveTo(10, 10) // well clear of the list
	in.click()

	if got := step(list, in); got != nothingChosen {
		t.Errorf("clicking beside the list chose %d, want nothing chosen", got)
	}
}

// A pointer lying still over the list must not keep dragging the focus back,
// because that would make the arrow keys look broken.
func TestMenuListMouseOnlyTakesFocusWhenItMoves(t *testing.T) {
	list := newTestList()
	in := newFakeInput()

	row := list.bounds(0)
	in.moveTo(row.Min.X+10, row.Min.Y+10)
	step(list, in) // the mouse arrives on the first row
	step(list, in)

	in.press(ebiten.KeyDown)
	step(list, in)
	if list.selected != 1 {
		t.Fatalf("Down moved to %d, want 1", list.selected)
	}

	step(list, in) // the mouse has not moved since

	if list.selected != 1 {
		t.Errorf("a still mouse pulled the focus back to %d, want it left on 1", list.selected)
	}
}
