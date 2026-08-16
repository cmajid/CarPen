package carpen

import "embed"

// assets holds the game's images. They are compiled into the binary so the game
// runs from any working directory instead of only from the repository root.
//
//go:embed assets/*.png
var assets embed.FS
