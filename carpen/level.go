package carpen

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"path"
	"strings"
)

// A level is data, not code: where the player's car starts, the bay it has to
// end up in, and what stands in the way. The definitions live in levels/ as
// JSON and are compiled into the binary the same way the sprites are, so adding
// a level is adding one file.
//
//go:embed levels/*.json
var levelFiles embed.FS

// levelDir is where the definitions sit inside levelFiles.
const levelDir = "levels"

// Level is one parking puzzle.
type Level struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Lot       Lot        `json:"lot"`
	Car       CarStart   `json:"car"`
	Bay       Bay        `json:"bay"`
	Obstacles []Obstacle `json:"obstacles"`

	// Attempts is how many tries the player gets before the level is failed.
	Attempts int `json:"attempts"`
}

// Lot is the ground the level is played on, in pixels.
type Lot struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// CarStart is where the player's car is parked at the start of the level. X and
// Y are the car's own X and Y, the point a Car is built around, and Rotation is
// in degrees clockwise, as everything else on the lot is.
type CarStart struct {
	Color    string  `json:"color"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Rotation float64 `json:"rotation"`
}

// Bay is the space the car has to be parked in. X and Y are the middle of the
// bay rather than a corner, so a bay is turned about its own centre and the win
// condition has the point to measure to. Width is across the bay, Height is
// along it, and the car drives in over the near edge — the one Rotation leaves
// facing up at 0 degrees.
type Bay struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Rotation float64 `json:"rotation"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
}

// Corners returns where the bay's four corners fall on the lot, with the near
// pair on the edge the car drives in over. Everything that has to reason about
// the bay — drawing it, and later deciding whether a car is parked in it — works
// from these rather than turning the rectangle itself a second time.
func (b Bay) Corners() (nearLeft, nearRight, farLeft, farRight Vector) {
	halfWidth, halfHeight := b.Width/2, b.Height/2

	return b.point(-halfWidth, -halfHeight),
		b.point(halfWidth, -halfHeight),
		b.point(-halfWidth, halfHeight),
		b.point(halfWidth, halfHeight)
}

// point places a point given in the bay's own frame, where the bay's middle is
// the origin and the entrance is towards -Y, onto the lot.
func (b Bay) point(x, y float64) Vector {
	sin, cos := math.Sincos(b.Rotation * math.Pi / 180)
	return Vector{
		X: x*cos - y*sin + b.X,
		Y: x*sin + y*cos + b.Y,
	}
}

// ObstacleType is what an obstacle is made of. Each one is drawn from a sprite
// that is already embedded, so the set is closed and the loader can reject a
// level asking for anything else.
type ObstacleType string

const (
	ObstacleBush ObstacleType = "bush"
	ObstacleCar  ObstacleType = "car"
)

// Obstacle is one thing in the way. Color and Rotation are only read for a
// parked car; a bush is round and comes in one shade.
type Obstacle struct {
	Type     ObstacleType `json:"type"`
	Color    string       `json:"color"`
	X        float64      `json:"x"`
	Y        float64      `json:"y"`
	Rotation float64      `json:"rotation"`
}

// defaultCarColor is the paint the player's car turns up in when a level does
// not ask for one. It is the car the game is illustrated with, on the menu and
// in the screenshot.
const defaultCarColor = "yellow"

// Levels returns the game's own levels, in the order their files sort in —
// level-01, level-02, and so on. It reads them afresh each call; the levels are
// small and this happens once at startup.
func Levels() ([]Level, error) {
	dir, err := fs.Sub(levelFiles, levelDir)
	if err != nil {
		return nil, err
	}
	return LoadLevels(dir)
}

// LoadLevels reads every .json file at the root of fsys as a level. Levels come
// back in filename order, which is the order they are meant to be played in, and
// one bad file fails the lot: a level that cannot be loaded is a level the
// player would be dropped into a broken world by.
func LoadLevels(fsys fs.FS) ([]Level, error) {
	names, err := fs.Glob(fsys, "*.json")
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, errors.New("no level files found")
	}

	levels := make([]Level, 0, len(names))
	seen := make(map[string]string, len(names))
	for _, name := range names {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, err
		}

		level, err := ParseLevel(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if first, duplicate := seen[level.ID]; duplicate {
			return nil, fmt.Errorf("%s: level id %q is already used by %s", name, level.ID, first)
		}

		seen[level.ID] = name
		levels = append(levels, level)
	}

	return levels, nil
}

