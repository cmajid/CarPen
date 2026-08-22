package scene

import (
	"testing"

	"github.com/cmajid/carpen/carpen"
	"github.com/hajimehoshi/ebiten/v2"
)

// aLevel is a level written for these tests: two obstacles of each kind, in
// positions no other level uses, so anything the scene reads from it shows.
func aLevel() carpen.Level {
	return carpen.Level{
		ID:       "level-test",
		Name:     "A Test Level",
		Lot:      carpen.Lot{Width: 640, Height: 480},
		Car:      carpen.CarStart{Color: "red", X: 120, Y: 240, Rotation: 30},
		Bay:      carpen.Bay{X: 400, Y: 120, Width: 130, Height: 210},
		Attempts: 3,
		Obstacles: []carpen.Obstacle{
			{Type: carpen.ObstacleBush, X: 20, Y: 30},
			{Type: carpen.ObstacleCar, Color: "yellow", X: 300, Y: 400, Rotation: 90},
			{Type: carpen.ObstacleBush, X: 60, Y: 70},
		},
	}
}

// The world is built out of the level and nothing else: every car and bush the
// level asks for is there, in the colour, place and heading it asked for.
func TestGameplayBuildsTheWorldFromTheLevel(t *testing.T) {
	level := aLevel()

	game := newGameplay(newFakeInput(), level)

	if len(game.bushes) != 2 {
		t.Errorf("built %d bushes, want the 2 the level asks for", len(game.bushes))
	}
	for i, want := range []carpen.Direction{{X: 20, Y: 30}, {X: 60, Y: 70}} {
		if game.bushes[i].Direction != want {
			t.Errorf("bush %d is at %v, want %v", i, game.bushes[i].Direction, want)
		}
	}

	if len(game.cars) != 2 {
		t.Fatalf("built %d cars, want the player's and the level's one parked car", len(game.cars))
	}
	for i, want := range []struct {
		colour   string
		x, y     float64
		rotation float64
	}{
		{colour: "red", x: 120, y: 240, rotation: 30},
		{colour: "yellow", x: 300, y: 400, rotation: 90},
	} {
		car := game.cars[i]
		if car.Color != want.colour || car.X != want.x || car.Y != want.y || car.Rotation != want.rotation {
			t.Errorf("car %d is a %s at %g,%g facing %g, want a %s at %g,%g facing %g",
				i, car.Color, car.X, car.Y, car.Rotation, want.colour, want.x, want.y, want.rotation)
		}
	}
}

// The player drives the level's own car, which is the first one built, whatever
// else the level parks around it.
func TestGameplayStartsOnTheLevelsCar(t *testing.T) {
	game := newGameplay(newFakeInput(), aLevel())

	if game.activeCar != 0 {
		t.Fatalf("the player starts on car %d, want the level's own car", game.activeCar)
	}
	if !game.cars[0].IsActive {
		t.Error("the car the player is driving is not marked active")
	}
	if game.cars[1].IsActive {
		t.Error("a parked car is marked active")
	}
}

// Tab hands the keys on around the cars of the level being played, however many
// of them there are, rather than between a fixed two.
func TestGameplayTabWrapsAroundTheLevelsCars(t *testing.T) {
	in := newFakeInput()
	level := aLevel()
	level.Obstacles = append(level.Obstacles, carpen.Obstacle{Type: carpen.ObstacleCar, Color: "red", X: 500, Y: 200})
	game := newGameplay(in, level)

	for _, want := range []int{1, 2, 0} {
		in.release(ebiten.KeyTab)
		if _, err := game.Update(); err != nil {
			t.Fatalf("Update: %v", err)
		}

		if game.activeCar != want {
			t.Errorf("active car = %d, want %d", game.activeCar, want)
		}
	}
}

// A level with a single car is the shape the parking puzzle is heading for, and
// it must not leave Tab pointing at a car that is not there.
func TestGameplayTabIsHarmlessWithOneCar(t *testing.T) {
	in := newFakeInput()
	level := aLevel()
	level.Obstacles = nil
	game := newGameplay(in, level)

	in.release(ebiten.KeyTab)
	if _, err := game.Update(); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if game.activeCar != 0 {
		t.Errorf("active car = %d, want the only car there is", game.activeCar)
	}
}

