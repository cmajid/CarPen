package cp_vector

import (
	"math"

	"github.com/cmajid/carpen/cp_pivot"
)

type Vector struct {
	X, Y float64
}

func (v *Vector) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}
func (norm *Vector) Normalize() cp_pivot.Direction {

	return cp_pivot.Direction{
		X: norm.X / norm.Length(),
		Y: norm.Y / norm.Length(),
	}
}