// ParseLevel reads one level definition and checks that it describes a puzzle
// that can actually be played. Unknown fields are refused along with missing
// ones, so a misspelled key is reported rather than quietly ignored.
func ParseLevel(data []byte) (Level, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var level Level
	if err := decoder.Decode(&level); err != nil {
		return Level{}, fmt.Errorf("malformed level: %w", err)
	}
	if decoder.More() {
		return Level{}, errors.New("malformed level: more than one level in the file")
	}

	if level.Car.Color == "" {
		level.Car.Color = defaultCarColor
	}
	if err := level.validate(); err != nil {
		return Level{}, err
	}

	return level, nil
}

// validate reports the first thing that would leave the level unplayable. The
// messages name the field, because the reader of one is looking at a JSON file
// and needs to know which line to fix.
func (l Level) validate() error {
	if l.ID == "" {
		return errors.New("id is missing")
	}
	if l.Name == "" {
		return fmt.Errorf("level %q: name is missing", l.ID)
	}
	if l.Lot.Width <= 0 || l.Lot.Height <= 0 {
		return fmt.Errorf("level %q: lot needs a positive width and height, got %gx%g", l.ID, l.Lot.Width, l.Lot.Height)
	}
	if l.Attempts <= 0 {
		return fmt.Errorf("level %q: attempts must be at least 1, got %d", l.ID, l.Attempts)
	}

	if err := checkCarColor(l.Car.Color); err != nil {
		return fmt.Errorf("level %q: car: %w", l.ID, err)
	}
	if !l.Lot.contains(l.Car.X, l.Car.Y) {
		return fmt.Errorf("level %q: car starts at %g,%g, outside the lot", l.ID, l.Car.X, l.Car.Y)
	}

	// A level with no bay is a level with nowhere to park, which is the one
	// thing the game asks the player to do.
	if l.Bay.Width <= 0 || l.Bay.Height <= 0 {
		return fmt.Errorf("level %q: bay needs a positive width and height, got %gx%g", l.ID, l.Bay.Width, l.Bay.Height)
	}
	if !l.Lot.contains(l.Bay.X, l.Bay.Y) {
		return fmt.Errorf("level %q: bay sits at %g,%g, outside the lot", l.ID, l.Bay.X, l.Bay.Y)
	}

	for i, obstacle := range l.Obstacles {
		if err := obstacle.validate(); err != nil {
			return fmt.Errorf("level %q: obstacle %d: %w", l.ID, i, err)
		}
		if !l.Lot.contains(obstacle.X, obstacle.Y) {
			return fmt.Errorf("level %q: obstacle %d sits at %g,%g, outside the lot", l.ID, i, obstacle.X, obstacle.Y)
		}
	}

	return nil
}

func (o Obstacle) validate() error {
	switch o.Type {
	case ObstacleBush:
		if o.Color != "" {
			return fmt.Errorf("a bush has no colour, got %q", o.Color)
		}
	case ObstacleCar:
		if o.Color == "" {
			return errors.New("a parked car needs a colour")
		}
		return checkCarColor(o.Color)
	case "":
		return errors.New("type is missing")
	default:
		return fmt.Errorf("unknown type %q, want %q or %q", o.Type, ObstacleBush, ObstacleCar)
	}

	return nil
}

// contains reports whether a point is on the lot. Sprites are drawn from their
// anchor and can hang over an edge on purpose, so this is only asked about the
// anchor itself — enough to catch a coordinate typed with a digit too many.
func (l Lot) contains(x, y float64) bool {
	return x >= 0 && y >= 0 && x <= l.Width && y <= l.Height
}

// checkCarColor accepts the colours the game has a car sprite for. It reads the
// embedded assets rather than a list written out here, so dropping in
// car-blue.png is all it takes to make "blue" a colour levels can ask for.
func checkCarColor(colour string) error {
	if _, err := fs.Stat(assets, "assets/car-"+colour+".png"); err == nil {
		return nil
	}
	return fmt.Errorf("unknown colour %q, have %s", colour, strings.Join(carColors(), ", "))
}

// carColors is the colours car sprites are embedded for, sorted.
func carColors() []string {
	names, err := fs.Glob(assets, "assets/car-*.png")
	if err != nil {
		return nil
	}

	colours := make([]string, 0, len(names))
	for _, name := range names {
		colours = append(colours, strings.TrimSuffix(strings.TrimPrefix(path.Base(name), "car-"), ".png"))
	}
	return colours
}
