package carpen

import (
	"image"
	"log"
	"math"

	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Car struct {
	IsActive bool
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
	MaxSpeed,
	Acceleration float64
	Pivot          Pivot
	FrontPivot     FrontPivot
	RearPivot      RearPivot
	DirectionPivot DirectionPivot
	RearPivotAbs   RearPivotAbs
	Wheels         []Wheel
	Direction      Direction
	Image          *ebiten.Image
	Color          string
}

func (c *Car) Init() {
	var err error
	c.Image, _, err = ebitenutil.NewImageFromFile("car-" + c.Color + ".png")
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

// Update advances the car by one simulation tick. It must be called from the
// game's Update(), which Ebiten runs at a fixed tick rate; calling it from the
// draw path would tie the car's speed and steering to the display's refresh
// rate.
func (car *Car) Update() error {
	if err := car.Move(); err != nil {
		return err
	}
	car.Steer()
	return nil
}

// Steer advances the front wheels towards the angle the player is asking for
// and recomputes the direction the car travels in.
func (car *Car) Steer() {
	if car.RotateLeft && car.WheelAngle > -car.WheelMaxAngle {
		car.WheelAngle = math.Max(car.WheelAngle-car.WheelRotationStep, -car.WheelMaxAngle)
	}
	if car.RotateRight && car.WheelAngle < car.WheelMaxAngle {
		car.WheelAngle = math.Min(car.WheelAngle+car.WheelRotationStep, car.WheelMaxAngle)
	}

	car.DirectionPivot.X = 50*math.Cos((car.WheelAngle+car.Rotation-90)*math.Pi/180) + car.FrontPivot.X
	car.DirectionPivot.Y = 50*math.Sin((car.WheelAngle+car.Rotation-90)*math.Pi/180) + car.FrontPivot.Y

	car.UpdateDirection()
}

func (car *Car) DrawCar() image.Image {
	dc := gg.NewContext(640, 480)
	car.DrawWheels(dc)
	dc.Translate(car.Pivot.X, car.Pivot.Y)
	dc.Rotate(car.Rotation * math.Pi / 180)
	dc.DrawImage(car.Image, -60, -30)
	dc.Fill()
	return dc.Image()
}

// DrawWheels renders the wheels at their current angle. It only reads car
// state; the angle itself is stepped in Steer().
func (car *Car) DrawWheels(dc *gg.Context) {
	for i := 0; i < len(car.Wheels); i++ {
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
}

func (car *Car) Move() error {

	forceStop := true
	moveFast := car.Accelerate && car.Speed < car.MaxSpeed
	tryToStop := car.Speed > 0
	moveBackward := car.Decelerate && car.Speed > -3
	tryToStopBackward := car.Speed < -0.3

	if moveFast {
		car.Speed += car.Acceleration
		forceStop = false
	} else if tryToStop {
		car.Speed -= car.Acceleration
		forceStop = false
	}

	if moveBackward {
		car.Speed -= car.Acceleration
		forceStop = false
	} else if tryToStopBackward {
		car.Speed += car.Acceleration
		forceStop = false
	}

	if forceStop {
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
