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
func (norm *Vector) Normalize() Direction {

	return Direction{
		X: norm.X / norm.Length(),
		Y: norm.Y / norm.Length(),
	}
}
