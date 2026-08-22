package scene

import (
	"github.com/cmajid/carpen/carpen"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// The results screen's choices. Another race is first: it is what a player who
// has just finished one is most likely to want.
const (
	resultsAgain = iota
	resultsMenu
)

// Results is where a finished race lands. It has nothing to report yet — what a
// race is worth is decided by the win condition, which comes later in the
// roadmap (#17) — so for now it says so plainly rather than showing an empty
// scoreboard.
type Results struct {
	in   Input
	list *menuList
	fade fade

	// level is the one just played, kept so that Race Again deals the same
	// puzzle out afresh rather than dropping the player somewhere else.
	level carpen.Level

	// width is the screen the manager last reported; the card is centred on it.
	width int
}

func newResults(in Input, level carpen.Level) *Results {
	r := &Results{
		in:    in,
		list:  newMenuList(0, 254, 228, "Race Again", "Main Menu"),
		level: level,
	}
	r.resize(viewport{width: DesignWidth, height: DesignHeight})
	return r
}

func (r *Results) resize(v viewport) {
	r.width = v.width
	r.list.centreOn(float64(v.width))
}

func (r *Results) Update() (Scene, error) {
	r.fade.update()

	switch r.list.update(r.in) {
	case resultsAgain:
		return newGameplay(r.in, r.level), nil
	case resultsMenu:
		return NewMenu(r.in, r.level), nil
	}

	if justPressed(r.in, actionCancel) {
		return NewMenu(r.in, r.level), nil
	}

	return nil, nil
}

func (r *Results) Draw(screen *ebiten.Image) {
	screen.Fill(colourInk)

	const panelWidth = 340
	centre := float64(r.width) / 2

	drawPanel(screen, centre-panelWidth/2, 130, panelWidth, 230)
	drawText(screen, "Results", fontHeading, centre, 178, colourText, text.AlignCenter, text.AlignCenter)
	drawText(screen, "Nothing to score yet.", fontBody, centre, 212, colourTextMuted, text.AlignCenter, text.AlignCenter)

	r.list.draw(screen)

	drawMenuPrompts(screen, r.in, "Menu")
	r.fade.draw(screen)
}
