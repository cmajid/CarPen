package cp_clock

import (
	"image"
	"time"

	"github.com/fogleman/gg"
)

var (
	angle float64
)

func ClockImage(t time.Time) image.Image {

	dc := gg.NewContext(200, 200)

	// dc.Translate(50, 50)
	// dc.SetColor(color.RGBA{0xff, 0, 0, 0xff})
	// dc.DrawRectangle(0, 0, 100, 100)
	// dc.Fill()

	// angle += 0.01
	// dc.RotateAbout(angle, 50, 50)
	// dc.DrawRectangle(0, 0, 100, 100)
	// dc.SetRGB(0, 0, 0)

	// dc.Fill()

	return dc.Image()
}
