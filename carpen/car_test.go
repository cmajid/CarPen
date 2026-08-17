package carpen

import (
	"math"
	"reflect"
	"testing"
)

// newTestCar builds a car in the same shape as main.go's createCar, without
// the image (loading one needs a running graphics context).
func newTestCar() *Car {
	c := &Car{
		MaxSpeed:          6,
		WheelWidth:        12,
		WheelHeight:       30,
		WheelRotationStep: 2.4,
		WheelMaxAngle:     45,
		X:                 400,
		Y:                 300,
		RearPivot:         RearPivot{X: 0, Y: 160},
		Wheels: []Wheel{
			{X: -40, Y: 10},
			{X: 45, Y: 10},
			{X: -41, Y: 145},
			{X: 46, Y: 145},
		},
		Speed:        5,
		Acceleration: 0.2,
	}
	c.Pivot = Pivot{X: c.X + 50, Y: c.Y + 20}
	c.DirectionPivot = DirectionPivot{X: c.FrontPivot.X, Y: c.FrontPivot.Y - 50}
	c.UpdateRearPivotAbs()
	v := Vector{X: c.DirectionPivot.X - c.FrontPivot.X, Y: c.DirectionPivot.Y - c.FrontPivot.Y}
	c.Direction = v.Normalize()
	return c
}

func TestUpdateStepsWheelAngleOncePerTick(t *testing.T) {
	c := newTestCar()
	c.Steering = -1

	c.Update()

	if got, want := c.WheelAngle, -c.WheelRotationStep; got != want {
		t.Errorf("WheelAngle after one tick = %v, want %v", got, want)
	}
}

// Full lock is as far as the wheels go, whatever is asking. The reading beyond
// the ends is the guard against a pad reporting a stick past its own stop.
func TestSteerClampsToMaxAngle(t *testing.T) {
	for _, tc := range []struct {
		name     string
		steering float64
		want     float64
	}{
		{name: "left", steering: -1, want: -45},
		{name: "right", steering: 1, want: 45},
		{name: "past the left stop", steering: -1.4, want: -45},
		{name: "past the right stop", steering: 1.4, want: 45},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestCar()
			c.Steering = tc.steering

			for i := 0; i < 200; i++ {
				c.Steer()
			}

			if c.WheelAngle != tc.want {
				t.Errorf("WheelAngle = %v, want %v", c.WheelAngle, tc.want)
			}
		})
	}
}

// The whole point of a stick over a pair of arrow keys: half over is half the
// angle, and the wheels stop there rather than winding on to full lock the way
// a held key does.
func TestSteerSettlesWhereTheWheelIsAskedFor(t *testing.T) {
	for _, tc := range []struct {
		steering float64
		want     float64
	}{
		{steering: -0.5, want: -22.5},
		{steering: 0.25, want: 11.25},
		{steering: 0.75, want: 33.75},
	} {
		c := newTestCar()
		c.Steering = tc.steering

		for i := 0; i < 200; i++ {
			c.Steer()
		}

		if !closeTo(c.WheelAngle, tc.want) {
			t.Errorf("steering %v settled the wheels at %v, want %v", tc.steering, c.WheelAngle, tc.want)
		}
	}
}

// Asking for less than the wheels are already giving winds them back, which is
// how a stick eased off straightens the car without being pushed the other way.
func TestSteerWindsBackTowardsALesserAngle(t *testing.T) {
	c := newTestCar()
	c.Steering = -1
	for i := 0; i < 200; i++ {
		c.Steer()
	}

	c.Steering = -0.25
	for i := 0; i < 200; i++ {
		c.Steer()
	}

	if !closeTo(c.WheelAngle, -11.25) {
		t.Errorf("WheelAngle = %v, want the wheels wound back to the lesser angle", c.WheelAngle)
	}
}

// Asking for nothing is not asking for straight. A wheel let go of stays where
// it was left — the car has never had self-centring, and gaining it would be a
// change to how it drives rather than to how finely it is steered.
func TestSteerHoldsTheWheelWhenNothingIsAsked(t *testing.T) {
	c := newTestCar()
	c.Steering = -1
	c.Steer()
	held := c.WheelAngle

	c.Steering = 0
	for i := 0; i < 100; i++ {
		c.Steer()
	}

	if c.WheelAngle != held {
		t.Errorf("WheelAngle drifted to %v with nothing asked, want it left at %v", c.WheelAngle, held)
	}
}

func TestMoveAdvancesPivot(t *testing.T) {
	c := newTestCar()
	before := c.Pivot

	c.Move()

	if c.Pivot == before {
		t.Errorf("Pivot did not move: still %v", c.Pivot)
	}
}

