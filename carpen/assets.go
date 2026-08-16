package carpen

import (
	"embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// assets holds the game's images. They are compiled into the binary so the game
// runs from any working directory instead of only from the repository root.
//
//go:embed assets/*.png
var assets embed.FS

// CarImage loads the sprite of a car in the given colour. It is how a car gets
// its own image, and how the menu borrows one to decorate itself with.
func CarImage(colour string) *ebiten.Image {
	image, _, err := ebitenutil.NewImageFromFileSystem(assets, "assets/car-"+colour+".png")
	if err != nil {
		log.Fatal(err)
	}
	return image
}
