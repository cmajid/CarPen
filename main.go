package main

import (
	"log"

	"github.com/cmajid/carpen/scene"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

func main() {
	game := scene.NewManager(screenWidth, screenHeight, scene.NewMenu(scene.Keyboard{}))

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Car Pen")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
