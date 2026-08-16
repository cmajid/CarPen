package carpen

import (
	"math"
	"testing"
)

func TestIntersects(t *testing.T) {
	cases := []struct {
		name string
		a, b *OBB
		want bool
	}{
		{
			name: "separated",
			a:    NewOBB(0, 0, 10, 10, 0),
			b:    NewOBB(30, 0, 10, 10, 0),
			want: false,
		},
		{
			name: "overlapping",
			a:    NewOBB(0, 0, 10, 10, 0),
			b:    NewOBB(8, 0, 10, 10, 0),
			want: true,
		},
		{
			// A 30-wide box reaches 15 from its centre, short of a box whose
			// near edge is 20 away — until it turns 45 degrees and its corner
			// reaches out to just past 21.
			name: "rotation makes the hit",
			a:    NewOBB(0, 0, 30, 30, 45),
			b:    NewOBB(25, 0, 10, 10, 0),
			want: true,
		},
		{
			name: "same boxes unrotated miss",
			a:    NewOBB(0, 0, 30, 30, 0),
			b:    NewOBB(25, 0, 10, 10, 0),
			want: false,
		},
		{
			// Rotation is degrees clockwise on a screen whose Y grows
			// downward, so a long box turned 45 degrees points down-right:
			// it reaches a box diagonally below it and not the mirror-image
			// one above. These two cases pin the direction down — with the
			// sign turned the wrong way they swap answers.
			name: "clockwise reaches down-right",
			a:    NewOBB(0, 0, 100, 20, 45),
			b:    NewOBB(30, 30, 10, 10, 0),
			want: true,
		},
		{
			name: "clockwise misses up-right",
			a:    NewOBB(0, 0, 100, 20, 45),
			b:    NewOBB(30, -30, 10, 10, 0),
			want: false,
		},
		{
			// Touching is not yet hitting: the boxes share an edge but no
			// area, and a car grazing along a wall is not a crash.
			name: "edge touching",
			a:    NewOBB(0, 0, 10, 10, 0),
			b:    NewOBB(10, 0, 10, 10, 0),
			want: false,
		},
		{
			// A box wholly inside another has no crossing edges, which is
			// exactly the case resolv's own IsIntersecting misses — this is
			// the regression test for doing SAT projections instead.
			name: "contained",
			a:    NewOBB(0, 0, 100, 100, 0),
			b:    NewOBB(5, 5, 20, 20, 30),
			want: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Intersects(c.a, c.b); got != c.want {
				t.Errorf("Intersects(a, b) = %t, want %t", got, c.want)
			}
			// Overlap has no direction; the answer cannot depend on the order
			// the boxes are asked in.
			if got := Intersects(c.b, c.a); got != c.want {
				t.Errorf("Intersects(b, a) = %t, want %t", got, c.want)
			}
		})
	}
}

// TestOBBOutline pins the whole transform down at once: a 4x2 box at (10, 20)
// turned 90 degrees clockwise stands upright, so its corners are the centre
// plus (±1, ±2), starting from where the unturned top-left corner went.
func TestOBBOutline(t *testing.T) {
	obb := NewOBB(10, 20, 4, 2, 90)

	want := []Vector{
		{X: 11, Y: 18},
		{X: 11, Y: 22},
		{X: 9, Y: 22},
		{X: 9, Y: 18},
	}

	outline := obb.Outline()
	if len(outline) != len(want) {
		t.Fatalf("outline has %d corners, want %d", len(outline), len(want))
	}
	for i := range want {
		if math.Abs(outline[i].X-want[i].X) > 1e-9 || math.Abs(outline[i].Y-want[i].Y) > 1e-9 {
			t.Errorf("corner %d = (%g, %g), want (%g, %g)", i, outline[i].X, outline[i].Y, want[i].X, want[i].Y)
		}
	}
}

// TestBeveledOBBMissesAtTheCorner probes a box only just overlapping a plain
// rectangle's corner region: the bevel cuts that corner off, so the beveled
// shape — same width, same height — clears what the sharp one catches.
func TestBeveledOBBMissesAtTheCorner(t *testing.T) {
	probe := NewOBB(10.5, 4.5, 2, 2, 0)

	if !Intersects(NewOBB(0, 0, 20, 10, 0), probe) {
		t.Fatal("the sharp-cornered box misses the probe; the test is set up wrong")
	}
	if Intersects(NewBeveledOBB(0, 0, 20, 10, 3, 0), probe) {
		t.Error("the beveled box still catches the probe its cut corner should clear")
	}
}

