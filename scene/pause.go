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

	// width is the screen the manager last reported. The overlay is a card in
	// the middle of the screen, so all of it is placed against this.
	width int
}

func newPause(g *Gameplay) *Pause {
	p := &Pause{
		in:     g.in,
		list:   newMenuList(0, 216, 228, "Resume", "Restart", "Quit to Menu"),
		resume: g,
	}
	p.resize(g.view)

	g.releaseControls()

	return p
}

// resize centres the overlay, and passes the size on to the race underneath —
// which is still being drawn behind it, and would otherwise sit off to one side
// of its own pause menu if the screen changed while the game was paused.
func (p *Pause) resize(v viewport) {
	p.width = v.width
	p.list.centreOn(float64(v.width))
	p.resume.resize(v)
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

	const panelWidth = 300
	centre := float64(p.width) / 2

	drawPanel(screen, centre-panelWidth/2, 130, panelWidth, 232)
	drawText(screen, "Paused", fontHeading, centre, 176, colourText, text.AlignCenter, text.AlignCenter)

	p.list.draw(screen)

	drawMenuPrompts(screen, p.in, "Resume")
}
