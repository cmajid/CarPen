package scene

import (
	"fmt"
	"image/color"

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

	g.walls = lotWalls(level.Lot)
	g.OnCollision = func(event carpen.CollisionEvent) { g.crash = &event }

	for i := range g.cars {
		g.cars[i].Init()

		g.cars[i].Pivot = carpen.Pivot{X: g.cars[i].X + 50, Y: g.cars[i].Y + 20}
		g.cars[i].DirectionPivot = carpen.DirectionPivot{X: g.cars[i].FrontPivot.X, Y: g.cars[i].FrontPivot.Y - 50}
		g.cars[i].UpdateRearPivotAbs()

		v1 := carpen.Vector{X: g.cars[i].DirectionPivot.X - g.cars[i].FrontPivot.X, Y: g.cars[i].DirectionPivot.Y - g.cars[i].FrontPivot.Y}
		g.cars[i].Direction = v1.Normalize()
	}

	event, hit := g.findCollision()
	g.colliding, g.touching = hit, event.Obstruction

	return g
}

func newCar(paint string, x float64, y float64, rotate float64, active bool) carpen.Car {
	car := carpen.Car{
		Color:       paint,
		IsActive:    active,
		RotateLeft:  false,
		RotateRight: false,
		Accelerate:  false,
		Decelerate:  false,

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
func lotWalls(lot carpen.Lot) []*carpen.OBB {
	const thickness = 100.0
	w, h := lot.Width, lot.Height

	return []*carpen.OBB{
		carpen.NewOBB(w/2, -thickness/2, w+2*thickness, thickness, 0),  // top
		carpen.NewOBB(w/2, h+thickness/2, w+2*thickness, thickness, 0), // bottom
		carpen.NewOBB(-thickness/2, h/2, thickness, h+2*thickness, 0),  // left
		carpen.NewOBB(w+thickness/2, h/2, thickness, h+2*thickness, 0), // right
	}
}

func (g *Gameplay) Update() (Scene, error) {
	g.fade.update()

	if justPressed(g.in, actionCancel) {
		return newPause(g), nil
	}
	// Standing in for the win condition, which cannot tell the race is over
	// until the cars can hit something (#20): Enter ends the race by hand.
	if justPressed(g.in, actionFinish) {
		return newResults(g.in, g.level), nil
	}

	// The input handlers only record intent; every change to Speed is made by
	// Car.Move(), which is the one place that honours MaxSpeed and the reverse
	// limit.
	if justPressed(g.in, actionThrottle) {
		g.cars[g.activeCar].Accelerate = true
	}
	if justPressed(g.in, actionBrake) {
		g.cars[g.activeCar].Decelerate = true
	}
	if justPressed(g.in, actionSteerLeft) {
		g.cars[g.activeCar].RotateLeft = true
	}
	if justPressed(g.in, actionSteerRight) {
		g.cars[g.activeCar].RotateRight = true
	}

	if justReleased(g.in, actionThrottle) {
		g.cars[g.activeCar].Accelerate = false
	}
	if justReleased(g.in, actionBrake) {
		g.cars[g.activeCar].Decelerate = false
	}
	if justReleased(g.in, actionSteerRight) {
		g.cars[g.activeCar].RotateRight = false
	}
	if justReleased(g.in, actionSteerLeft) {
		g.cars[g.activeCar].RotateLeft = false
	}
	// Swapping hands the keys to the next car in the level, which is how the
	// prototype's two cars are still both drivable. It comes to nothing on a
	// level with a single car, and goes for good once a parked car is something
	// to be avoided rather than driven (#20).
	if justReleased(g.in, actionSwapCar) {
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
	// The lot is the ground the level is played on. It is the whole screen on
	// every level so far, and anything the screen has left over is drawn in the
	// menus' dark ink rather than in more lot.
	screen.Fill(colourInk)
	fillRect(screen, 0, 0, g.level.Lot.Width, g.level.Lot.Height, color.White)

	g.drawBay(screen)

	for i := range g.cars {
		g.cars[i].DrawCar(screen)
	}
	for i := range g.bushes {
		g.bushes[i].Draw(screen)
	}

	if g.debugOBB {
		g.drawOBBs(screen)
	}

	g.drawHUD(screen)
	g.fade.draw(screen)
}

// drawOBBs outlines every box collision sees — the development overlay behind
// F3. The active car's box turns red for exactly the ticks it overlaps
// something, so a clipped corner is seen the moment it happens rather than
// found later in the crash notice. Reading an entity's OBB() re-places its box
// but steps no physics, so drawing it here breaks nothing about the fixed
// tick rate.
func (g *Gameplay) drawOBBs(screen *ebiten.Image) {
	for i := range g.cars {
		colour := colourAccent
		if i == g.activeCar && g.touching != "" {
			colour = colourDanger
		}
		strokeOBB(screen, g.cars[i].OBB(), colour)
	}
	for i := range g.bushes {
		circle := g.bushes[i].Collider()
		strokeCircle(screen, circle.Center(), circle.Radius(), 2, colourAccent)
	}
	for _, wall := range g.walls {
		strokeOBB(screen, wall, colourBay)
	}

	g.drawDebugStatus(screen)
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

func strokeOBB(dst *ebiten.Image, obb *carpen.OBB, colour color.Color) {
	outline := obb.Outline()
	for i := range outline {
		strokeLine(dst, outline[i], outline[(i+1)%len(outline)], 2, colour)
	}
}

// drawBay marks out the space the level asks the player to park in: the two
// sides and the back of the bay, painted on the ground and left open on the
// edge the car drives in over. It is drawn before the cars, so a car parked in
// the bay stands on the lines rather than under them.
func (g *Gameplay) drawBay(screen *ebiten.Image) {
	nearLeft, nearRight, farLeft, farRight := g.level.Bay.Corners()

	strokeLine(screen, nearLeft, farLeft, bayLineWidth, colourBay)
	strokeLine(screen, nearRight, farRight, bayLineWidth, colourBay)
	strokeLine(screen, farLeft, farRight, bayLineWidth, colourBay)
}

const bayLineWidth = 4

// drawHUD writes the driving keys and the tick rate on a dark strip. The race is
// played on a white ground, which pale text disappears into, so the strip is
// what keeps them readable wherever the cars happen to be.
func (g *Gameplay) drawHUD(screen *ebiten.Image) {
	fillRect(screen, 0, 0, float64(screen.Bounds().Dx()), 26, colourHUD)

	// The triggers are the half of the pad worth naming: a player reaches for
	// the stick to steer without being told, and reaches for it to accelerate
	// too — this is the line that puts that hand right.
	drawText(screen, hint(g.in, "Arrows  Drive", "RT / LT  Drive"), fontPrompt, 14, 13, colourText, text.AlignStart, text.AlignCenter)
	drawText(screen, hint(g.in, "Tab  Swap car", "X  Swap car"), fontPrompt, 116, 13, colourTextMuted, text.AlignStart, text.AlignCenter)
	drawText(screen, hint(g.in, "Enter  Finish", "Y  Finish"), fontPrompt, 218, 13, colourTextMuted, text.AlignStart, text.AlignCenter)
	drawText(screen, hint(g.in, "Esc  Pause", "Start  Pause"), fontPrompt, 320, 13, colourAccent, text.AlignStart, text.AlignCenter)
	// The box overlay is on no pad button, so its key is named either way.
	drawText(screen, "F3  Boxes", fontPrompt, 414, 13, colourTextMuted, text.AlignStart, text.AlignCenter)

	drawText(screen, fmt.Sprintf("%0.0f TPS", ebiten.ActualTPS()), fontPrompt, float64(screen.Bounds().Dx())-14, 13, colourTextMuted, text.AlignEnd, text.AlignCenter)

	// The crash notice is the default rules showing they heard the collision
	// event. Once a crash fails the attempt (#17), this strip goes and the
	// results screen says it instead.
	if g.crash != nil {
		fillRect(screen, 0, 26, float64(screen.Bounds().Dx()), 22, colourHUD)
		drawText(screen, fmt.Sprintf("Crashed into a %s", g.crash.Obstruction), fontPrompt, 14, 37, colourAccent, text.AlignStart, text.AlignCenter)
	}
}

// releaseControls lets go of every key the player was holding. Driving is worked
// out from key presses and releases rather than from what is held down, so a key
// released while another scene is running is never seen going up; without this
// the car would drive on by itself once the race came back.
func (g *Gameplay) releaseControls() {
	for i := range g.cars {
		g.cars[i].Accelerate = false
		g.cars[i].Decelerate = false
		g.cars[i].RotateLeft = false
		g.cars[i].RotateRight = false
	}
}
