package scene

import (
	"errors"
	"image"
	"math"
	"testing"

	"github.com/cmajid/carpen/carpen"
	"github.com/hajimehoshi/ebiten/v2"
)

// fakeInput is the keyboard, mouse and pad the tests work the game with. A real
// keyboard reports a key as just pressed for the one tick it goes down, so a
// test sets the keys for the tick it is about to run and they are gone by the
// next one; the pad's buttons work the same way.
type fakeInput struct {
	pressed  map[ebiten.Key]bool
	released map[ebiten.Key]bool
	cursorX  int
	cursorY  int
	clicked  bool

	padPressed  map[ebiten.StandardGamepadButton]bool
	padReleased map[ebiten.StandardGamepadButton]bool
	padOn       bool

	// padValue and axes are the pad's analogue half — a trigger held part way,
	// a stick leant on. A button pressed without one of these is a button to
	// the floor, which is what every test written before the pad had any travel
	// meant by pressing it.
	padValue map[ebiten.StandardGamepadButton]float64
	axes     map[ebiten.StandardGamepadAxis]float64

	// touches are the fingers on the screen this tick, in the order a real
	// screen would report them. A finger outlives a tick the way a held key
	// does, so a test that wants one held says so for each tick it is down.
	touches []fakeTouch

	// touchOn is the screen having been touched at all, which latches on the
	// real device and latches here.
	touchOn bool
}

// fakeTouch is one finger: which one it is, and where it is resting.
type fakeTouch struct {
	id   ebiten.TouchID
	x, y int
}

func newFakeInput() *fakeInput {
	return &fakeInput{
		pressed:     map[ebiten.Key]bool{},
		released:    map[ebiten.Key]bool{},
		padPressed:  map[ebiten.StandardGamepadButton]bool{},
		padReleased: map[ebiten.StandardGamepadButton]bool{},
		padValue:    map[ebiten.StandardGamepadButton]float64{},
		axes:        map[ebiten.StandardGamepadAxis]float64{},
	}
}

func (f *fakeInput) IsKeyJustPressed(key ebiten.Key) bool { return f.pressed[key] }

func (f *fakeInput) IsKeyJustReleased(key ebiten.Key) bool { return f.released[key] }

// IsKeyPressed is the key still being held on the tick it went down. A test
// says what is held one tick at a time, so the two are the same thing here.
func (f *fakeInput) IsKeyPressed(key ebiten.Key) bool { return f.pressed[key] }

func (f *fakeInput) CursorPosition() (int, int) { return f.cursorX, f.cursorY }

func (f *fakeInput) IsMouseButtonJustPressed(ebiten.MouseButton) bool { return f.clicked }

func (f *fakeInput) IsGamepadButtonJustPressed(b ebiten.StandardGamepadButton) bool {
	return f.padPressed[b]
}

func (f *fakeInput) IsGamepadButtonJustReleased(b ebiten.StandardGamepadButton) bool {
	return f.padReleased[b]
}

func (f *fakeInput) GamepadButtonValue(b ebiten.StandardGamepadButton) float64 {
	if value, held := f.padValue[b]; held {
		return value
	}
	if f.padPressed[b] {
		return 1
	}
	return 0
}

func (f *fakeInput) GamepadAxisValue(a ebiten.StandardGamepadAxis) float64 { return f.axes[a] }

func (f *fakeInput) GamepadConnected() bool { return f.padOn }

func (f *fakeInput) AppendTouchIDs(ids []ebiten.TouchID) []ebiten.TouchID {
	for _, t := range f.touches {
		ids = append(ids, t.id)
	}
	return ids
}

func (f *fakeInput) TouchPosition(id ebiten.TouchID) (int, int) {
	for _, t := range f.touches {
		if t.id == id {
			return t.x, t.y
		}
	}
	return 0, 0
}

func (f *fakeInput) TouchActive() bool { return f.touchOn }

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

