package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/cmajid/carpen/cp_car"
	"github.com/cmajid/carpen/cp_pivot"
	"github.com/cmajid/carpen/cp_vector"
	"github.com/cmajid/carpen/cp_wheel"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	car cp_car.Car
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

func init() {

}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)
	//ebitenutil.DebugPrint(screen, "Hello, World!")
	g.car.Move()
	img := g.car.DrawCar()
	em := ebiten.NewImageFromImage(img)
	screen.DrawImage(em, &ebiten.DrawImageOptions{})

	// m := cp_clock.ClockImage(g.time)
	// em := ebiten.NewImageFromImage(m)
	// screen.DrawImage(em, &ebiten.DrawImageOptions{})

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()))

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	g := &Game{}

	g.car = cp_car.Car{

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
		FrontPivot:        cp_pivot.FrontPivot{X: 0, Y: 0},
		RearPivot:         cp_pivot.RearPivot{X: 0, Y: 160},
		Rotation:          90,

		TempDirPivot: cp_pivot.TempDirPivot{X: 0, Y: 0},
		Wheels: []cp_wheel.Wheel{
			{X: -50, Y: 0},
			{X: 50, Y: 0},
			{X: -50, Y: 160},
			{X: 50, Y: 160},
		},
		Speed:        0,
		Acceleration: 0.2,
	}

	g.car.Pivot = cp_pivot.Pivot{X: g.car.X + 50, Y: g.car.Y + 20}
	g.car.DirectionPivot = cp_pivot.DirectionPivot{X: g.car.FrontPivot.X, Y: g.car.FrontPivot.Y - 50}
	g.car.RearPivotAbs = cp_pivot.RearPivotAbs{
		X: 160*math.Cos((g.car.Rotation+90)*math.Pi/180) + g.car.Pivot.X,
		Y: 160*math.Sin((g.car.Rotation+90)*math.Pi/180) + g.car.Pivot.Y,
	}

	v1 := cp_vector.Vector{X: g.car.DirectionPivot.X - g.car.FrontPivot.X, Y: g.car.DirectionPivot.Y - g.car.FrontPivot.Y}
	g.car.Direction = v1.Normalize()

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Car Pen")

	// pt1 := car.Point{X: 2, Y: 3}
	// fmt.Println(pt1)
	// fmt.Println(pt1.Length())

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
