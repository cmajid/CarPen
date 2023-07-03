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
	car  carpen.Car
	bush []carpen.Bush
	Width,
	Height int
}

func (g *Game) Init() {
	g.Width = 640
	g.Height = 480
	g.car = carpen.Car{
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
		X:                 350,
		Y:                 60,
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
	g.car.Init()

	bush1 := g.createBush(0, 0)
	bush2 := g.createBush(100, 100)

	g.bush = make([]carpen.Bush, 0)
	g.bush = append(g.bush, bush1, bush2)

	for i := 0; i < len(g.bush); i++ {
		g.bush[i].Init()
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
	screen.DrawImage(ebiten.NewImageFromImage(img), &ebiten.DrawImageOptions{})

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
