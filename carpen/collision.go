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
	return NewBeveledOBB(centerX, centerY, width, height, 0, rotation)
}

// NewBeveledOBB builds a box with its corners cut off at 45 degrees, bevel
// pixels in from each side — the octagon a rounded rectangle flattens to. A
// hard corner catches on things the drawn shape visibly clears, so the car
// carries a bevel; a bevel of zero is a plain box. SAT does not care: an
// octagon is as convex as a rectangle, just with more edges to fetch axes
// from.
func NewBeveledOBB(centerX, centerY, width, height, bevel, rotation float64) *OBB {
	halfW, halfH := width/2, height/2
	bevel = math.Min(bevel, math.Min(halfW, halfH))

	points := []float64{
		-halfW, -halfH,
		halfW, -halfH,
		halfW, halfH,
		-halfW, halfH,
	}
	if bevel > 0 {
		points = []float64{
			-halfW + bevel, -halfH,
			halfW - bevel, -halfH,
			halfW, -halfH + bevel,
			halfW, halfH - bevel,
			halfW - bevel, halfH,
			-halfW + bevel, halfH,
			-halfW, halfH - bevel,
			-halfW, -halfH + bevel,
		}
	}

	obb := &OBB{shape: resolv.NewConvexPolygon(centerX, centerY, points)}
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

// Outline returns the shape's corners on the lot, in order around it — four
// for a plain box, eight for a beveled one — for the debug overlay to draw
// its outline from.
func (o *OBB) Outline() []Vector {
	transformed := o.shape.Transformed()
	outline := make([]Vector, len(transformed))
	for i, point := range transformed {
		outline[i] = Vector{X: point.X, Y: point.Y}
	}
	return outline
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

// Circle is a round collider, the shape the bushes actually are: their box's
// empty corners kept reporting crashes the player could see were not
// happening. It keeps its own centre and radius; the SAT arithmetic against a
// polygon borrows resolv's projections the same way OBB does.
type Circle struct {
	center Vector
	radius float64
}

// NewCircle builds a circle of the given radius around a centre. Like an OBB,
// it is built once and re-placed with SetPosition as its owner moves.
func NewCircle(centerX, centerY, radius float64) *Circle {
	return &Circle{center: Vector{X: centerX, Y: centerY}, radius: radius}
}

// SetPosition re-places the circle's centre on the lot.
func (c *Circle) SetPosition(centerX, centerY float64) {
	c.center = Vector{X: centerX, Y: centerY}
}

// Center returns the circle's centre on the lot.
func (c *Circle) Center() Vector { return c.center }

// Radius returns the circle's radius.
func (c *Circle) Radius() float64 { return c.radius }

// IntersectsOBB reports whether the circle overlaps the box, by the same
// separating axis theorem as Intersects. A circle has no edges of its own to
// take normals from; the one direction it can be separated along that the
// box's edge normals miss is the line from its centre to the box's nearest
// corner, so that axis joins the box's. Overlap is strict here too: resting
// against a face is not a hit.
func (c *Circle) IntersectsOBB(o *OBB) bool {
	axes := o.shape.SATAxes()

	nearest, nearestDistance := resolv.Vector{}, math.Inf(1)
	for _, vertex := range o.shape.Transformed() {
		dx, dy := vertex.X-c.center.X, vertex.Y-c.center.Y
		if distance := dx*dx + dy*dy; distance < nearestDistance {
			nearest, nearestDistance = resolv.Vector{X: dx, Y: dy}, distance
		}
	}
	// A centre sitting exactly on the corner has no direction to be pushed
	// away along; it is inside by any measure, and the axis would be zero.
	if nearestDistance > 0 {
		axes = append(axes, nearest)
	}

	for _, axis := range axes {
		unit := axis.Unit()
		centre := unit.Dot(resolv.Vector{X: c.center.X, Y: c.center.Y})
		projection := resolv.Projection{Min: centre - c.radius, Max: centre + c.radius}
		if !o.shape.Project(unit).IsOverlapping(projection) {
			return false
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
