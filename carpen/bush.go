package carpen

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Bush struct {
	Rotation float64
	Width,
	Height int

	Direction Direction
	Image     *ebiten.Image
}

func (b *Bush) Init() {
	var err error
	b.Image, _, err = ebitenutil.NewImageFromFileSystem(assets, "assets/bush-small.png")
	if err != nil {
		log.Fatal(err)
	}
}

// Draw blits the bush onto screen. Direction is the bush's position on screen
// and is applied here, so this is the only place the bush is positioned.
func (bush *Bush) Draw(screen *ebiten.Image) {
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Rotate(bush.Rotation * math.Pi / 180)
	opt.GeoM.Translate(bush.Direction.X, bush.Direction.Y)
	screen.DrawImage(bush.Image, opt)
}
