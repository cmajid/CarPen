package carpen

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Bush struct {
	Rotation float64

	Direction Direction
	Image     *ebiten.Image

	// width and height are the sprite's size, kept as numbers so the collider
	// can be worked out without asking the image; the circle itself is built
	// once and re-placed on every Collider() call.
	width, height float64
	circle        *Circle
}

func (b *Bush) Init() {
	var err error
	b.Image, _, err = ebitenutil.NewImageFromFileSystem(assets, "assets/bush-small.png")
	if err != nil {
		log.Fatal(err)
	}
	b.width = float64(b.Image.Bounds().Dx())
	b.height = float64(b.Image.Bounds().Dy())
}

// bushInset draws the bush's circle in a little from the sprite's edge, the
// same forgiveness the car's shape carries: brushing the outermost leaves is
// not a crash. Raise it for a more forgiving game.
const bushInset = 4.0

// Collider returns the bush's collision circle. The bush is round, and a box
// around it kept crashing cars into its empty corners, so it collides as the
// circle it looks like. The sprite is drawn from its top-left corner at
// Direction and turned about that corner (see Draw), so the centre is half
// the sprite away from Direction, turned by Rotation — which is all the
// rotation there is to a circle.
func (bush *Bush) Collider() *Circle {
	sin, cos := math.Sincos(bush.Rotation * math.Pi / 180)
	dx := bush.width / 2
	dy := bush.height / 2
	centerX := dx*cos - dy*sin + bush.Direction.X
	centerY := dx*sin + dy*cos + bush.Direction.Y

	if bush.circle == nil {
		bush.circle = NewCircle(centerX, centerY, math.Min(bush.width, bush.height)/2-bushInset)
	}
	bush.circle.SetPosition(centerX, centerY)
	return bush.circle
}

// Draw blits the bush onto screen. Direction is the bush's position in the
// level and is applied here, so this is the only place the bush is positioned;
// origin is where the level's (0, 0) falls on the destination, as for a car.
func (bush *Bush) Draw(screen *ebiten.Image, origin Vector) {
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Rotate(bush.Rotation * math.Pi / 180)
	opt.GeoM.Translate(bush.Direction.X+origin.X, bush.Direction.Y+origin.Y)
	screen.DrawImage(bush.Image, opt)
}
