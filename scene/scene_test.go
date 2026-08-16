package scene

import (
	"errors"
	"testing"

	"github.com/cmajid/carpen/carpen"
	"github.com/hajimehoshi/ebiten/v2"
)

// fakeInput is the keyboard and mouse the tests work the game with. A real
// keyboard reports a key as just pressed for the one tick it goes down, so a
// test sets the keys for the tick it is about to run and they are gone by the
// next one.
type fakeInput struct {
	pressed  map[ebiten.Key]bool
	released map[ebiten.Key]bool
	cursorX  int
	cursorY  int
	clicked  bool
}

func newFakeInput() *fakeInput {
	return &fakeInput{pressed: map[ebiten.Key]bool{}, released: map[ebiten.Key]bool{}}
}

func (f *fakeInput) IsKeyJustPressed(key ebiten.Key) bool { return f.pressed[key] }

func (f *fakeInput) IsKeyJustReleased(key ebiten.Key) bool { return f.released[key] }

func (f *fakeInput) CursorPosition() (int, int) { return f.cursorX, f.cursorY }

func (f *fakeInput) IsMouseButtonJustPressed(ebiten.MouseButton) bool { return f.clicked }

// press makes keys the only ones going down on the next tick.
func (f *fakeInput) press(keys ...ebiten.Key) {
	f.clear()
	for _, key := range keys {
		f.pressed[key] = true
	}
}

// release makes keys the only ones coming up on the next tick.
func (f *fakeInput) release(keys ...ebiten.Key) {
	f.clear()
	for _, key := range keys {
		f.released[key] = true
	}
}

// moveTo puts the mouse somewhere, as a player moving it would. Where the mouse
// is outlives a tick; what was pressed on that tick does not.
func (f *fakeInput) moveTo(x, y int) {
	f.cursorX, f.cursorY = x, y
}

func (f *fakeInput) click() {
	f.clicked = true
}

func (f *fakeInput) clear() {
	clear(f.pressed)
	clear(f.released)
	f.clicked = false
}

// stubScene stands in for a real screen, so the manager can be tested on its own:
// it hands over whatever it is told to and counts what it was asked to do.
type stubScene struct {
	next    Scene
	err     error
	updates int
	draws   int
}

func (s *stubScene) Update() (Scene, error) {
	s.updates++
	return s.next, s.err
}

func (s *stubScene) Draw(*ebiten.Image) {
	s.draws++
}

// tick runs one update of the manager with keys held down for that tick.
func tick(t *testing.T, m *Manager, in *fakeInput, keys ...ebiten.Key) {
	t.Helper()

	in.press(keys...)
	if err := m.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}
	in.clear()
}

// choose walks the focus down to the item at index and presses Enter on it, the
// way a player picks something out of a menu.
func choose(t *testing.T, m *Manager, in *fakeInput, index int) {
	t.Helper()

	for i := 0; i < index; i++ {
		tick(t, m, in, ebiten.KeyDown)
	}
	tick(t, m, in, ebiten.KeyEnter)
}

func TestManagerStaysOnASceneThatReturnsNoNext(t *testing.T) {
	first := &stubScene{}
	m := NewManager(640, 480, first)

	for i := 0; i < 3; i++ {
		if err := m.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}

	if m.current != Scene(first) {
		t.Errorf("current scene = %#v, want the scene it started on", m.current)
	}
	if first.updates != 3 {
		t.Errorf("scene updated %d times, want 3", first.updates)
	}
}

// The scene taking over waits a tick before its first update. Both scenes read
// the same keyboard, so a scene updated in the same tick as the one it replaced
// would see the very key press that ended that scene and act on it too.
func TestManagerUpdatesTheNewSceneOnTheNextTickOnly(t *testing.T) {
	second := &stubScene{}
	m := NewManager(640, 480, &stubScene{next: second})

	if err := m.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if m.current != Scene(second) {
		t.Fatalf("current scene = %#v, want the scene the first one returned", m.current)
	}
	if second.updates != 0 {
		t.Errorf("new scene updated %d times on the tick it took over, want 0", second.updates)
	}

	if err := m.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if second.updates != 1 {
		t.Errorf("new scene updated %d times after a further tick, want 1", second.updates)
	}
}