// The race, the pause and the results all lead back to the level being played,
// so a player who restarts, or races again, gets the same puzzle.
func TestTheLevelSurvivesEveryWayBackToTheGame(t *testing.T) {
	level := aLevel()

	t.Run("restarting from the pause", func(t *testing.T) {
		in := newFakeInput()
		m := NewManager(newGameplay(in, level))

		tick(t, m, in, ebiten.KeyEscape)
		choose(t, m, in, pauseRestart)

		game, ok := m.current.(*Gameplay)
		if !ok {
			t.Fatalf("the game is on %T, want *Gameplay", m.current)
		}
		if game.level.ID != level.ID {
			t.Errorf("restarted on level %q, want %q", game.level.ID, level.ID)
		}
	})

	t.Run("racing again from the results", func(t *testing.T) {
		in := newFakeInput()
		m := NewManager(newGameplay(in, level))

		tick(t, m, in, ebiten.KeyEnter)
		choose(t, m, in, resultsAgain)

		game, ok := m.current.(*Gameplay)
		if !ok {
			t.Fatalf("the game is on %T, want *Gameplay", m.current)
		}
		if game.level.ID != level.ID {
			t.Errorf("raced again on level %q, want %q", game.level.ID, level.ID)
		}
	})

	t.Run("out to the menu and back in", func(t *testing.T) {
		in := newFakeInput()
		m := NewManager(newGameplay(in, level))

		tick(t, m, in, ebiten.KeyEscape)
		choose(t, m, in, pauseQuit)
		choose(t, m, in, menuStart)

		game, ok := m.current.(*Gameplay)
		if !ok {
			t.Fatalf("the game is on %T, want *Gameplay", m.current)
		}
		if game.level.ID != level.ID {
			t.Errorf("started level %q from the menu, want %q", game.level.ID, level.ID)
		}
	})
}

// hitsAWall reports whether a box is inside any of the boxes that close the lot
// in — which is what a car driving off the edge finds out the hard way.
func hitsAWall(walls []*carpen.OBB, box *carpen.OBB) bool {
	for _, wall := range walls {
		if carpen.Intersects(box, wall) {
			return true
		}
	}
	return false
}

// The tarmac reaches the screen's edges. A phone is far wider than the 4:3 the
// level is drawn at, and the ground that fills the difference is ground the car
// can actually use — otherwise it is a picture of a car park with a wall drawn
// down the middle of it.
func TestPlayAreaReachesTheScreenEdges(t *testing.T) {
	g := newGameplay(newFakeInput(), aLevel())
	g.resize(viewport{width: 1040, height: 480}) // a handset in landscape

	area := g.playArea()

	if area.width != 1040 || area.height != 480 {
		t.Errorf("play area is %.0fx%.0f, want the screen's own 1040x480", area.width, area.height)
	}
	if area.left != -200 || area.top != 0 {
		t.Errorf("play area starts at %.0f, %.0f, want -200, 0 — the level centred in it", area.left, area.top)
	}
}

// The level is not moved by any of this; it is placed. Where the screen has more
// room than the level asked for, the level goes in the middle of it and stays
// exactly the puzzle it was — so what a wider phone gives is room at the edges
// and never a different arrangement of the things being parked around.
func TestABiggerScreenPlacesTheLevelWithoutMovingIt(t *testing.T) {
	level := aLevel()
	g := newGameplay(newFakeInput(), level)

	before := g.cars[0].Pivot
	g.resize(viewport{width: 1040, height: 480})

	if g.cars[0].Pivot != before {
		t.Errorf("the car moved to %+v on a wider screen, from %+v", g.cars[0].Pivot, before)
	}
	if origin := g.origin(); origin.X != 200 || origin.Y != 0 {
		t.Errorf("the level is drawn from %+v, want 200, 0 — centred in the screen's spare room", origin)
	}
}

// What the extra ground is for: it can be driven on. The wall is out at the
// screen's edge, not at the level's, so the thumbs rest on tarmac the car is
// allowed to reach rather than on a painted-on boundary.
func TestTheCarMayDriveOnTheExtraGround(t *testing.T) {
	g := newGameplay(newFakeInput(), aLevel())
	g.resize(viewport{width: 1040, height: 480})

	// Out past where the level's own rectangle ends, but on screen.
	onTheMargin := carpen.NewOBB(-100, 240, 40, 40, 0)
	if hitsAWall(g.walls, onTheMargin) {
		t.Error("the ground beside the level is walled off, so the screen shows tarmac the car cannot use")
	}

	// Past the screen's edge, which is where the wall belongs.
	offTheScreen := carpen.NewOBB(-260, 240, 40, 40, 0)
	if !hitsAWall(g.walls, offTheScreen) {
		t.Error("driving off the side of the screen is not a collision")
	}
}

// On a screen the shape of the level there is nothing to extend, and the lot is
// the level's own — the game a desktop window has been playing all along.
func TestPlayAreaIsTheLevelWhenTheScreenFitsIt(t *testing.T) {
	g := newGameplay(newFakeInput(), aLevel())
	g.resize(viewport{width: 640, height: 480})

	area := g.playArea()

	if area != (lot{left: 0, top: 0, width: 640, height: 480}) {
		t.Errorf("play area is %+v, want the level's own 640x480 at the origin", area)
	}
}
