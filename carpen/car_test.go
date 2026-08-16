package carpen

import (
	"math"
	"reflect"
	"testing"

	"github.com/fogleman/gg"
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
		Width:             100,
		Height:            200,
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
	c.RearPivotAbs = RearPivotAbs{
		X: 160*math.Cos((c.Rotation+90)*math.Pi/180) + c.Pivot.X,
		Y: 160*math.Sin((c.Rotation+90)*math.Pi/180) + c.Pivot.Y,
	}
	v := Vector{X: c.DirectionPivot.X - c.FrontPivot.X, Y: c.DirectionPivot.Y - c.FrontPivot.Y}
	c.Direction = v.Normalize()
	return c
}

func TestUpdateStepsWheelAngleOncePerTick(t *testing.T) {
	c := newTestCar()
	c.RotateLeft = true

	if err := c.Update(); err != nil {
		t.Fatalf("Update() returned %v", err)
	}

	if got, want := c.WheelAngle, -c.WheelRotationStep; got != want {
		t.Errorf("WheelAngle after one tick = %v, want %v", got, want)
	}
}

func TestSteerClampsToMaxAngle(t *testing.T) {
	for _, tc := range []struct {
		name  string
		left  bool
		right bool
		want  float64
	}{
		{name: "left", left: true, want: -45},
		{name: "right", right: true, want: 45},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestCar()
			c.RotateLeft, c.RotateRight = tc.left, tc.right

			for i := 0; i < 200; i++ {
				c.Steer()
			}

			if c.WheelAngle != tc.want {
				t.Errorf("WheelAngle = %v, want %v", c.WheelAngle, tc.want)
			}
		})
	}
}

func TestMoveAdvancesPivot(t *testing.T) {
	c := newTestCar()
	before := c.Pivot

	if err := c.Move(); err != nil {
		t.Fatalf("Move() returned %v", err)
	}

	if c.Pivot == before {
		t.Errorf("Pivot did not move: still %v", c.Pivot)
	}
}

// The draw path must be free of side effects, otherwise the simulation speed
// follows the display's refresh rate instead of Ebiten's fixed tick rate.
func TestDrawWheelsDoesNotMutateCar(t *testing.T) {
	c := newTestCar()
	c.RotateLeft = true
	c.Steer()

	before := *c
	c.DrawWheels(gg.NewContext(640, 480))

	if !reflect.DeepEqual(*c, before) {
		t.Errorf("DrawWheels mutated the car:\n got %+v\nwant %+v", *c, before)
	}
}
