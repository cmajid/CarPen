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
	cars      []carpen.Car
	bushes    []carpen.Bush
	activeCar int
	fade      fade
}

// newGameplay lays out a fresh world. Where everything stands is written out by
// hand here for now; issue #19 moves the layout into level data.
func newGameplay(in Input) *Gameplay {
	g := &Gameplay{in: in}

	g.bushes = []carpen.Bush{newBush(0, 0), newBush(100, 100)}
	for i := range g.bushes {
		g.bushes[i].Init()
	}

	g.cars = []carpen.Car{
		newCar("yellow", 400, 300, 0, true),
		newCar("red", 350, 100, 90, false),
	}
	for i := range g.cars {
		g.cars[i].Init()

		g.cars[i].Pivot = carpen.Pivot{X: g.cars[i].X + 50, Y: g.cars[i].Y + 20}
		g.cars[i].DirectionPivot = carpen.DirectionPivot{X: g.cars[i].FrontPivot.X, Y: g.cars[i].FrontPivot.Y - 50}
		g.cars[i].UpdateRearPivotAbs()

		v1 := carpen.Vector{X: g.cars[i].DirectionPivot.X - g.cars[i].FrontPivot.X, Y: g.cars[i].DirectionPivot.Y - g.cars[i].FrontPivot.Y}
		g.cars[i].Direction = v1.Normalize()
	}

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

func (g *Gameplay) Update() (Scene, error) {
	g.fade.update()

	if g.in.IsKeyJustPressed(ebiten.KeyEscape) {
		return newPause(g), nil
	}
	// Standing in for the win condition, which cannot tell the race is over
	// until the cars can hit something (#20): Enter ends the race by hand.
	if g.in.IsKeyJustPressed(ebiten.KeyEnter) {
		return newResults(g.in), nil
	}

	// The key handlers only record intent; every change to Speed is made by
	// Car.Move(), which is the one place that honours MaxSpeed and the reverse
	// limit.
	if g.in.IsKeyJustPressed(ebiten.KeyUp) {
		g.cars[g.activeCar].Accelerate = true
	}
	if g.in.IsKeyJustPressed(ebiten.KeyDown) {
		g.cars[g.activeCar].Decelerate = true
	}
	if g.in.IsKeyJustPressed(ebiten.KeyLeft) {
		g.cars[g.activeCar].RotateLeft = true
	}
	if g.in.IsKeyJustPressed(ebiten.KeyRight) {
		g.cars[g.activeCar].RotateRight = true
	}

	if g.in.IsKeyJustReleased(ebiten.KeyUp) {
		g.cars[g.activeCar].Accelerate = false
	}
	if g.in.IsKeyJustReleased(ebiten.KeyDown) {
		g.cars[g.activeCar].Decelerate = false
	}
	if g.in.IsKeyJustReleased(ebiten.KeyRight) {
		g.cars[g.activeCar].RotateRight = false
	}
	if g.in.IsKeyJustReleased(ebiten.KeyLeft) {
		g.cars[g.activeCar].RotateLeft = false
	}
	if g.in.IsKeyJustReleased(ebiten.KeyTab) {
		if g.activeCar == 1 {
			g.activeCar = 0
		} else {
			g.activeCar = 1
		}
	}

	for i := range g.cars {
		g.cars[i].Update()
	}

	return nil, nil
}

func (g *Gameplay) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)

	for i := range g.cars {
		g.cars[i].DrawCar(screen)
	}
	for i := range g.bushes {
		g.bushes[i].Draw(screen)
	}

	g.drawHUD(screen)
	g.fade.draw(screen)
}

// drawHUD writes the driving keys and the tick rate on a dark strip. The race is
// played on a white ground, which pale text disappears into, so the strip is
// what keeps them readable wherever the cars happen to be.
func (g *Gameplay) drawHUD(screen *ebiten.Image) {
	fillRect(screen, 0, 0, float64(screen.Bounds().Dx()), 26, colourHUD)

	drawText(screen, "Arrows  Drive", fontPrompt, 14, 13, colourText, text.AlignStart, text.AlignCenter)
	drawText(screen, "Tab  Swap car", fontPrompt, 116, 13, colourTextMuted, text.AlignStart, text.AlignCenter)
	drawText(screen, "Enter  Finish", fontPrompt, 218, 13, colourTextMuted, text.AlignStart, text.AlignCenter)
	drawText(screen, "Esc  Pause", fontPrompt, 320, 13, colourAccent, text.AlignStart, text.AlignCenter)

	drawText(screen, fmt.Sprintf("%0.0f TPS", ebiten.ActualTPS()), fontPrompt, float64(screen.Bounds().Dx())-14, 13, colourTextMuted, text.AlignEnd, text.AlignCenter)
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
