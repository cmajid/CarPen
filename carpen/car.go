package carpen

import (
	"image/color"
	"log"
	"math"

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
	WheelHeight int
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

	// wheelImage is the black rectangle every wheel is drawn from. It is built
	// once in Init() and reused for all four wheels of every frame.
	wheelImage *ebiten.Image
}

func (c *Car) Init() {
	var err error
	c.Image, _, err = ebitenutil.NewImageFromFileSystem(assets, "assets/car-"+c.Color+".png")
	if err != nil {
		log.Fatal(err)
	}

	c.wheelImage = ebiten.NewImage(c.WheelWidth, c.WheelHeight)
	c.wheelImage.Fill(color.Black)
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
func (car *Car) Update() {
	car.Move()
	car.Steer()
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

// DrawCar blits the wheels and the body straight onto screen. The sprites are
// placed with a GeoM per draw, so a frame allocates no image and no off-screen
// buffer however many cars there are.
func (car *Car) DrawCar(screen *ebiten.Image) {
	car.DrawWheels(screen)

	opt := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	opt.GeoM = car.bodyGeoM()
	screen.DrawImage(car.Image, opt)
}

// DrawWheels renders the wheels at their current angle. It only reads car
// state; the angle itself is stepped in Steer().
func (car *Car) DrawWheels(screen *ebiten.Image) {
	opt := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	for i := 0; i < len(car.Wheels); i++ {
		opt.GeoM = car.wheelGeoM(i)
		screen.DrawImage(car.wheelImage, opt)
	}
}

// bodyGeoM places the body sprite: its (60, 30) point sits on the car's pivot,
// and the sprite turns with the car.
func (car *Car) bodyGeoM() ebiten.GeoM {
	var g ebiten.GeoM
	g.Translate(-60, -30)
	g.Rotate(car.Rotation * math.Pi / 180)
	g.Translate(car.Pivot.X, car.Pivot.Y)
	return g
}

// wheelGeoM places wheel i: the rectangle is centred on the wheel's offset from
// the pivot, the two front wheels (i < 2) additionally turn with the steering
// angle, and the whole thing turns with the car.
//
// Ebiten applies a GeoM's calls in the order they are made, so these read as
// the innermost transform first — the reverse of a nested push/pop stack.
func (car *Car) wheelGeoM(i int) ebiten.GeoM {
	var g ebiten.GeoM
	g.Translate(-6, -15)
	if i < 2 {
		g.Rotate(car.WheelAngle * math.Pi / 180)
	}
	g.Translate(car.Wheels[i].X, car.Wheels[i].Y)
	g.Rotate(car.Rotation * math.Pi / 180)
	g.Translate(car.Pivot.X, car.Pivot.Y)
	return g
}

// UpdateRearPivotAbs recomputes where the rear axle sits on screen: RearPivot is
// the axle's offset from the pivot in the car's own frame, so it turns with the
// car and is then placed at the pivot.
func (car *Car) UpdateRearPivotAbs() {
	sin, cos := math.Sincos(car.Rotation * math.Pi / 180)
	car.RearPivotAbs = RearPivotAbs{
		X: car.RearPivot.X*cos - car.RearPivot.Y*sin + car.Pivot.X,
		Y: car.RearPivot.X*sin + car.RearPivot.Y*cos + car.Pivot.Y,
	}
}

func (car *Car) Move() {

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

	car.UpdateRearPivotAbs()
}