// pressPad and releasePad are press and release for a player holding a pad
// instead. Plugging one in is part of pressing it: nothing reports a button
// from a pad that is not there.
func (f *fakeInput) pressPad(buttons ...ebiten.StandardGamepadButton) {
	f.clear()
	f.padOn = true
	for _, button := range buttons {
		f.padPressed[button] = true
	}
}

func (f *fakeInput) releasePad(buttons ...ebiten.StandardGamepadButton) {
	f.clear()
	f.padOn = true
	for _, button := range buttons {
		f.padReleased[button] = true
	}
}

// holdPad presses buttons only part of the way down, the way a trigger rests
// under a finger that has not committed. Pressing them the whole way is what
// pressPad already does.
func (f *fakeInput) holdPad(value float64, buttons ...ebiten.StandardGamepadButton) {
	f.clear()
	f.padOn = true
	for _, button := range buttons {
		f.padValue[button] = value
	}
}

// holdStick leans the left stick somewhere, in Ebiten's −1..1 with up and left
// negative.
func (f *fakeInput) holdStick(x, y float64) {
	f.clear()
	f.padOn = true
	f.axes[ebiten.StandardGamepadAxisLeftStickHorizontal] = x
	f.axes[ebiten.StandardGamepadAxisLeftStickVertical] = y
}

// moveTo puts the mouse somewhere, as a player moving it would. Where the mouse
// is outlives a tick; what was pressed on that tick does not.
func (f *fakeInput) moveTo(x, y int) {
	f.cursorX, f.cursorY = x, y
}

func (f *fakeInput) click() {
	f.clicked = true
}

// touchAt puts fingers on the screen at each of the given points, and they are
// the only ones down for the tick that follows. Touching the screen is part of
// touching it: nothing draws the on-screen controls for a player who has never
// reached for them, so the first finger latches that on here as it does on a
// real screen.
func (f *fakeInput) touchAt(points ...image.Point) {
	f.clear()
	f.touches = f.touches[:0]
	for i, p := range points {
		f.touches = append(f.touches, fakeTouch{id: ebiten.TouchID(i + 1), x: p.X, y: p.Y})
	}
	if len(points) > 0 {
		f.touchOn = true
	}
}

// lift takes every finger off the screen, without unlatching the fact that the
// player is playing by touch.
func (f *fakeInput) lift() { f.touchAt() }

func (f *fakeInput) clear() {
	clear(f.pressed)
	clear(f.released)
	clear(f.padPressed)
	clear(f.padReleased)
	clear(f.padValue)
	clear(f.axes)
	f.clicked = false
}

// testLevel is the level the screens are tested on: the game's own first one,
// so the tests are played in the world the player is given rather than in one
// written for them.
func testLevel(t *testing.T) carpen.Level {
	t.Helper()

	levels, err := carpen.Levels()
	if err != nil {
		t.Fatalf("loading the levels: %v", err)
	}
	return levels[0]
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

// stubResizer is a stubScene that also listens for the screen's size, so that
// what the manager tells a scene can be tested without a real one.
type stubResizer struct {
	stubScene
	view viewport
}

func (s *stubResizer) resize(v viewport) { s.view = v }

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
	m := NewManager(first)

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
	m := NewManager(&stubScene{next: second})

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
	m := NewManager(first)

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
	m := NewManager(first)

	m.Draw(nil)

	if first.draws != 1 {
		t.Errorf("running scene drawn %d times, want 1", first.draws)
	}
}

