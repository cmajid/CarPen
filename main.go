package main

import (
	"log"

	"github.com/cmajid/carpen/carpen"
	"github.com/cmajid/carpen/scene"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

func main() {
	// The levels are compiled into the binary, so one that will not load is a
	// broken build rather than a missing file, and there is nothing to play.
	levels, err := carpen.Levels()
	if err != nil {
		log.Fatal(err)
	}

	// The menu opens on the first level. Working through the rest of them is
	// the level progression, which comes later in the roadmap (#17).
	game := scene.NewManager(screenWidth, screenHeight, scene.NewMenu(scene.NewDevices(), levels[0]))

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Car Pen")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