func TestManagerReportsASceneError(t *testing.T) {
	wantErr := errors.New("scene broke")
	first := &stubScene{next: &stubScene{}, err: wantErr}
	m := NewManager(640, 480, first)

	err := m.Update()

	if !errors.Is(err, wantErr) {
		t.Errorf("Update error = %v, want %v", err, wantErr)
	}
	if m.current != Scene(first) {
		t.Errorf("current scene = %#v, want the game left on the scene that failed", m.current)
	}
}

func TestManagerDrawsTheRunningScene(t *testing.T) {
	first := &stubScene{}
	m := NewManager(640, 480, first)

	m.Draw(nil)

	if first.draws != 1 {
		t.Errorf("running scene drawn %d times, want 1", first.draws)
	}
}

func TestManagerLayoutIsFixed(t *testing.T) {
	m := NewManager(640, 480, &stubScene{})

	width, height := m.Layout(1920, 1080)

	if width != 640 || height != 480 {
		t.Errorf("Layout = %dx%d, want 640x480 whatever the window size", width, height)
	}
}

// The cycle the game is built around: the menu starts a race, a finished race
// shows its results, and the results lead back to the menu.
func TestMenuGameplayResultsCycle(t *testing.T) {
	in := newFakeInput()
	m := NewManager(640, 480, NewMenu(in))

	choose(t, m, in, menuStart)
	if _, ok := m.current.(*Gameplay); !ok {
		t.Fatalf("after starting from the menu the game is on %T, want *Gameplay", m.current)
	}

	tick(t, m, in, ebiten.KeyEnter)
	if _, ok := m.current.(*Results); !ok {
		t.Fatalf("after finishing the race the game is on %T, want *Results", m.current)
	}

	choose(t, m, in, resultsMenu)
	if _, ok := m.current.(*Menu); !ok {
		t.Fatalf("after choosing the main menu the game is on %T, want *Menu", m.current)
	}
}

// Esc means "back" on every screen it appears on, which is what makes it worth
// printing once along the bottom of all of them.
func TestEscapeGoesBack(t *testing.T) {
	t.Run("results back to the menu", func(t *testing.T) {
		in := newFakeInput()
		m := NewManager(640, 480, newResults(in))

		tick(t, m, in, ebiten.KeyEscape)

		if _, ok := m.current.(*Menu); !ok {
			t.Errorf("game is on %T, want *Menu", m.current)
		}
	})

	t.Run("pause back to the race", func(t *testing.T) {
		in := newFakeInput()
		game := newGameplay(in)
		m := NewManager(640, 480, game)

		tick(t, m, in, ebiten.KeyEscape)
		tick(t, m, in, ebiten.KeyEscape)

		if m.current != Scene(game) {
			t.Errorf("game is on %#v, want the race it paused", m.current)
		}
	})

	t.Run("menu ends the game", func(t *testing.T) {
		in := newFakeInput()
		m := NewManager(640, 480, NewMenu(in))

		in.press(ebiten.KeyEscape)

		if err := m.Update(); !errors.Is(err, ebiten.Termination) {
			t.Errorf("Update error = %v, want ebiten.Termination", err)
		}
	})
}

// Quitting from the menu closes the window, which Ebiten asks for as Termination
// rather than as a failure.
func TestMenuQuitEndsTheGame(t *testing.T) {
	in := newFakeInput()
	m := NewManager(640, 480, NewMenu(in))

	in.press(ebiten.KeyDown)
	if err := m.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}
	in.press(ebiten.KeyEnter)

	if err := m.Update(); !errors.Is(err, ebiten.Termination) {
		t.Errorf("Update error = %v, want ebiten.Termination", err)
	}
}

// Pausing and resuming has to come back to the same race, not to a fresh one.
func TestPauseResumesTheSameRace(t *testing.T) {
	in := newFakeInput()
	game := newGameplay(in)
	m := NewManager(640, 480, game)

	tick(t, m, in, ebiten.KeyEscape)
	paused, ok := m.current.(*Pause)
	if !ok {
		t.Fatalf("after Esc the game is on %T, want *Pause", m.current)
	}
	if paused.resume != game {
		t.Errorf("pause holds a different race than the one that was running")
	}

	choose(t, m, in, pauseResume)
	if m.current != Scene(game) {
		t.Errorf("after resuming the game is on %#v, want the race it paused", m.current)
	}
}

