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
	car carpen.Car
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.car.Accelerate = true
		g.car.Speed += g.car.Acceleration
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.car.Decelerate = true
		g.car.Speed -= g.car.Acceleration
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.car.RotateLeft = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.car.RotateRight = true
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyUp) {
		g.car.Accelerate = false
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		g.car.Decelerate = false
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyRight) {
		g.car.RotateRight = false
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyLeft) {
		g.car.RotateLeft = false
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)
	g.car.Move()
	img := g.car.DrawCar()
	em := ebiten.NewImageFromImage(img)
	screen.DrawImage(em, &ebiten.DrawImageOptions{})
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func (g *Game) Init() {
	g.car = carpen.Car{
		RotateLeft:  false,
		RotateRight: false,
		Accelerate:  false,
		Decelerate:  false,

		WheelWidth:        12,
		WheelHeight:       30,
		WheelRotationStep: 0.6,
		WheelMaxAngle:     45,
		WheelAngle:        0,
		Width:             100,
		Height:            200,
		X:                 350,
		Y:                 60,
		FrontPivot:        carpen.FrontPivot{X: 0, Y: 0},
		RearPivot:         carpen.RearPivot{X: 0, Y: 160},
		Rotation:          90,

		TempDirPivot: carpen.TempDirPivot{X: 0, Y: 0},
		Wheels: []carpen.Wheel{
			{X: -50, Y: 0},
			{X: 50, Y: 0},
			{X: -50, Y: 160},
			{X: 50, Y: 160},
		},
		Speed:        0,
		Acceleration: 0.2,
	}

	g.car.Pivot = carpen.Pivot{X: g.car.X + 50, Y: g.car.Y + 20}
	g.car.DirectionPivot = carpen.DirectionPivot{X: g.car.FrontPivot.X, Y: g.car.FrontPivot.Y - 50}
	g.car.RearPivotAbs = carpen.RearPivotAbs{
		X: 160*math.Cos((g.car.Rotation+90)*math.Pi/180) + g.car.Pivot.X,
		Y: 160*math.Sin((g.car.Rotation+90)*math.Pi/180) + g.car.Pivot.Y,
	}

	v1 := carpen.Vector{X: g.car.DirectionPivot.X - g.car.FrontPivot.X, Y: g.car.DirectionPivot.Y - g.car.FrontPivot.Y}
	g.car.Direction = v1.Normalize()
}

func main() {
	g := &Game{}
	g.Init()

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Car Pen")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
