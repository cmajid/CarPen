package scene

import (
	"fmt"
	"image/color"
	"math"

	"github.com/cmajid/carpen/carpen"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Gameplay is the game itself: the cars, the bushes, and the driving.
type Gameplay struct {
	in        Input
	level     carpen.Level
	cars      []carpen.Car
	bushes    []carpen.Bush
	walls     []*carpen.OBB
	activeCar int
	fade      fade

	// touch is the controls drawn on the screen, for a player with no keyboard
	// and no pad. Every read of the controls below goes through it rather than
	// through the package's own analog and justPressed, which is what folds a
	// finger in alongside the other two devices.
	touch touchControls

	// view is the screen the manager last reported, which the HUD and the
	// on-screen controls are laid out against — its size, and how big one game
	// pixel is on the device holding it.
	view viewport

	// OnCollision is the rules layer's ear. Detection only raises the event;
	// whatever is plugged in here decides what a crash means, so no game-over
	// lives anywhere near the collision code.
	OnCollision func(carpen.CollisionEvent)

	// crash is the last collision the default rules recorded, shown on the
	// HUD. It stands in for the attempt-failure rule the epic plans (#17),
	// the same way Enter stands in for the win condition.
	crash *carpen.CollisionEvent

	// colliding remembers whether the active car was already touching
	// something last tick, so OnCollision fires once when a crash starts
	// rather than every tick the car sits overlapping. It is seeded with the
	// spawn state, because a crash is running into something: a level that
	// places the car already overlapping (level-01 hangs its rear over the
	// bottom edge) is not a crash the player made.
	colliding bool

	// touching is what the active car is overlapping right now — empty when
	// it is clear. Where crash remembers the first collision the way the
	// rules will, this is the live truth of this tick, and it is what the F3
	// overlay paints red and names in its status line.
	touching carpen.Obstruction

	// debugOBB, toggled with F3, outlines every box collision sees.
	debugOBB bool
}

// newGameplay lays out a fresh world from a level. Nothing about where things
// stand is decided here: the level says what goes where, and this only turns
// each of those into the car or the bush that drives and draws.
//
// The player's car is always cars[0], so the level's own car is the one the
// player is given the keys to.
func newGameplay(in Input, level carpen.Level) *Gameplay {
	g := &Gameplay{in: in, level: level}

	g.cars = []carpen.Car{newCar(level.Car.Color, level.Car.X, level.Car.Y, level.Car.Rotation, true)}
	for _, obstacle := range level.Obstacles {
		switch obstacle.Type {
		case carpen.ObstacleBush:
			g.bushes = append(g.bushes, newBush(obstacle.X, obstacle.Y))
		case carpen.ObstacleCar:
			g.cars = append(g.cars, newCar(obstacle.Color, obstacle.X, obstacle.Y, obstacle.Rotation, false))
		}
	}

	for i := range g.bushes {
		g.bushes[i].Init()
	}

	g.OnCollision = func(event carpen.CollisionEvent) { g.crash = &event }

	for i := range g.cars {
		g.cars[i].Init()

		g.cars[i].Pivot = carpen.Pivot{X: g.cars[i].X + 50, Y: g.cars[i].Y + 20}
		g.cars[i].DirectionPivot = carpen.DirectionPivot{X: g.cars[i].FrontPivot.X, Y: g.cars[i].FrontPivot.Y - 50}
		g.cars[i].UpdateRearPivotAbs()

		v1 := carpen.Vector{X: g.cars[i].DirectionPivot.X - g.cars[i].FrontPivot.X, Y: g.cars[i].DirectionPivot.Y - g.cars[i].FrontPivot.Y}
		g.cars[i].Direction = v1.Normalize()
	}

	// The level's own size stands in until the manager says how big the screen
	// really is, so a race drawn before its first resize is drawn somewhere
	// sensible rather than at nothing by nothing. This also builds the walls,
	// which is why it runs before the spawn state is read below.
	g.resize(viewport{width: int(level.Lot.Width), height: int(level.Lot.Height)})

	event, hit := g.findCollision()
	g.colliding, g.touching = hit, event.Obstruction

	return g
}

// resize lays the race out on a screen of this size: the furniture around it,
// and the tarmac itself, which reaches the screen's edges however far away they
// are. The walls go with the tarmac, so they are rebuilt here rather than once
// at the start — a screen that changes shape changes where the lot ends.
func (g *Gameplay) resize(v viewport) {
	g.view = v
	g.walls = lotWalls(g.playArea())
	g.touch.resize(v)
}

func newCar(paint string, x float64, y float64, rotate float64, active bool) carpen.Car {
	car := carpen.Car{
		Color:    paint,
		IsActive: active,

		MaxSpeed:          6,
		WheelWidth:        12,
		WheelHeight:       30,
		WheelRotationStep: 2.4, // degrees of steering per tick (Ebiten runs 60 ticks/second)
		WheelMaxAngle:     45,
		WheelAngle:        0,
		X:                 x,
		Y:                 y,
		FrontPivot:        carpen.FrontPivot{X: 0, Y: 0},
		RearPivot:         carpen.RearPivot{X: 0, Y: 160},
		Rotation:          rotate,
		Wheels: []carpen.Wheel{
			{X: -40, Y: 10},
			{X: 45, Y: 10},
			{X: -41, Y: 145},
			{X: 46, Y: 145},
		},
		Speed:        5,
		Acceleration: 0.2,
	}

	return car
}

func newBush(x, y float64) carpen.Bush {
	bush := carpen.Bush{
		Direction: carpen.Direction{
			X: x,
			Y: y,
		},
	}
	return bush
}

// lotWalls closes the lot in: four boxes pressed against the outside of its
// edges, so driving off any side is a collision like any other. They are far
// thicker than the longest step a car can take in a tick (MaxSpeed is 6), so
// a car can never jump one between two checks.
//
// The lot they close is the play area rather than the level's own rectangle —
// on a screen wider than the level, the tarmac reaches the edges and so do
// these. They are rebuilt whenever the screen changes size, which is the one
// thing about this game that a device is allowed to decide.
func lotWalls(area lot) []*carpen.OBB {
	const thickness = 100.0
	l, t, w, h := area.left, area.top, area.width, area.height

	return []*carpen.OBB{
		carpen.NewOBB(l+w/2, t-thickness/2, w+2*thickness, thickness, 0),   // top
		carpen.NewOBB(l+w/2, t+h+thickness/2, w+2*thickness, thickness, 0), // bottom
		carpen.NewOBB(l-thickness/2, t+h/2, thickness, h+2*thickness, 0),   // left
		carpen.NewOBB(l+w+thickness/2, t+h/2, thickness, h+2*thickness, 0), // right
	}
}

// lot is the ground the car may drive on, in the level's own coordinates. The
// level says how big its own rectangle is; where the screen is bigger than
// that, the tarmac is extended evenly on every side to meet the edges, so its
// left and top are negative — the level's (0, 0) is no longer the corner of the
// world, it is just the corner of the part of the world the level furnished.
//
// This is what keeps a wider phone from being a different puzzle: the bay, the
// bushes and the parked cars are all placed by the level and none of them move.
// What a bigger screen buys is empty tarmac out at the sides, which is also
// where the player's thumbs are.
type lot struct {
	left, top, width, height float64
}

// playArea is that ground for the screen the game is currently on.
func (g *Gameplay) playArea() lot {
	origin := g.origin()

	return lot{
		left:   -origin.X,
		top:    -origin.Y,
		width:  math.Max(g.level.Lot.Width, float64(g.view.width)),
		height: math.Max(g.level.Lot.Height, float64(g.view.height)),
	}
}

// origin is where the level's own (0, 0) falls on the screen: half of whatever
// the screen has over and above the level, so the level sits in the middle of
// it. Everything drawn places itself against this, and it is rounded to a whole
// pixel so that sprites do not land on half of one.
func (g *Gameplay) origin() carpen.Vector {
	return carpen.Vector{
		X: math.Round(math.Max(0, float64(g.view.width)-g.level.Lot.Width) / 2),
		Y: math.Round(math.Max(0, float64(g.view.height)-g.level.Lot.Height) / 2),
	}
}

func (g *Gameplay) Update() (Scene, error) {
	g.fade.update()

	// The screen is read first, because everything below asks what the player
	// is doing and the answer for a finger is worked out here.
	g.touch.update(g.in)

	if g.touch.justPressed(g.in, actionCancel) {
		return newPause(g), nil
	}
	// Standing in for the win condition, which cannot tell the race is over
	// until the cars can hit something (#20): Enter ends the race by hand.
	if g.touch.justPressed(g.in, actionFinish) {
		return newResults(g.in, g.level), nil
	}

	// The controls are read as they stand this tick rather than on their edges.
	// A trigger's pull and a stick's lean are positions and have no press to
	// catch, and a position read afresh every tick cannot be left behind by an
	// edge that happened on another screen.
	//
	// This only records what is being asked for; every change to Speed is made
	// by Car.Move(), which is the one place that honours MaxSpeed and the
	// reverse limit.
	active := &g.cars[g.activeCar]
	active.Throttle = g.touch.analog(g.in, actionThrottle)
	active.Brake = g.touch.analog(g.in, actionBrake)
	active.Steering = g.touch.analog(g.in, actionSteerRight) - g.touch.analog(g.in, actionSteerLeft)

	// Swapping hands the keys to the next car in the level, which is how the
	// prototype's two cars are still both drivable. It comes to nothing on a
	// level with a single car, and goes for good once a parked car is something
	// to be avoided rather than driven (#20).
	//
	// The car being handed over has to be let go of first. Only the active car
	// is read, so whatever the one before was last told stands for good: a swap
	// made with the accelerator down used to leave that car flat out with
	// nothing left to lift.
	if g.touch.justReleased(g.in, actionSwapCar) {
		g.releaseControls()
		g.activeCar = (g.activeCar + 1) % len(g.cars)
	}

	if justPressed(g.in, actionDebugBoxes) {
		g.debugOBB = !g.debugOBB
	}

	for i := range g.cars {
		g.cars[i].Update()
	}

	// Collision is looked for after every car has taken its step, so a crash
	// is raised in the very tick the boxes first meet. The event only fires on
	// the tick a crash starts; the rules decide what it means from there.
	event, hit := g.findCollision()
	if hit && !g.colliding && g.OnCollision != nil {
		g.OnCollision(event)
	}
	g.colliding = hit
	g.touching = event.Obstruction

	return nil, nil
}

// findCollision reports the first thing the active car is overlapping: a wall,
// a bush, or another car. Only the active car is checked — everything else
// stands still, and things that stand still do not run into each other. At
// this handful of boxes, checking each in turn is the whole broad phase.
func (g *Gameplay) findCollision() (carpen.CollisionEvent, bool) {
	active := g.cars[g.activeCar].OBB()

	for _, wall := range g.walls {
		if carpen.Intersects(active, wall) {
			return carpen.CollisionEvent{Obstruction: carpen.ObstructionWall}, true
		}
	}
	for i := range g.bushes {
		if g.bushes[i].Collider().IntersectsOBB(active) {
			return carpen.CollisionEvent{Obstruction: carpen.ObstructionBush}, true
		}
	}
	for i := range g.cars {
		if i == g.activeCar {
			continue
		}
		if carpen.Intersects(active, g.cars[i].OBB()) {
			return carpen.CollisionEvent{Obstruction: carpen.ObstructionCar}, true
		}
	}

	return carpen.CollisionEvent{}, false
}

func (g *Gameplay) Draw(screen *ebiten.Image) {
	// The lot is the ground the level is played on, and it is the whole screen:
	// the level furnishes the middle of it and the rest is tarmac out to the
	// edges (see playArea). Everything below places itself in the level's own
	// coordinates and is put where the level was put by origin, so nothing in
	// the game has to know what shape the device is.
	screen.Fill(color.White)

	origin := g.origin()
	g.drawBay(screen, origin)

	for i := range g.cars {
		g.cars[i].DrawCar(screen, origin)
	}
	for i := range g.bushes {
		g.bushes[i].Draw(screen, origin)
	}

	if g.debugOBB {
		g.drawOBBs(screen, origin)
	}

	// The HUD and the controls belong to the screen rather than to the lot, so
	// they are drawn after it and in its own coordinates.
	g.drawHUD(screen)
	if g.debugOBB {
		g.drawDebugStatus(screen)
	}
	if g.in.TouchActive() {
		g.touch.draw(screen)
	}

	g.fade.draw(screen)
}

// drawOBBs outlines every box collision sees — the development overlay behind
// F3. The active car's box turns red for exactly the ticks it overlaps
// something, so a clipped corner is seen the moment it happens rather than
// found later in the crash notice. Reading an entity's OBB() re-places its box
// but steps no physics, so drawing it here breaks nothing about the fixed
// tick rate.
func (g *Gameplay) drawOBBs(screen *ebiten.Image, origin carpen.Vector) {
	for i := range g.cars {
		colour := colourAccent
		if i == g.activeCar && g.touching != "" {
			colour = colourDanger
		}
		strokeOBB(screen, g.cars[i].OBB(), origin, colour)
	}
	for i := range g.bushes {
		circle := g.bushes[i].Collider()
		strokeCircle(screen, shift(circle.Center(), origin), circle.Radius(), 2, colourAccent)
	}
	for _, wall := range g.walls {
		strokeOBB(screen, wall, origin, colourBay)
	}
}

// shift moves a point from the level's coordinates onto the screen.
func shift(point, origin carpen.Vector) carpen.Vector {
	return carpen.Vector{X: point.X + origin.X, Y: point.Y + origin.Y}
}

// drawDebugStatus writes the active car's live numbers, and what its box is
// overlapping this very tick, on a strip along the bottom. This is the truth
// of now, where the crash notice is the memory of the first hit — the pair is
// what tells "I am touching it" apart from "I touched it once, back then".
func (g *Gameplay) drawDebugStatus(screen *ebiten.Image) {
	car := &g.cars[g.activeCar]

	status, colour := "clear", colourText
	if g.touching != "" {
		status, colour = "touching "+string(g.touching), colourDanger
	}

	width := float64(screen.Bounds().Dx())
	bottom := float64(screen.Bounds().Dy())
	fillRect(screen, 0, bottom-22, width, 22, colourHUD)
	drawText(screen,
		fmt.Sprintf("pivot %.0f, %.0f    rotation %.0f    speed %.1f    %s",
			car.Pivot.X, car.Pivot.Y, car.Rotation, car.Speed, status),
		fontPrompt, 14, bottom-11, colour, text.AlignStart, text.AlignCenter)
}

func strokeOBB(dst *ebiten.Image, obb *carpen.OBB, origin carpen.Vector, colour color.Color) {
	outline := obb.Outline()
	for i := range outline {
		strokeLine(dst, shift(outline[i], origin), shift(outline[(i+1)%len(outline)], origin), 2, colour)
	}
}

// drawBay marks out the space the level asks the player to park in: the two
// sides and the back of the bay, painted on the ground and left open on the
// edge the car drives in over. It is drawn before the cars, so a car parked in
// the bay stands on the lines rather than under them.
func (g *Gameplay) drawBay(screen *ebiten.Image, origin carpen.Vector) {
	nearLeft, nearRight, farLeft, farRight := g.level.Bay.Corners()

	strokeLine(screen, shift(nearLeft, origin), shift(farLeft, origin), bayLineWidth, colourBay)
	strokeLine(screen, shift(nearRight, origin), shift(farRight, origin), bayLineWidth, colourBay)
	strokeLine(screen, shift(farLeft, origin), shift(farRight, origin), bayLineWidth, colourBay)
}

const bayLineWidth = 4

// drawHUD writes the driving keys and the tick rate on a dark strip. The race is
// played on a white ground, which pale text disappears into, so the strip is
// what keeps them readable wherever the cars happen to be.
func (g *Gameplay) drawHUD(screen *ebiten.Image) {
	fillRect(screen, 0, 0, float64(screen.Bounds().Dx()), hudHeight, colourHUD)

	// A player who is driving with the screen has every control in front of
	// them with its own name written on it, so naming keys as well would be
	// listing hardware that is not in the room.
	if !g.in.TouchActive() {
		// The triggers are the half of the pad worth naming: a player reaches for
		// the stick to steer without being told, and reaches for it to accelerate
		// too — this is the line that puts that hand right.
		drawText(screen, hint(g.in, "Arrows  Drive", "RT / LT  Drive"), fontPrompt, 14, 13, colourText, text.AlignStart, text.AlignCenter)
		drawText(screen, hint(g.in, "Tab  Swap car", "X  Swap car"), fontPrompt, 116, 13, colourTextMuted, text.AlignStart, text.AlignCenter)
		drawText(screen, hint(g.in, "Enter  Finish", "Y  Finish"), fontPrompt, 218, 13, colourTextMuted, text.AlignStart, text.AlignCenter)
		drawText(screen, hint(g.in, "Esc  Pause", "Start  Pause"), fontPrompt, 320, 13, colourAccent, text.AlignStart, text.AlignCenter)
		// The box overlay is on no pad button, so its key is named either way.
		drawText(screen, "F3  Boxes", fontPrompt, 414, 13, colourTextMuted, text.AlignStart, text.AlignCenter)
	}

	drawText(screen, fmt.Sprintf("%0.0f TPS", ebiten.ActualTPS()), fontPrompt, float64(screen.Bounds().Dx())-14, 13, colourTextMuted, text.AlignEnd, text.AlignCenter)

	// The crash notice is the default rules showing they heard the collision
	// event. Once a crash fails the attempt (#17), this strip goes and the
	// results screen says it instead.
	if g.crash != nil {
		fillRect(screen, 0, hudHeight, float64(screen.Bounds().Dx()), 22, colourHUD)
		drawText(screen, fmt.Sprintf("Crashed into a %s", g.crash.Obstruction), fontPrompt, 14, 37, colourAccent, text.AlignStart, text.AlignCenter)
	}
}

// hudHeight is the strip along the top of the race. The on-screen controls are
// placed under it (touch.go), so it is a number both of them read rather than
// one each.
const hudHeight = 26

// releaseControls takes every car's foot off the pedals and its hands off the
// wheel. Only the car being driven is read each tick, so any other car keeps
// whatever it was last told for as long as it is left alone — which is a car
// still accelerating, on a screen where nothing can lift it. Pausing and
// swapping both hand the driving somewhere else, and both let go first.
func (g *Gameplay) releaseControls() {
	for i := range g.cars {
		g.cars[i].Throttle = 0
		g.cars[i].Brake = 0
		g.cars[i].Steering = 0
	}
}