// Holding the accelerator must not wind the speed up without end: Move() is the
// only writer of Speed, and it stops adding once MaxSpeed is reached, so the
// speed settles within one acceleration step of the limit.
func TestMoveClampsForwardSpeedToMaxSpeed(t *testing.T) {
	c := newTestCar()
	c.Throttle = 1

	for i := 0; i < 200; i++ {
		c.Move()

		if c.Speed > c.MaxSpeed+c.Acceleration {
			t.Fatalf("Speed reached %v after %d ticks, want no more than %v", c.Speed, i+1, c.MaxSpeed+c.Acceleration)
		}
	}

	if c.Speed < c.MaxSpeed-c.Acceleration {
		t.Errorf("Speed settled at %v, want within %v of MaxSpeed %v", c.Speed, c.Acceleration, c.MaxSpeed)
	}
}

// Reverse has its own, lower limit, so the car backs up slowly however long the
// key is held.
func TestMoveClampsReverseSpeed(t *testing.T) {
	const maxReverse = -3.0

	c := newTestCar()
	c.Brake = 1

	for i := 0; i < 200; i++ {
		c.Move()

		if c.Speed < maxReverse-c.Acceleration {
			t.Fatalf("Speed reached %v after %d ticks, want no less than %v", c.Speed, i+1, maxReverse-c.Acceleration)
		}
	}

	if c.Speed > maxReverse+c.Acceleration {
		t.Errorf("Speed settled at %v, want within %v of %v", c.Speed, c.Acceleration, maxReverse)
	}
}

// A pedal held part way puts down that much of the acceleration, which is what
// a trigger buys over a key: the same car pulled away gently.
func TestMovePutsDownAsMuchOfThePedalAsIsAskedFor(t *testing.T) {
	for _, tc := range []struct {
		name     string
		throttle float64
		brake    float64
		want     float64
	}{
		{name: "throttle to the floor", throttle: 1, want: 0.2},
		{name: "throttle half way", throttle: 0.5, want: 0.1},
		{name: "throttle a quarter", throttle: 0.25, want: 0.05},
		{name: "brake half way", brake: 0.5, want: -0.1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestCar()
			c.Speed = 0 // from rest, so nothing else is moving the speed
			c.Throttle, c.Brake = tc.throttle, tc.brake

			c.Move()

			if !closeTo(c.Speed, tc.want) {
				t.Errorf("Speed after one tick = %v, want %v", c.Speed, tc.want)
			}
		})
	}
}

// A pedal barely off its stop is still the player asking for something, and the
// car has to answer it rather than round it away to nothing.
func TestMoveCreepsOnTheGentlestThrottle(t *testing.T) {
	c := newTestCar()
	c.Speed = 0
	c.Throttle = 0.05

	for i := 0; i < 10; i++ {
		c.Move()
	}

	if c.Speed <= 0 {
		t.Errorf("Speed after ten gentle ticks = %v, want the car creeping forward", c.Speed)
	}
	if c.Speed >= c.MaxSpeed {
		t.Errorf("Speed after ten gentle ticks = %v, want a crawl and not a launch", c.Speed)
	}
}

// The car slowing itself is not the player asking for anything, so coasting is
// unscaled: with nothing held the car rolls to a full stop and stays there,
// rather than creeping backwards past zero.
func TestMoveCoastsToAStop(t *testing.T) {
	c := newTestCar()

	for i := 0; i < 100; i++ {
		c.Move()
	}

	if c.Speed != 0 {
		t.Errorf("Speed after coasting = %v, want 0", c.Speed)
	}
}

// The car ends every tick aimed along the line from its rear axle to its pivot.
// That lag between where the nose points and where the car is travelling is the
// drift. Rotation is degrees clockwise from straight up the screen.
func TestMoveAimsCarFromRearAxleToPivot(t *testing.T) {
	for _, tc := range []struct {
		name        string
		axleToPivot Vector
		want        float64
	}{
		{name: "nose up", axleToPivot: Vector{X: 0, Y: -160}, want: 0},
		{name: "nose right", axleToPivot: Vector{X: 160, Y: 0}, want: 90},
		{name: "nose down", axleToPivot: Vector{X: 0, Y: 160}, want: 180},
		{name: "nose left", axleToPivot: Vector{X: -160, Y: 0}, want: -90},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestCar()
			c.Direction = Direction{} // hold the pivot still, so only the aim is under test
			c.RearPivotAbs = RearPivotAbs{X: c.Pivot.X - tc.axleToPivot.X, Y: c.Pivot.Y - tc.axleToPivot.Y}

			c.Move()

			if !sameAngle(c.Rotation, tc.want) {
				t.Errorf("Rotation = %v, want %v", c.Rotation, tc.want)
			}
		})
	}
}

// Direction carries both where the car is heading and how fast: Move() adds it
// to the pivot as-is, so Steer() has to scale the unit heading by the speed.
func TestUpdateDirectionScalesHeadingBySpeed(t *testing.T) {
	c := newTestCar()
	c.Speed = 3
	c.FrontPivot = FrontPivot{X: 100, Y: 100}
	c.DirectionPivot = DirectionPivot{X: 100, Y: 50} // 50px straight up

	c.UpdateDirection()

	if !closeTo(c.Direction.X, 0) || !closeTo(c.Direction.Y, -3) {
		t.Errorf("Direction = (%v, %v), want (0, -3)", c.Direction.X, c.Direction.Y)
	}
}

