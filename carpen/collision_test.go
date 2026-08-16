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

// TestOBBCorners pins the whole transform down at once: a 4x2 box at (10, 20)
// turned 90 degrees clockwise stands upright, so its corners are the centre
// plus (±1, ±2), starting from where the unturned top-left corner went.
func TestOBBCorners(t *testing.T) {
	obb := NewOBB(10, 20, 4, 2, 90)

	want := [4]Vector{
		{X: 11, Y: 18},
		{X: 11, Y: 22},
		{X: 9, Y: 22},
		{X: 9, Y: 18},
	}

	corners := obb.Corners()
	for i := range want {
		if math.Abs(corners[i].X-want[i].X) > 1e-9 || math.Abs(corners[i].Y-want[i].Y) > 1e-9 {
			t.Errorf("corner %d = (%g, %g), want (%g, %g)", i, corners[i].X, corners[i].Y, want[i].X, want[i].Y)
		}
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

	// Turned half a circle, the same offset points the other way.
	car.Rotation = 180
	centre = car.OBB().Center()
	wantX, wantY = 200.0-(62-60), 100.0-(109.5-30)
	if math.Abs(centre.X-wantX) > 1e-9 || math.Abs(centre.Y-wantY) > 1e-9 {
		t.Errorf("turned car OBB centre = (%g, %g), want (%g, %g)", centre.X, centre.Y, wantX, wantY)
	}
}

// TestBushOBB pins the bush's box to how it is drawn: the sprite's top-left
// corner sits at Direction, so the box is centred half a sprite further in.
func TestBushOBB(t *testing.T) {
	bush := Bush{
		Direction: Direction{X: 40, Y: 60},
		width:     109,
		height:    108,
	}

	centre := bush.OBB().Center()
	if math.Abs(centre.X-(40+54.5)) > 1e-9 || math.Abs(centre.Y-(60+54)) > 1e-9 {
		t.Errorf("bush OBB centre = (%g, %g), want (%g, %g)", centre.X, centre.Y, 40+54.5, 60.0+54)
	}
}
