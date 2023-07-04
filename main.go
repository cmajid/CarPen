package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/cmajid/carpen/carpen"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	car  []carpen.Car
	bush []carpen.Bush
	Width,
	Height,
	ActiveCar int
}

func (g *Game) Init() {
	g.Width = 640
	g.Height = 480
	g.ActiveCar = 0

	// BUSH
	bush1 := g.createBush(0, 0)
	bush2 := g.createBush(100, 100)

	g.bush = make([]carpen.Bush, 0)
	g.bush = append(g.bush, bush1, bush2)

	for i := 0; i < len(g.bush); i++ {
		g.bush[i].Init()
	}

	// CAR
	car1 := g.createCar("yellow", 400, 300, true)
	car2 := g.createCar("red", 350, 100, false)
	g.car = make([]carpen.Car, 0)
	g.car = append(g.car, car1, car2)

	for i := 0; i < len(g.car); i++ {
		g.car[i].Init()

		g.car[i].Pivot = carpen.Pivot{X: g.car[i].X + 50, Y: g.car[i].Y + 20}
		g.car[i].DirectionPivot = carpen.DirectionPivot{X: g.car[i].FrontPivot.X, Y: g.car[i].FrontPivot.Y - 50}
		g.car[i].RearPivotAbs = carpen.RearPivotAbs{
			X: 160*math.Cos((g.car[i].Rotation+90)*math.Pi/180) + g.car[i].Pivot.X,
			Y: 160*math.Sin((g.car[i].Rotation+90)*math.Pi/180) + g.car[i].Pivot.Y,
		}

		v1 := carpen.Vector{X: g.car[i].DirectionPivot.X - g.car[i].FrontPivot.X, Y: g.car[i].DirectionPivot.Y - g.car[i].FrontPivot.Y}
		g.car[i].Direction = v1.Normalize()
	}
}

func (g *Game) createCar(color string, x float64, y float64, active bool) carpen.Car {
	car := carpen.Car{
		Color:       color,
		IsActive:    active,
		RotateLeft:  false,
		RotateRight: false,
		Accelerate:  false,
		Decelerate:  false,

		MaxSpeed:          6,
		WheelWidth:        12,
		WheelHeight:       30,
		WheelRotationStep: 0.6,
		WheelMaxAngle:     45,
		WheelAngle:        0,
		Width:             100,
		Height:            200,
		X:                 x,
		Y:                 y,
		FrontPivot:        carpen.FrontPivot{X: 0, Y: 0},
		RearPivot:         carpen.RearPivot{X: 0, Y: 160},
		Rotation:          90,
		Wheels: []carpen.Wheel{
			{X: -40, Y: 10},
			{X: 45, Y: 10},
			{X: -41, Y: 145},
			{X: 46, Y: 145},
		},
		Speed:        0,
		Acceleration: 0.2,
	}

	return car
}

func (*Game) createBush(x, y float64) carpen.Bush {
	bush2 := carpen.Bush{
		Width:  109,
		Height: 109,
		Direction: carpen.Direction{
			X: x,
			Y: y,
		},
	}
	return bush2
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.car[g.ActiveCar].Accelerate = true
		g.car[g.ActiveCar].Speed += g.car[g.ActiveCar].Acceleration
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.car[g.ActiveCar].Decelerate = true
		g.car[g.ActiveCar].Speed -= g.car[g.ActiveCar].Acceleration
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.car[g.ActiveCar].RotateLeft = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.car[g.ActiveCar].RotateRight = true
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyUp) {
		g.car[g.ActiveCar].Accelerate = false
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		g.car[g.ActiveCar].Decelerate = false
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyRight) {
		g.car[g.ActiveCar].RotateRight = false
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyLeft) {
		g.car[g.ActiveCar].RotateLeft = false
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyTab) {
		if g.ActiveCar == 1 {
			g.ActiveCar = 0
		} else {
			g.ActiveCar = 1
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)

	for i := 0; i < len(g.car); i++ {
		g.car[i].Move()
		img := g.car[i].DrawCar()
		screen.DrawImage(ebiten.NewImageFromImage(img), &ebiten.DrawImageOptions{})
	}
	for _, b := range g.bush {
		bushImage := b.DrawBush()
		opt := &ebiten.DrawImageOptions{}
		opt.GeoM.Apply(b.Direction.X, b.Direction.Y)
		screen.DrawImage(ebiten.NewImageFromImage(bushImage), opt)
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.Width, g.Height
}

func main() {
	g := &Game{}
	g.Init()

	ebiten.SetWindowSize(g.Width, g.Height)
	ebiten.SetWindowTitle("Car Pen")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
