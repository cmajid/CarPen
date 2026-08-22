package carpen

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Car struct {
	IsActive bool

	// Throttle and Brake are how hard each pedal is being asked for, from 0
	// (not at all) to 1 (to the floor), and Steering is where the front wheels
	// are being asked to point, from −1 (full left) through 0 to 1 (full
	// right). A key has only the two ends to offer; a trigger and a stick have
	// everything in between, which is the whole reason these are not the
	// booleans they were.
	//
	// Brake is the one pedal that slows the car and then backs it up, as it
	// always has.
	Throttle,
	Brake,
	Steering float64

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

	// bodyWidth and bodyHeight are the body sprite's size, kept as numbers so
	// the OBB can be worked out without asking the image, and the obb itself
	// is built once and re-placed on every OBB() call.
	bodyWidth, bodyHeight float64
	obb                   *OBB
}

func (c *Car) Init() {
	c.Image = CarImage(c.Color)
	c.bodyWidth = float64(c.Image.Bounds().Dx())
	c.bodyHeight = float64(c.Image.Bounds().Dy())

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
//
// Steering says where the wheels are wanted, not which way to turn them: the
// deflection picks an angle between straight and full lock, and the wheels
// travel to it at WheelRotationStep a tick. A key only ever asks for full lock,
// so holding an arrow steers exactly as it always did, while a stick pushed
// half way settles half way.
//
// Asking for nothing is not the same as asking for straight. A wheel let go of
// stays where it was left, as it always has, and is straightened by steering
// the other way rather than by letting go: self-centring would be a change to
// how the car drives rather than to how finely it is steered.
func (car *Car) Steer() {
	if car.Steering != 0 {
		target := clamp(car.Steering, -1, 1) * car.WheelMaxAngle
		if target < car.WheelAngle {
			car.WheelAngle = math.Max(car.WheelAngle-car.WheelRotationStep, target)
		} else {
			car.WheelAngle = math.Min(car.WheelAngle+car.WheelRotationStep, target)
		}
	}

	car.DirectionPivot.X = 50*math.Cos((car.WheelAngle+car.Rotation-90)*math.Pi/180) + car.FrontPivot.X
	car.DirectionPivot.Y = 50*math.Sin((car.WheelAngle+car.Rotation-90)*math.Pi/180) + car.FrontPivot.Y

	car.UpdateDirection()
}

// DrawCar blits the wheels and the body straight onto screen. The sprites are
// placed with a GeoM per draw, so a frame allocates no image and no off-screen
// buffer however many cars there are.
//
// origin is where the level's own (0, 0) falls on the destination. A car knows
// where it stands in the level and nothing about the screen showing it, so
// everything about the shape of that screen — how much ground there is around
// the level, and therefore how far in the level has been placed — arrives here
// as this one offset.
func (car *Car) DrawCar(screen *ebiten.Image, origin Vector) {
	car.DrawWheels(screen, origin)

	opt := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	opt.GeoM = car.bodyGeoM()
	opt.GeoM.Translate(origin.X, origin.Y)
	screen.DrawImage(car.Image, opt)
}

// DrawWheels renders the wheels at their current angle. It only reads car
// state; the angle itself is stepped in Steer().
func (car *Car) DrawWheels(screen *ebiten.Image, origin Vector) {
	opt := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	for i := 0; i < len(car.Wheels); i++ {
		opt.GeoM = car.wheelGeoM(i)
		opt.GeoM.Translate(origin.X, origin.Y)
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

// The car's collision shape is drawn in a little from the sprite and has its
// corners cut, both tuned by eye against the artwork: the sprite's paint
// reaches within a few pixels of its edge, and its corners are rounded. A
// touch of forgiveness here makes a near miss look like the near miss it was;
// raise the inset for a more forgiving game.
const (
	carBodyInset   = 4.0
	carCornerBevel = 16.0
)

// OBB returns the car's collision shape: the body sprite's footprint on the
// lot, pulled in by carBodyInset with carCornerBevel corners, turned with the
// car. The sprite hangs (-60, -30) from the pivot in the car's own frame (see
// bodyGeoM), so its centre sits at that offset plus half the sprite, turned
// by Rotation and placed at the pivot.
func (car *Car) OBB() *OBB {
	sin, cos := math.Sincos(car.Rotation * math.Pi / 180)
	dx := car.bodyWidth/2 - 60
	dy := car.bodyHeight/2 - 30
	centerX := dx*cos - dy*sin + car.Pivot.X
	centerY := dx*sin + dy*cos + car.Pivot.Y

	if car.obb == nil {
		car.obb = NewBeveledOBB(centerX, centerY,
			car.bodyWidth-2*carBodyInset, car.bodyHeight-2*carBodyInset,
			carCornerBevel, car.Rotation)
	}
	car.obb.SetTransform(centerX, centerY, car.Rotation)
	return car.obb
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

// Move is the one place Speed changes. A pedal held down part way puts down
// that much of the car's acceleration, so a feathered trigger pulls away gently
// where a key, which has only the floor to offer, pulls away as briskly as it
// always did. The car slowing itself is not the player asking for anything, so
// coasting and the crawl back to a stop are unscaled.
func (car *Car) Move() {
	throttle := clamp(car.Throttle, 0, 1)
	brake := clamp(car.Brake, 0, 1)

	forceStop := true
	moveFast := throttle > 0 && car.Speed < car.MaxSpeed
	tryToStop := car.Speed > 0
	moveBackward := brake > 0 && car.Speed > -3
	tryToStopBackward := car.Speed < -0.3

	if moveFast {
		car.Speed += car.Acceleration * throttle
		forceStop = false
	} else if tryToStop {
		car.Speed -= car.Acceleration
		forceStop = false
	}

	if moveBackward {
		car.Speed -= car.Acceleration * brake
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

// clamp holds v inside low..high. The controls are taken as read rather than
// trusted: a pad driver that reports a stick a shade past its stop must not be
// able to steer the wheels past full lock.
func clamp(v, low, high float64) float64 {
	return math.Min(math.Max(v, low), high)
}
