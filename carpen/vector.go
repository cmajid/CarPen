package carpen

import (
	"math"
)

type Vector struct {
	X, Y float64
}

func (v *Vector) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Normalize returns the unit-length direction the vector points in. A zero
// vector points nowhere, so it normalizes to zero rather than to the NaN that
// dividing by its length would give — a NaN here spreads to the car's
// direction, then to its position, and never leaves.
func (v *Vector) Normalize() Direction {
	length := v.Length()
	if length == 0 {
		return Direction{}
	}

	return Direction{
		X: v.X / length,
		Y: v.Y / length,
	}
}
