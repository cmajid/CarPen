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

	// width and height are the sprite's size, kept as numbers so the OBB can
	// be worked out without asking the image; the obb itself is built once and
	// re-placed on every OBB() call.
	width, height float64
	obb           *OBB
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

// OBB returns the bush's oriented bounding box. The sprite is drawn from its
// top-left corner at Direction and turned about that corner (see Draw), so the
// centre is half the sprite away from Direction, turned by Rotation.
func (bush *Bush) OBB() *OBB {
	sin, cos := math.Sincos(bush.Rotation * math.Pi / 180)
	dx := bush.width / 2
	dy := bush.height / 2
	centerX := dx*cos - dy*sin + bush.Direction.X
	centerY := dx*sin + dy*cos + bush.Direction.Y

	if bush.obb == nil {
		bush.obb = NewOBB(centerX, centerY, bush.width, bush.height, bush.Rotation)
	}
	bush.obb.SetTransform(centerX, centerY, bush.Rotation)
	return bush.obb
}

// Draw blits the bush onto screen. Direction is the bush's position on screen
// and is applied here, so this is the only place the bush is positioned.
func (bush *Bush) Draw(screen *ebiten.Image) {
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Rotate(bush.Rotation * math.Pi / 180)
	opt.GeoM.Translate(bush.Direction.X, bush.Direction.Y)
	screen.DrawImage(bush.Image, opt)
}
