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
}

func newResults(in Input, level carpen.Level) *Results {
	return &Results{
		in:    in,
		list:  newMenuList(206, 254, 228, "Race Again", "Main Menu"),
		level: level,
	}
}

func (r *Results) Update() (Scene, error) {
	r.fade.update()

	switch r.list.update(r.in) {
	case resultsAgain:
		return newGameplay(r.in, r.level), nil
	case resultsMenu:
		return NewMenu(r.in, r.level), nil
	}

	if r.in.IsKeyJustPressed(ebiten.KeyEscape) {
		return NewMenu(r.in, r.level), nil
	}

	return nil, nil
}

func (r *Results) Draw(screen *ebiten.Image) {
	screen.Fill(colourInk)

	drawPanel(screen, 150, 130, 340, 230)
	drawText(screen, "Results", fontHeading, 320, 178, colourText, text.AlignCenter, text.AlignCenter)
	drawText(screen, "Nothing to score yet.", fontBody, 320, 212, colourTextMuted, text.AlignCenter, text.AlignCenter)

	r.list.draw(screen)

	drawPrompts(screen,
		prompt{key: "Up / Down", does: "Move"},
		prompt{key: "Enter", does: "Select"},
		prompt{key: "Esc", does: "Menu"},
	)
	r.fade.draw(screen)
}