// A paused race is a still one: the cars only move on a tick of the gameplay
// scene, and while the pause overlay is the running scene there are none.
func TestPauseFreezesTheCars(t *testing.T) {
	in := newFakeInput()
	game := newGameplay(in)
	m := NewManager(640, 480, game)

	tick(t, m, in, ebiten.KeyEscape)
	before := game.cars[0].Pivot

	for i := 0; i < 10; i++ {
		tick(t, m, in)
	}

	if game.cars[0].Pivot != before {
		t.Errorf("car moved from %v to %v while paused, want it held still", before, game.cars[0].Pivot)
	}
}

// Pausing lets go of the wheel. Driving is worked out from presses and releases,
// so a key released during the pause is never seen coming up, and the car would
// otherwise carry on accelerating after resuming with nothing held down.
func TestPauseReleasesHeldControls(t *testing.T) {
	in := newFakeInput()
	game := newGameplay(in)
	m := NewManager(640, 480, game)

	tick(t, m, in, ebiten.KeyUp, ebiten.KeyLeft)
	if !game.cars[0].Accelerate || !game.cars[0].RotateLeft {
		t.Fatal("holding Up and Left did not reach the car")
	}

	tick(t, m, in, ebiten.KeyEscape)

	if car := game.cars[0]; car.Accelerate || car.Decelerate || car.RotateLeft || car.RotateRight {
		t.Errorf("car still driving itself after the pause: %+v", car)
	}
}

func TestPauseQuitReturnsToTheMenu(t *testing.T) {
	in := newFakeInput()
	m := NewManager(640, 480, newGameplay(in))

	tick(t, m, in, ebiten.KeyEscape)
	choose(t, m, in, pauseQuit)

	if _, ok := m.current.(*Menu); !ok {
		t.Errorf("after quitting the pause the game is on %T, want *Menu", m.current)
	}
}

// Restarting from the pause menu deals a fresh world rather than dropping the
// player back into the one they were stuck in.
func TestPauseRestartDealsAFreshRace(t *testing.T) {
	in := newFakeInput()
	first := newGameplay(in)
	m := NewManager(640, 480, first)

	// Drive for a while, so a race that carried on would be somewhere else.
	for i := 0; i < 20; i++ {
		tick(t, m, in)
	}
	tick(t, m, in, ebiten.KeyEscape)
	choose(t, m, in, pauseRestart)

	second, ok := m.current.(*Gameplay)
	if !ok {
		t.Fatalf("the game is on %T, want *Gameplay", m.current)
	}
	if second == first {
		t.Fatal("restarting carried on the same race, want a fresh one")
	}
	if second.cars[0].Pivot != newGameplay(in).cars[0].Pivot {
		t.Errorf("restarted race starts at %v, want where a first race starts", second.cars[0].Pivot)
	}
}

// The driving keys reach the car the player is driving, and only that one.
func TestGameplayDrivesTheActiveCar(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  ebiten.Key
		held func(c carpen.Car) bool
	}{
		{name: "up", key: ebiten.KeyUp, held: func(c carpen.Car) bool { return c.Accelerate }},
		{name: "down", key: ebiten.KeyDown, held: func(c carpen.Car) bool { return c.Decelerate }},
		{name: "left", key: ebiten.KeyLeft, held: func(c carpen.Car) bool { return c.RotateLeft }},
		{name: "right", key: ebiten.KeyRight, held: func(c carpen.Car) bool { return c.RotateRight }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := newFakeInput()
			game := newGameplay(in)

			in.press(tc.key)
			if _, err := game.Update(); err != nil {
				t.Fatalf("Update: %v", err)
			}

			if !tc.held(game.cars[game.activeCar]) {
				t.Errorf("%v did not reach the active car", tc.key)
			}
			if tc.held(game.cars[1]) {
				t.Errorf("%v reached the car the player is not driving", tc.key)
			}
		})
	}
}

func TestGameplayTabSwapsTheActiveCar(t *testing.T) {
	in := newFakeInput()
	game := newGameplay(in)

	for _, want := range []int{1, 0} {
		in.release(ebiten.KeyTab)
		if _, err := game.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}

		if game.activeCar != want {
			t.Errorf("active car = %d, want %d", game.activeCar, want)
		}
	}
}

// The cars are stepped from Update, at Ebiten's fixed tick rate, and never from
// the draw path.
func TestGameplayStepsTheCarsEachTick(t *testing.T) {
	in := newFakeInput()
	game := newGameplay(in)
	before := game.cars[0].Pivot

	if _, err := game.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if game.cars[0].Pivot == before {
		t.Errorf("car did not move on a tick: still at %v", before)
	}
}
