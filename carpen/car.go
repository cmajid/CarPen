package carpen

import (
	"image"
	"log"
	"math"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Point struct {
	X, Y float64
}

type Car struct {
	RotateLeft,
	RotateRight,
	Accelerate,
	Decelerate bool
	WheelWidth,
	WheelHeight,
	Width,
	Height int
	WheelRotationStep,
	WheelMaxAngle,
	WheelAngle,
	X,
	Y,
	Rotation,
	Speed,
	Acceleration float64
	Pivot          Pivot
	FrontPivot     FrontPivot
	RearPivot      RearPivot
	DirectionPivot DirectionPivot
	RearPivotAbs   RearPivotAbs
	TempDirPivot   TempDirPivot
	Wheels         []Wheel
	Direction      Direction
	img 		   *ebiten.Image
}

func (c *Car) Init() {
	var err error
	c.img, _, err = ebitenutil.NewImageFromFile("car-yellow.png")
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Car) UpdateDirection() {
	v1 := Vector{X: c.DirectionPivot.X - c.FrontPivot.X, Y: c.DirectionPivot.Y - c.FrontPivot.Y}
	c.Direction = v1.Normalize()
	c.Direction.X *= c.Speed
	c.Direction.Y *= c.Speed
}

func (p Point) Length() float64 {
	return math.Sqrt(math.Pow(p.X, 2) + math.Pow(p.Y, 2))
}

func (car *Car) DrawCar() image.Image {

	dc := gg.NewContext(640, 480)
	car.DrawWheels(dc)
	dc.Translate(car.Pivot.X, car.Pivot.Y)
	dc.Rotate(car.Rotation * math.Pi / 180)
	dc.DrawImage(car.img, -60, -30)
	dc.Fill()
	return dc.Image()
}

func (car *Car) DrawWheels(dc *gg.Context) gg.Context {

	for i := 0; i < len(car.Wheels); i++ {
		if car.RotateLeft {
			if car.WheelAngle > -car.WheelMaxAngle {
				car.WheelAngle = car.WheelAngle - car.WheelRotationStep

			}
		}
		if car.RotateRight {
			if car.WheelAngle < car.WheelMaxAngle {
				car.WheelAngle = car.WheelAngle + car.WheelRotationStep
			}
		}
		dc.Push()
		dc.Translate(car.Pivot.X, car.Pivot.Y)
		dc.Rotate(car.Rotation * math.Pi / 180)
		var o = car.Wheels[i]
		dc.Translate(o.X, o.Y)
		if i < 2 {
			dc.Rotate((car.WheelAngle) * math.Pi / 180)
		}
		dc.SetRGB(0, 0, 0)
		dc.DrawRectangle(-6, -15, float64(car.WheelWidth), float64(car.WheelHeight))
		dc.Fill()
		dc.Pop()

	}
	car.DirectionPivot.X = 50*math.Cos((car.WheelAngle+car.Rotation-90)*math.Pi/180) + car.FrontPivot.X
	car.DirectionPivot.Y = 50*math.Sin((car.WheelAngle+car.Rotation-90)*math.Pi/180) + car.FrontPivot.Y

	car.TempDirPivot.X = 50*math.Cos((car.WheelAngle-90)*math.Pi/180) + car.FrontPivot.X
	car.TempDirPivot.Y = 50*math.Sin((car.WheelAngle-90)*math.Pi/180) + car.FrontPivot.X

	car.UpdateDirection()
	return *dc
}

func (car *Car) Move() error {
	_flag := false
	if car.Accelerate && car.Speed < 6 {
		car.Speed += car.Acceleration
		_flag = true
	} else if car.Speed > 0 {
		car.Speed -= car.Acceleration
		_flag = true
	}
	if car.Decelerate && car.Speed > -3 {
		car.Speed -= car.Acceleration
		_flag = true
	} else if car.Speed < -0.3 {
		car.Speed += car.Acceleration
		_flag = true
	}

	if !_flag {
		car.Speed = 0
	}
	car.Pivot.X += car.Direction.X
	car.Pivot.Y += car.Direction.Y

	// Drift!!! :D
	v := Vector{X: car.Pivot.X - car.RearPivotAbs.X, Y: car.Pivot.Y - car.RearPivotAbs.Y}
	var rotation = math.Atan2(-v.Y, v.X) * 180 / math.Pi

	rotation += 180
	rotation = 360 - rotation - 90
	car.Rotation = rotation

	car.RearPivotAbs = RearPivotAbs{
		X: 160*math.Cos((car.Rotation+90)*math.Pi/180) + car.Pivot.X,
		Y: 160*math.Sin((car.Rotation+90)*math.Pi/180) + car.Pivot.Y,
	}
	return nil
}
