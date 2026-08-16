package carpen

import (
	"math"
	"testing"
)

func TestLength(t *testing.T) {
	for _, tc := range []struct {
		name string
		v    Vector
		want float64
	}{
		{name: "zero", v: Vector{X: 0, Y: 0}, want: 0},
		{name: "along x", v: Vector{X: 3, Y: 0}, want: 3},
		{name: "along y", v: Vector{X: 0, Y: -4}, want: 4},
		{name: "3-4-5", v: Vector{X: 3, Y: 4}, want: 5},
		{name: "negative components", v: Vector{X: -3, Y: -4}, want: 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.v.Length(); !closeTo(got, tc.want) {
				t.Errorf("Length() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNormalizeKeepsDirectionAtUnitLength(t *testing.T) {
	for _, tc := range []struct {
		name string
		v    Vector
		want Direction
	}{
		{name: "along x", v: Vector{X: 7, Y: 0}, want: Direction{X: 1, Y: 0}},
		{name: "up the screen", v: Vector{X: 0, Y: -50}, want: Direction{X: 0, Y: -1}},
		{name: "3-4-5", v: Vector{X: 3, Y: 4}, want: Direction{X: 0.6, Y: 0.8}},
		{name: "already unit", v: Vector{X: 0, Y: 1}, want: Direction{X: 0, Y: 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.v.Normalize()

			if !closeTo(got.X, tc.want.X) || !closeTo(got.Y, tc.want.Y) {
				t.Errorf("Normalize() = (%v, %v), want (%v, %v)", got.X, got.Y, tc.want.X, tc.want.Y)
			}
			if l := (&Vector{X: got.X, Y: got.Y}).Length(); !closeTo(l, 1) {
				t.Errorf("normalized length = %v, want 1", l)
			}
		})
	}
}

// A zero vector has no direction to preserve. Dividing by its length used to
// yield NaN, which then travelled into Car.Direction and Car.Pivot and put the
// car somewhere it could never be drawn back from.
func TestNormalizeZeroVectorIsZeroNotNaN(t *testing.T) {
	v := Vector{X: 0, Y: 0}

	got := v.Normalize()

	if math.IsNaN(got.X) || math.IsNaN(got.Y) {
		t.Fatalf("Normalize() of the zero vector = (%v, %v), want no NaN", got.X, got.Y)
	}
	if got != (Direction{}) {
		t.Errorf("Normalize() of the zero vector = %+v, want %+v", got, Direction{})
	}
}
