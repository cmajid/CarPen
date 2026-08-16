package carpen

import (
	"math"

	"github.com/solarlune/resolv"
)

// OBB is an oriented bounding box on the lot: a rectangle described by its
// centre, its size, and a rotation in degrees clockwise — the convention
// everything on the lot is drawn with. It wraps a resolv ConvexPolygon so the
// separating-axis arithmetic (edge normals, projections) is the library's.
//
// This file is the only place resolv is imported. The library speaks radians
// counter-clockwise and misreports containment (see Intersects); both are
// translated away here so the rest of the game never has to know.
type OBB struct {
	shape *resolv.ConvexPolygon
}

// NewOBB builds a box of the given width and height around a centre, turned
// rotation degrees clockwise. An entity keeps the one OBB it is given and
// re-places it with SetTransform each time it moves, rather than building a
// fresh box every tick.
func NewOBB(centerX, centerY, width, height, rotation float64) *OBB {
	obb := &OBB{shape: resolv.NewRectangle(centerX, centerY, width, height)}
	obb.SetTransform(centerX, centerY, rotation)
	return obb
}

// SetTransform re-places the box: centre on the lot, rotation in degrees
// clockwise. Resolv turns shapes counter-clockwise, the math.Atan2 way, while
// the lot turns clockwise because Y grows downward — so the angle changes sign
// here and nowhere else.
func (o *OBB) SetTransform(centerX, centerY, rotation float64) {
	o.shape.SetPosition(centerX, centerY)
	o.shape.SetRotation(-rotation * math.Pi / 180)
}

// Center returns the box's centre on the lot.
func (o *OBB) Center() Vector {
	position := o.shape.Position()
	return Vector{X: position.X, Y: position.Y}
}

// Corners returns the box's four corners on the lot, in order around the box,
// for the debug overlay to draw its outline from.
func (o *OBB) Corners() [4]Vector {
	var corners [4]Vector
	for i, point := range o.shape.Transformed() {
		corners[i] = Vector{X: point.X, Y: point.Y}
	}
	return corners
}

// Intersects reports whether two boxes overlap, by the separating axis
// theorem: the boxes miss each other exactly when, on some edge normal of one
// of them, their projections come apart. The overlap test is strict, so two
// boxes just touching edge-to-edge are not yet a hit.
//
// The test is projections all the way down rather than resolv's own
// IsIntersecting because that looks for crossing edges, and a box wholly
// inside another has none to find.
func Intersects(a, b *OBB) bool {
	for _, shape := range [2]*resolv.ConvexPolygon{a.shape, b.shape} {
		for _, axis := range shape.SATAxes() {
			if !a.shape.Project(axis).IsOverlapping(b.shape.Project(axis)) {
				return false
			}
		}
	}
	return true
}

// Obstruction is the kind of thing a car can run into. The values read as
// words so a message can name them directly.
type Obstruction string

const (
	ObstructionWall Obstruction = "wall"
	ObstructionBush Obstruction = "bush"
	ObstructionCar  Obstruction = "car"
)

// CollisionEvent says the active car ran into something. Detection stops at
// raising the event; what a crash means — failing the attempt, ending the
// level — is decided by whoever receives it, so the rules can change without
// this file knowing.
type CollisionEvent struct {
	Obstruction Obstruction
}
