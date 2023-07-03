package carpen

import (
	"image"
	"log"
	"math"

	"github.com/fogleman/gg"
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
	b.Image, _, err = ebitenutil.NewImageFromFile("bush-small.png")
	if err != nil {
		log.Fatal(err)
	}
}

func (bush *Bush) DrawBush() image.Image {
	dc := gg.NewContext(640, 480)
	dc.Translate(bush.Direction.X, bush.Direction.Y)
	dc.Rotate(bush.Rotation * math.Pi / 180)
	dc.DrawImage(bush.Image, 0, 0)
	dc.Fill()
	return dc.Image()
}