// A car standing exactly on its own direction pivot has no heading to compute.
// The division by a zero length used to make Direction NaN, and from there the
// pivot and the rotation too, with no way back.
func TestUpdateDirectionWithNoHeadingIsZeroNotNaN(t *testing.T) {
	c := newTestCar()
	c.FrontPivot = FrontPivot{X: 100, Y: 100}
	c.DirectionPivot = DirectionPivot{X: 100, Y: 100}

	c.UpdateDirection()
	c.Move()

	if math.IsNaN(c.Direction.X) || math.IsNaN(c.Direction.Y) {
		t.Fatalf("Direction = (%v, %v), want no NaN", c.Direction.X, c.Direction.Y)
	}
	if math.IsNaN(c.Pivot.X) || math.IsNaN(c.Pivot.Y) || math.IsNaN(c.Rotation) {
		t.Errorf("car left at pivot (%v, %v) rotation %v, want no NaN", c.Pivot.X, c.Pivot.Y, c.Rotation)
	}
}

// RearPivot is the rear axle's offset in the car's own frame, so the axle's
// place on screen has to follow the car's rotation about the pivot.
func TestUpdateRearPivotAbsTurnsWithTheCar(t *testing.T) {
	c := newTestCar()
	c.Rotation = 30
	c.UpdateRearPivotAbs()

	wantX, wantY := rotate(c.RearPivot.X, c.RearPivot.Y, c.Rotation)
	wantX, wantY = wantX+c.Pivot.X, wantY+c.Pivot.Y

	if !closeTo(c.RearPivotAbs.X, wantX) || !closeTo(c.RearPivotAbs.Y, wantY) {
		t.Errorf("rear axle at (%v, %v), want (%v, %v)", c.RearPivotAbs.X, c.RearPivotAbs.Y, wantX, wantY)
	}
}

// The draw path must be free of side effects, otherwise the simulation speed
// follows the display's refresh rate instead of Ebiten's fixed tick rate.
func TestDrawGeometryDoesNotMutateCar(t *testing.T) {
	c := newTestCar()
	c.Steering = -1
	c.Steer()

	before := *c
	c.bodyGeoM()
	for i := range c.Wheels {
		c.wheelGeoM(i)
	}

	if !reflect.DeepEqual(*c, before) {
		t.Errorf("computing the draw geometry mutated the car:\n got %+v\nwant %+v", *c, before)
	}
}

// The sprites are placed by a single GeoM each rather than by a stack of nested
// transforms, and Ebiten composes a GeoM's calls in the opposite order, so the
// placement is worth pinning down.
func TestBodyGeoMPlacesSpriteOnPivot(t *testing.T) {
	c := newTestCar()
	c.Rotation = 30

	// The sprite's (60, 30) point is the car's pivot, whatever the rotation.
	g := c.bodyGeoM()
	x, y := g.Apply(60, 30)

	if !closeTo(x, c.Pivot.X) || !closeTo(y, c.Pivot.Y) {
		t.Errorf("body sprite pivot at (%v, %v), want (%v, %v)", x, y, c.Pivot.X, c.Pivot.Y)
	}
}

func TestWheelGeoMPlacesWheels(t *testing.T) {
	c := newTestCar()
	c.Rotation = 30
	c.WheelAngle = 20

	for i := range c.Wheels {
		g := c.wheelGeoM(i)
		x, y := g.Apply(0, 0)
		wantX, wantY := wantWheelCorner(c, i)

		if !closeTo(x, wantX) || !closeTo(y, wantY) {
			t.Errorf("wheel %d corner at (%v, %v), want (%v, %v)", i, x, y, wantX, wantY)
		}
	}
}

// wantWheelCorner is the wheel rectangle's top-left corner, worked out as the
// chain of transforms the wheel sits under: offset within the wheel, steering
// (front wheels only), the wheel's place on the body, the body's rotation, and
// finally the pivot's position on screen.
func wantWheelCorner(c *Car, i int) (float64, float64) {
	x, y := -6.0, -15.0
	if i < 2 {
		x, y = rotate(x, y, c.WheelAngle)
	}
	x, y = x+c.Wheels[i].X, y+c.Wheels[i].Y
	x, y = rotate(x, y, c.Rotation)
	return x + c.Pivot.X, y + c.Pivot.Y
}

func rotate(x, y, degrees float64) (float64, float64) {
	s, cos := math.Sincos(degrees * math.Pi / 180)
	return x*cos - y*s, x*s + y*cos
}

func closeTo(got, want float64) bool {
	return math.Abs(got-want) < 1e-9
}

// sameAngle compares two headings in degrees, treating those a whole turn apart
// as equal: the car's rotation is an angle, not a winding count.
func sameAngle(got, want float64) bool {
	diff := math.Mod(got-want, 360)
	if diff > 180 {
		diff -= 360
	} else if diff < -180 {
		diff += 360
	}
	return math.Abs(diff) < 1e-9
}
