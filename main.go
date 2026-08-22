package main

import (
	"log"

	"github.com/cmajid/carpen/carpen"
	"github.com/cmajid/carpen/scene"
	"github.com/hajimehoshi/ebiten/v2"
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
	game := scene.NewManager(scene.NewMenu(scene.NewDevices(), levels[0]))

	// The window opens at the design size scaled up a whole number of times —
	// as many as the monitor can hold — so the game arrives at a comfortable
	// size on any desktop and its pixels stay crisp. It can be dragged to any
	// shape from there (Layout absorbs whatever shape it becomes), and F11
	// trades the window for the whole screen. On a phone or in a browser there
	// is no window to size, and these calls quietly do nothing.
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	monitorWidth, monitorHeight := ebiten.Monitor().Size()
	scale := min(monitorWidth/scene.DesignWidth, monitorHeight/scene.DesignHeight)
	if scale < 1 {
		scale = 1
	}
	ebiten.SetWindowSize(scene.DesignWidth*scale, scene.DesignHeight*scale)
	ebiten.SetWindowTitle("Car Pen")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