// The layout holds the height and follows the device's width. The height is
// what everything in the game is sized against, so it is the half that must not
// move; the width is what a phone has more of than the game was written for,
// and following it is what keeps the cars from being stretched.
func TestManagerLayoutHoldsHeightAndFollowsWidth(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		outsideWidth, outsideHeight int
		wantWidth                   int
	}{
		{"the window the game opens in", 640, 480, 640},
		{"a desktop monitor", 1920, 1080, 853},
		{"a phone in landscape", 852, 393, 1041},
		{"a phone in portrait, which is never narrower than the lot", 393, 852, 640},
		{"a tablet in landscape, which is nearly the lot's own shape", 1180, 820, 691},
		{"a big tablet, which is exactly it", 1376, 1032, 640},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := NewManager(&stubScene{})

			width, height := m.Layout(tc.outsideWidth, tc.outsideHeight)

			if width != tc.wantWidth || height != 480 {
				t.Errorf("Layout(%d, %d) = %dx%d, want %dx480",
					tc.outsideWidth, tc.outsideHeight, width, height, tc.wantWidth)
			}
		})
	}
}

// Layout is the one place the device's own measure can be read: outsideHeight
// arrives in points on a phone or a tablet, and against the height the game is
// pinned to that is what one game pixel comes to in the player's hand. Anything
// a thumb has to land on is sized through it, so a tablet's controls are not
// twice the size of a phone's.
func TestManagerLayoutRecordsWhatAGamePixelIsOnTheDevice(t *testing.T) {
	for _, tc := range []struct {
		name          string
		outsideHeight int
		want          float64
	}{
		{"a phone in landscape", 393, 393.0 / 480},
		{"a tablet in landscape", 820, 820.0 / 480},
		{"the window the game opens in", 480, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := NewManager(&stubScene{})

			m.Layout(tc.outsideHeight*2, tc.outsideHeight)

			if math.Abs(m.pointsPerPixel-tc.want) > 0.0001 {
				t.Errorf("one game pixel is %.4f points, want %.4f", m.pointsPerPixel, tc.want)
			}
		})
	}
}

// A window reporting nothing is a window the game cannot lay itself out in, so
// the last size it did have stands rather than a screen no pixels wide.
func TestManagerLayoutKeepsLastSizeWhenTheWindowIsEmpty(t *testing.T) {
	m := NewManager(&stubScene{})
	m.Layout(852, 393)

	width, height := m.Layout(0, 0)

	if width != 1041 || height != 480 {
		t.Errorf("Layout(0, 0) = %dx%d, want the 1041x480 it last settled on", width, height)
	}
}

// Scenes are told the size before they are updated, and again the moment they
// take over — a scene is drawn on the frame it takes over on, and hit-testing a
// tap against a layout from another screen size is how a button ends up
// somewhere other than where it was drawn.
func TestManagerTellsScenesTheScreenSize(t *testing.T) {
	taking := &stubResizer{}
	m := NewManager(&stubScene{next: taking})
	m.Layout(852, 393)

	if err := m.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if taking.view.width != 1041 || taking.view.height != 480 {
		t.Errorf("the scene taking over was told %dx%d, want 1041x480", taking.view.width, taking.view.height)
	}
	if want := 393.0 / 480; math.Abs(taking.view.pointsPerPixel-want) > 0.0001 {
		t.Errorf("the scene taking over was told one game pixel is %.4f points, want %.4f", taking.view.pointsPerPixel, want)
	}
}