func TestBeveledOBBOutline(t *testing.T) {
	outline := NewBeveledOBB(0, 0, 20, 10, 3, 0).Outline()

	want := []Vector{
		{X: -7, Y: -5},
		{X: 7, Y: -5},
		{X: 10, Y: -2},
		{X: 10, Y: 2},
		{X: 7, Y: 5},
		{X: -7, Y: 5},
		{X: -10, Y: 2},
		{X: -10, Y: -2},
	}
	if len(outline) != len(want) {
		t.Fatalf("outline has %d corners, want %d", len(outline), len(want))
	}
	for i := range want {
		if math.Abs(outline[i].X-want[i].X) > 1e-9 || math.Abs(outline[i].Y-want[i].Y) > 1e-9 {
			t.Errorf("corner %d = (%g, %g), want (%g, %g)", i, outline[i].X, outline[i].Y, want[i].X, want[i].Y)
		}
	}
}

func TestCircleIntersectsOBB(t *testing.T) {
	box := NewOBB(0, 0, 10, 10, 0)

	cases := []struct {
		name   string
		circle *Circle
		box    *OBB
		want   bool
	}{
		{"separated", NewCircle(30, 0, 5), box, false},
		{"overlapping a face", NewCircle(9, 0, 5), box, true},
		{
			// The case the circle collider exists for: near the box's corner,
			// every face axis overlaps, and only the centre-to-corner axis
			// tells that the round shape is still 5.66 away from it.
			name:   "clears the corner a square would catch",
			circle: NewCircle(9, 9, 5),
			box:    box,
			want:   false,
		},
		{"reaches the corner", NewCircle(8, 8, 5), box, true},
		{"contained", NewCircle(0, 0, 3), NewOBB(0, 0, 100, 100, 0), true},
		{"centre on the corner", NewCircle(5, 5, 1), box, true},
		{
			// A turned box turns its corners with it: at 45 degrees the box's
			// corner points along the X axis and reaches out to √50 ≈ 7.07,
			// catching a circle its unturned self stays clear of.
			name:   "rotated box reaches further",
			circle: NewCircle(11, 0, 4),
			box:    NewOBB(0, 0, 10, 10, 45),
			want:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.circle.IntersectsOBB(c.box); got != c.want {
				t.Errorf("IntersectsOBB = %t, want %t", got, c.want)
			}
		})
	}
}

func TestOBBSetTransformMoves(t *testing.T) {
	obb := NewOBB(0, 0, 10, 10, 0)
	obb.SetTransform(50, 60, 45)

	if centre := obb.Center(); centre.X != 50 || centre.Y != 60 {
		t.Errorf("centre after SetTransform = (%g, %g), want (50, 60)", centre.X, centre.Y)
	}
	if !Intersects(obb, NewOBB(50, 60, 4, 4, 0)) {
		t.Error("moved box does not cover its new centre")
	}
	if Intersects(obb, NewOBB(0, 0, 4, 4, 0)) {
		t.Error("moved box still covers where it came from")
	}
}

// TestCarOBB drives the box from the car's own numbers: the body sprite hangs
// (-60, -30) from the pivot, so an unturned car's box is centred at the pivot
// plus half the sprite less that overhang.
func TestCarOBB(t *testing.T) {
	car := Car{
		Pivot:      Pivot{X: 200, Y: 100},
		Rotation:   0,
		bodyWidth:  124,
		bodyHeight: 219,
	}

	centre := car.OBB().Center()
	wantX, wantY := 200.0+(62-60), 100.0+(109.5-30)
	if math.Abs(centre.X-wantX) > 1e-9 || math.Abs(centre.Y-wantY) > 1e-9 {
		t.Errorf("car OBB centre = (%g, %g), want (%g, %g)", centre.X, centre.Y, wantX, wantY)
	}
	if corners := len(car.OBB().Outline()); corners != 8 {
		t.Errorf("car outline has %d corners, want the beveled shape's 8", corners)
	}

	// Turned half a circle, the same offset points the other way.
	car.Rotation = 180
	centre = car.OBB().Center()
	wantX, wantY = 200.0-(62-60), 100.0-(109.5-30)
	if math.Abs(centre.X-wantX) > 1e-9 || math.Abs(centre.Y-wantY) > 1e-9 {
		t.Errorf("turned car OBB centre = (%g, %g), want (%g, %g)", centre.X, centre.Y, wantX, wantY)
	}
}

// TestBushCollider pins the bush's circle to how it is drawn: the sprite's
// top-left corner sits at Direction, so the circle is centred half a sprite
// further in, and its radius is the sprite's shorter half less the inset.
func TestBushCollider(t *testing.T) {
	bush := Bush{
		Direction: Direction{X: 40, Y: 60},
		width:     109,
		height:    108,
	}

	circle := bush.Collider()
	centre := circle.Center()
	if math.Abs(centre.X-(40+54.5)) > 1e-9 || math.Abs(centre.Y-(60+54)) > 1e-9 {
		t.Errorf("bush circle centre = (%g, %g), want (%g, %g)", centre.X, centre.Y, 40+54.5, 60.0+54)
	}
	if want := 108.0/2 - bushInset; circle.Radius() != want {
		t.Errorf("bush circle radius = %g, want %g", circle.Radius(), want)
	}
}
