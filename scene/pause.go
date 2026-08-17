package scene

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// The pause menu's choices. Resuming is first, so the pause can be undone with
// the same hands that made it.
const (
	pauseResume = iota
	pauseRestart
	pauseQuit
)

// Pause holds the race behind an overlay. It keeps the Gameplay scene itself
// rather than a copy of the world, so resuming carries on from exactly where the
// player stopped; the cars stand still simply because nothing ticks them while
// this is the scene the manager is running.
type Pause struct {
	in     Input
	list   *menuList
	resume *Gameplay
}

func newPause(g *Gameplay) *Pause {
	g.releaseControls()

	return &Pause{
		in:     g.in,
		list:   newMenuList(206, 216, 228, "Resume", "Restart", "Quit to Menu"),
		resume: g,
	}
}

func (p *Pause) Update() (Scene, error) {
	switch p.list.update(p.in) {
	case pauseResume:
		return p.resume, nil
	case pauseRestart:
		return newGameplay(p.in, p.resume.level), nil
	case pauseQuit:
		return NewMenu(p.in, p.resume.level), nil
	}

	// Backing out goes back a step everywhere in the game, and the step back
	// from a paused race is the race.
	if justPressed(p.in, actionCancel) {
		return p.resume, nil
	}

	return nil, nil
}

func (p *Pause) Draw(screen *ebiten.Image) {
	p.resume.Draw(screen)

	bounds := screen.Bounds()
	fillRect(screen, 0, 0, float64(bounds.Dx()), float64(bounds.Dy()), colourDim)

	drawPanel(screen, 170, 130, 300, 232)
	drawText(screen, "Paused", fontHeading, 320, 176, colourText, text.AlignCenter, text.AlignCenter)

	p.list.draw(screen)

	drawPrompts(screen,
		promptFor(p.in, "Up / Down", "Stick", "Move"),
		promptFor(p.in, "Enter", "A", "Select"),
		promptFor(p.in, "Esc", "B", "Resume"),
	)
}
