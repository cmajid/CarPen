package scene

import (
	"bytes"
	"image/color"
	"log"

	"github.com/cmajid/carpen/carpen"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

// The screens share one small design system: a fixed palette, a handful of type
// sizes, and one way of showing which thing is in focus. Every screen is drawn
// from these, which is what makes the menu, the pause overlay and the results
// feel like parts of the same game rather than three separate ones.

// The palette. Amber is the game's own colour, taken from the yellow car, and is
// spent only on the thing the player is about to act on.
var (
	colourInk       = color.RGBA{R: 0x0f, G: 0x14, B: 0x1b, A: 0xff} // the ground every menu screen is drawn on
	colourPanel     = color.RGBA{R: 0x19, G: 0x22, B: 0x2d, A: 0xff} // a card lifted off that ground
	colourText      = color.RGBA{R: 0xee, G: 0xf2, B: 0xf6, A: 0xff}
	colourTextMuted = color.RGBA{R: 0x93, G: 0xa1, B: 0xb1, A: 0xff}
	colourAccent    = color.RGBA{R: 0xff, G: 0xb0, B: 0x20, A: 0xff}
	colourAccentInk = color.RGBA{R: 0x14, G: 0x18, B: 0x1f, A: 0xff} // text written on top of the accent
	colourLine      = color.RGBA{R: 0x2a, G: 0x35, B: 0x42, A: 0xff}
	colourBay       = color.RGBA{R: 0x8a, G: 0x96, B: 0xa4, A: 0xff} // the bay painted on the lot, which is a pale ground
	colourDanger    = color.RGBA{R: 0xe5, G: 0x4b, B: 0x4b, A: 0xff} // a box the moment it is overlapping, on the F3 overlay

	// Both of these are drawn over something else, so they are written with
	// their alpha already folded in: Go's RGBA is alpha-premultiplied.
	colourDim = color.RGBA{R: 0x07, G: 0x0a, B: 0x0e, A: 0xcc} // the frozen race behind the pause overlay
	colourHUD = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xa8} // the strip the driving keys are written on
)

// The type scale. Three sizes and two weights are enough to tell a title from a
// choice from a hint, and few enough to stay consistent by accident.
var (
	regular = mustFont(goregular.TTF)
	bold    = mustFont(gobold.TTF)

	fontTitle    = &text.GoTextFace{Source: bold, Size: 46}
	fontHeading  = &text.GoTextFace{Source: bold, Size: 26}
	fontItem     = &text.GoTextFace{Source: regular, Size: 20}
	fontItemHot  = &text.GoTextFace{Source: bold, Size: 20}
	fontBody     = &text.GoTextFace{Source: regular, Size: 15}
	fontPrompt   = &text.GoTextFace{Source: regular, Size: 13}
	fontPromptOn = &text.GoTextFace{Source: bold, Size: 13}
)

// mustFont reads one of the Go fonts. The bytes are compiled into the binary, so
// failing here means the binary itself is broken and there is nothing to fall
// back to.
func mustFont(ttf []byte) *text.GoTextFaceSource {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(ttf))
	if err != nil {
		log.Fatal(err)
	}
	return source
}

// drawText writes str at x, y. primary aligns it across the line — pass
// text.AlignCenter and x is the middle of the text rather than its left edge —
// and secondary aligns it down the line, so a row can be centred on its middle.
func drawText(dst *ebiten.Image, str string, face text.Face, x, y float64, colour color.Color, primary, secondary text.Align) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(colour)
	op.PrimaryAlign = primary
	op.SecondaryAlign = secondary
	text.Draw(dst, str, face, op)
}

func fillRect(dst *ebiten.Image, x, y, width, height float64, colour color.Color) {
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(width), float32(height), colour, false)
}

// strokeLine draws a line between two points. It is antialiased, because the
// lines it draws are the markings of a parking bay, which a level is free to
// turn to any angle.
func strokeLine(dst *ebiten.Image, from, to carpen.Vector, width float64, colour color.Color) {
	vector.StrokeLine(dst, float32(from.X), float32(from.Y), float32(to.X), float32(to.Y), float32(width), colour, true)
}

// drawPanel draws a card with a line of accent along its top edge, the shape the
// pause overlay and the results screen both present their choices in.
func drawPanel(dst *ebiten.Image, x, y, width, height float64) {
	fillRect(dst, x, y, width, height, colourPanel)
	fillRect(dst, x, y, width, 3, colourAccent)
}

// prompt is one thing the player can press and what it does.
type prompt struct {
	key, does string
}

// drawPrompts writes the prompts in a row along the bottom of the screen. They
// are the same shape and in the same place on every screen, so the keys are
// learned once rather than hunted for screen by screen.
func drawPrompts(dst *ebiten.Image, prompts ...prompt) {
	const (
		gapAfterKey  = 7  // between a key and what it does
		gapAfterPair = 26 // between one prompt and the next
		baseline     = 458
	)

	width := 0.0
	for i, p := range prompts {
		keyWidth, _ := text.Measure(p.key, fontPromptOn, 0)
		doesWidth, _ := text.Measure(p.does, fontPrompt, 0)
		width += keyWidth + gapAfterKey + doesWidth
		if i < len(prompts)-1 {
			width += gapAfterPair
		}
	}

	fillRect(dst, 0, baseline-16, float64(dst.Bounds().Dx()), 36, colourHUD)

	x := (float64(dst.Bounds().Dx()) - width) / 2
	for _, p := range prompts {
		drawText(dst, p.key, fontPromptOn, x, baseline, colourAccent, text.AlignStart, text.AlignCenter)
		keyWidth, _ := text.Measure(p.key, fontPromptOn, 0)
		x += keyWidth + gapAfterKey

		drawText(dst, p.does, fontPrompt, x, baseline, colourTextMuted, text.AlignStart, text.AlignCenter)
		doesWidth, _ := text.Measure(p.does, fontPrompt, 0)
		x += doesWidth + gapAfterPair
	}
}

// fade darkens the first few ticks of a scene, so screens change with a blink
// instead of a jump cut. It is short on purpose: long enough to read as a
// transition, too short to stand between the player and what they asked for.
type fade struct {
	tick int
}

const fadeTicks = 10

func (f *fade) update() {
	if f.tick < fadeTicks {
		f.tick++
	}
}

func (f *fade) draw(dst *ebiten.Image) {
	if f.tick >= fadeTicks {
		return
	}

	// Black is the same premultiplied or not, so only the alpha has to be worked
	// out: solid on the first tick, gone by the last.
	alpha := uint8(0xff * (1 - float64(f.tick)/fadeTicks))
	bounds := dst.Bounds()
	fillRect(dst, 0, 0, float64(bounds.Dx()), float64(bounds.Dy()), color.RGBA{A: alpha})
}
