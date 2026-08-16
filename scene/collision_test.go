package scene

import (
	"testing"

	"github.com/cmajid/carpen/carpen"
	"github.com/hajimehoshi/ebiten/v2"
)

// collisionLevel is a lot with the player's car wherever a test needs it and
// nothing else in the way unless the test puts it there. It skips the loader
// on purpose: these tests place things by coordinate, not by level file.
func collisionLevel(car carpen.CarStart, obstacles ...carpen.Obstacle) carpen.Level {
	return carpen.Level{
		ID:        "collision-test",
		Name:      "Collision Test",
		Lot:       carpen.Lot{Width: 640, Height: 480},
		Car:       car,
		Bay:       carpen.Bay{X: 320, Y: 400, Width: 140, Height: 230},
		Attempts:  3,
		Obstacles: obstacles,
	}
}

// recordCollisions swaps the gameplay's rules for a notebook, so a test can
// see exactly which events detection raised and when.
func recordCollisions(g *Gameplay) *[]carpen.CollisionEvent {
	events := &[]carpen.CollisionEvent{}
	g.OnCollision = func(event carpen.CollisionEvent) { *events = append(*events, event) }
	return events
}

// TestGameplayDetectsWallSameTick starts the car a few pixels short of the top
// edge, coasting towards it. The tick whose step carries the car over the edge
// must be the tick the event fires — not the one after.
func TestGameplayDetectsWallSameTick(t *testing.T) {
	in := newFakeInput()
	g := newGameplay(in, collisionLevel(carpen.CarStart{Color: "yellow", X: 100, Y: 13}))
	events := recordCollisions(g)

	// The car starts with its body's top edge 3 pixels into the lot and rolls
	// up 1 pixel on the first tick: close, but still no collision.
	if _, err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if len(*events) != 0 {
		t.Fatalf("collision reported while the car was still inside the lot: %v", *events)
	}

	// The second tick's step carries the top edge past the boundary.
	if _, err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if len(*events) != 1 {
		t.Fatalf("got %d collision events, want 1", len(*events))
	}
	if (*events)[0].Obstruction != carpen.ObstructionWall {
		t.Errorf("crashed into a %q, want a %q", (*events)[0].Obstruction, carpen.ObstructionWall)
	}

	// Still overlapping a tick later; a crash that has already been reported
	// is not news — but the live state the F3 overlay reads still says so.
	if _, err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if len(*events) != 1 {
		t.Errorf("got %d collision events after sitting in the wall, want still 1", len(*events))
	}
	if g.touching != carpen.ObstructionWall {
		t.Errorf("live touching state = %q, want %q while sitting in the wall", g.touching, carpen.ObstructionWall)
	}
}

// TestGameplayDetectsBush puts a bush in the path the car coasts along from a
// standing start and lets it roll into it.
func TestGameplayDetectsBush(t *testing.T) {
	in := newFakeInput()
	g := newGameplay(in,
		collisionLevel(carpen.CarStart{Color: "yellow", X: 200, Y: 260},
			carpen.Obstacle{Type: carpen.ObstacleBush, X: 200, Y: 100}))
	events := recordCollisions(g)

	for i := 0; i < 30 && len(*events) == 0; i++ {
		if _, err := g.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if len(*events) != 1 || (*events)[0].Obstruction != carpen.ObstructionBush {
		t.Fatalf("got %v, want one crash into a %q", *events, carpen.ObstructionBush)
	}
}

// TestGameplayDetectsOtherCar starts a parked car ahead of the active one.
// Both coast from the level start — every car begins with the prototype's
// rolling speed — so the active car has to be driven flat out to run the
// decelerating one down.
func TestGameplayDetectsOtherCar(t *testing.T) {
	in := newFakeInput()
	g := newGameplay(in,
		collisionLevel(carpen.CarStart{Color: "yellow", X: 100, Y: 260},
			carpen.Obstacle{Type: carpen.ObstacleCar, Color: "yellow", X: 100, Y: 30}))
	events := recordCollisions(g)

	in.press(ebiten.KeyUp)
	for i := 0; i < 60 && len(*events) == 0; i++ {
		if _, err := g.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if len(*events) != 1 || (*events)[0].Obstruction != carpen.ObstructionCar {
		t.Fatalf("got %v, want one crash into a %q", *events, carpen.ObstructionCar)
	}
}

func TestGameplayClearRunRaisesNothing(t *testing.T) {
	in := newFakeInput()
	g := newGameplay(in, collisionLevel(carpen.CarStart{Color: "yellow", X: 250, Y: 150}))
	events := recordCollisions(g)

	for i := 0; i < 30; i++ {
		if _, err := g.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if len(*events) != 0 {
		t.Errorf("car crossing an empty lot crashed: %v", *events)
	}
	if g.touching != "" {
		t.Errorf("live touching state = %q on an empty lot, want clear", g.touching)
	}
}

// TestGameplaySpawnOverlapIsNotACrash starts the car already overlapping a
// bush, the way level-01 hangs the car's rear over the lot edge. A crash is
// running into something; what the level put the car on top of is not one —
// but backing into the same bush after driving clear of it is.
func TestGameplaySpawnOverlapIsNotACrash(t *testing.T) {
	in := newFakeInput()
	g := newGameplay(in,
		collisionLevel(carpen.CarStart{Color: "yellow", X: 200, Y: 150},
			carpen.Obstacle{Type: carpen.ObstacleBush, X: 250, Y: 320}))
	events := recordCollisions(g)

	// The car coasts up off the bush and comes to rest well clear of it.
	for i := 0; i < 40; i++ {
		if _, err := g.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if len(*events) != 0 {
		t.Fatalf("the overlap the car spawned in was reported as a crash: %v", *events)
	}

	// Reversing back down into the bush is a crash of the player's making.
	in.press(ebiten.KeyDown)
	for i := 0; i < 200 && len(*events) == 0; i++ {
		if _, err := g.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if len(*events) != 1 || (*events)[0].Obstruction != carpen.ObstructionBush {
		t.Fatalf("got %v, want one crash into a %q", *events, carpen.ObstructionBush)
	}
}

func TestGameplayDebugOverlayToggle(t *testing.T) {
	in := newFakeInput()
	g := newGameplay(in, collisionLevel(carpen.CarStart{Color: "yellow", X: 250, Y: 150}))

	in.press(ebiten.KeyF3)
	if _, err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if !g.debugOBB {
		t.Fatal("F3 did not switch the OBB overlay on")
	}
	in.clear()

	if _, err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if !g.debugOBB {
		t.Fatal("overlay switched itself off with no key pressed")
	}

	in.press(ebiten.KeyF3)
	if _, err := g.Update(); err != nil {
		t.Fatal(err)
	}
	if g.debugOBB {
		t.Fatal("second F3 did not switch the OBB overlay off")
	}
}
