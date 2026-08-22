// Package mobile is the game as a library, which is the shape ebitenmobile
// binds: an iOS framework or an Android archive is a library the platform's own
// app shell calls into, so there is no main to run and no window to open. What
// main.go does at start-up is done here in init instead.
//
//	ebitenmobile bind -target ios -o dist/Carpen.xcframework ./mobile
//	ebitenmobile bind -target android -o dist/carpen.aar ./mobile
package mobile

import (
	"github.com/cmajid/carpen/carpen"
	"github.com/cmajid/carpen/scene"
	"github.com/hajimehoshi/ebiten/v2/mobile"
)

func init() {
	levels, err := carpen.Levels()
	if err != nil {
		// The levels are compiled in, so one that will not load is a broken
		// build. There is no console to log it to here and nothing to play
		// without them, so this crashes the way main.go exits: loudly, in the
		// build that is wrong, rather than quietly in the player's hands.
		panic(err)
	}

	mobile.SetGame(scene.NewManager(scene.NewMenu(scene.NewDevices(), levels[0])))
}

// Dummy is here because gomobile will not bind a package that exports nothing,
// and everything this package does it does in init.
func Dummy() {}