// The cycle the game is built around: the menu starts a race, a finished race
// shows its results, and the results lead back to the menu.
func TestMenuGameplayResultsCycle(t *testing.T) {
	in := newFakeInput()
	m := NewManager(NewMenu(in, testLevel(t)))

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
		m := NewManager(newResults(in, testLevel(t)))

		tick(t, m, in, ebiten.KeyEscape)

		if _, ok := m.current.(*Menu); !ok {
			t.Errorf("game is on %T, want *Menu", m.current)
		}
	})

	t.Run("pause back to the race", func(t *testing.T) {
		in := newFakeInput()
		game := newGameplay(in, testLevel(t))
		m := NewManager(game)

		tick(t, m, in, ebiten.KeyEscape)
		tick(t, m, in, ebiten.KeyEscape)

		if m.current != Scene(game) {
			t.Errorf("game is on %#v, want the race it paused", m.current)
		}
	})

	t.Run("menu ends the game", func(t *testing.T) {
		in := newFakeInput()
		m := NewManager(NewMenu(in, testLevel(t)))

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
	m := NewManager(NewMenu(in, testLevel(t)))

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
	game := newGameplay(in, testLevel(t))
	m := NewManager(game)

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
	game := newGameplay(in, testLevel(t))
	m := NewManager(game)

	tick(t, m, in, ebiten.KeyEscape)
	before := game.cars[0].Pivot

	for i := 0; i < 10; i++ {
		tick(t, m, in)
	}

	if game.cars[0].Pivot != before {
		t.Errorf("car moved from %v to %v while paused, want it held still", before, game.cars[0].Pivot)
	}
}

// Pausing lets go of the wheel. Only the car being driven is read each tick, so
// a car left holding the accelerator holds it for as long as nobody is driving
// it, and the race would come back with the car already flat out.
func TestPauseReleasesHeldControls(t *testing.T) {
	in := newFakeInput()
	game := newGameplay(in, testLevel(t))
	m := NewManager(game)

	tick(t, m, in, ebiten.KeyUp, ebiten.KeyLeft)
	if game.cars[0].Throttle == 0 || game.cars[0].Steering == 0 {
		t.Fatal("holding Up and Left did not reach the car")
	}

	tick(t, m, in, ebiten.KeyEscape)

	if car := game.cars[0]; car.Throttle != 0 || car.Brake != 0 || car.Steering != 0 {
		t.Errorf("car still driving itself after the pause: %+v", car)
	}
}

// Swapping lets go of the wheel for the same reason pausing does, and used not
// to: the car handed over kept whatever it was last told, and a swap made with
// the accelerator down left it flat out for the rest of the race with nothing
// able to lift it.
func TestSwappingCarsReleasesTheCarHandedOver(t *testing.T) {
	in := newFakeInput()
	game := newGameplay(in, testLevel(t))

	in.press(ebiten.KeyUp)
	if _, err := game.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if game.cars[0].Throttle == 0 {
		t.Fatal("holding Up did not reach the car")
	}

	// Tab swaps on the way up, and the accelerator is still down as it does.
	in.clear()
	in.released[ebiten.KeyTab] = true
	in.pressed[ebiten.KeyUp] = true
	if _, err := game.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if game.activeCar != 1 {
		t.Fatalf("active car = %d, want the swap to have happened", game.activeCar)
	}
	if car := game.cars[0]; car.Throttle != 0 || car.Brake != 0 || car.Steering != 0 {
		t.Errorf("the car handed over is still driving itself: %+v", car)
	}
}

func TestPauseQuitReturnsToTheMenu(t *testing.T) {
	in := newFakeInput()
	m := NewManager(newGameplay(in, testLevel(t)))

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
	first := newGameplay(in, testLevel(t))
	m := NewManager(first)

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
	if second.cars[0].Pivot != newGameplay(in, testLevel(t)).cars[0].Pivot {
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
		{name: "up", key: ebiten.KeyUp, held: func(c carpen.Car) bool { return c.Throttle == 1 }},
		{name: "down", key: ebiten.KeyDown, held: func(c carpen.Car) bool { return c.Brake == 1 }},
		{name: "left", key: ebiten.KeyLeft, held: func(c carpen.Car) bool { return c.Steering == -1 }},
		{name: "right", key: ebiten.KeyRight, held: func(c carpen.Car) bool { return c.Steering == 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := newFakeInput()
			game := newGameplay(in, testLevel(t))

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
	game := newGameplay(in, testLevel(t))

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
	game := newGameplay(in, testLevel(t))
	before := game.cars[0].Pivot

	if _, err := game.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if game.cars[0].Pivot == before {
		t.Errorf("car did not move on a tick: still at %v", before)
	}
}
