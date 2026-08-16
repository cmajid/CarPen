package carpen

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"
)

// validLevel is a level that loads, written out as the JSON a level file holds.
// The tests that check a rejection break one thing in it, so what each of them
// is about is the one field it changes.
const validLevel = `{
  "id": "level-99",
  "name": "A Test Level",
  "lot": { "width": 640, "height": 480 },
  "car": { "color": "yellow", "x": 400, "y": 300, "rotation": 0 },
  "bay": { "x": 150, "y": 330, "rotation": 0, "width": 130, "height": 210 },
  "obstacles": [
    { "type": "bush", "x": 0, "y": 0 },
    { "type": "car", "color": "red", "x": 350, "y": 100, "rotation": 90 }
  ],
  "attempts": 3
}`

// levelWithout is validLevel with one field taken out, which is how a file that
// forgot it would arrive.
func levelWithout(t *testing.T, field string) string {
	t.Helper()

	var fields map[string]any
	if err := json.Unmarshal([]byte(validLevel), &fields); err != nil {
		t.Fatalf("unmarshalling the valid level: %v", err)
	}
	if _, ok := fields[field]; !ok {
		t.Fatalf("the valid level has no %q field to remove", field)
	}
	delete(fields, field)

	data, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("marshalling the level back: %v", err)
	}
	return string(data)
}

func TestParseLevelReadsEveryField(t *testing.T) {
	level, err := ParseLevel([]byte(validLevel))
	if err != nil {
		t.Fatalf("ParseLevel: %v", err)
	}

	want := Level{
		ID:   "level-99",
		Name: "A Test Level",
		Lot:  Lot{Width: 640, Height: 480},
		Car:  CarStart{Color: "yellow", X: 400, Y: 300},
		Bay:  Bay{X: 150, Y: 330, Width: 130, Height: 210},
		Obstacles: []Obstacle{
			{Type: ObstacleBush},
			{Type: ObstacleCar, Color: "red", X: 350, Y: 100, Rotation: 90},
		},
		Attempts: 3,
	}
	if !levelsEqual(level, want) {
		t.Errorf("ParseLevel = %+v, want %+v", level, want)
	}
}

// A level that does not name a colour for the player's car gets the game's own
// yellow one, which is the car the menu and the screenshot show.
func TestParseLevelDefaultsTheCarColour(t *testing.T) {
	level, err := ParseLevel([]byte(`{
	  "id": "level-99", "name": "A Test Level",
	  "lot": { "width": 640, "height": 480 },
	  "car": { "x": 400, "y": 300 },
	  "bay": { "x": 150, "y": 330, "width": 130, "height": 210 },
	  "attempts": 3
	}`))
	if err != nil {
		t.Fatalf("ParseLevel: %v", err)
	}

	if level.Car.Color != "yellow" {
		t.Errorf("car colour = %q, want yellow", level.Car.Color)
	}
}

// A level file that would leave the player in a world they cannot play is
// refused, and the complaint names what is wrong with it.
func TestParseLevelRejects(t *testing.T) {
	for _, tc := range []struct {
		name  string
		level string
		wants []string // fragments the error has to mention
	}{
		{
			name:  "a file that is not JSON at all",
			level: "not json",
			wants: []string{"malformed level"},
		},
		{
			name:  "a truncated file",
			level: strings.TrimSuffix(validLevel, "}"),
			wants: []string{"malformed level", "unexpected EOF"},
		},
		{
			name:  "a field of the wrong type",
			level: strings.Replace(validLevel, `"width": 640`, `"width": "wide"`, 1),
			wants: []string{"malformed level", "width"},
		},
		{
			name:  "a misspelled field",
			level: strings.Replace(validLevel, `"attempts"`, `"attemps"`, 1),
			wants: []string{"malformed level", "attemps"},
		},
		{
			name:  "no id",
			level: levelWithout(t, "id"),
			wants: []string{"id"},
		},
		{
			name:  "no name",
			level: levelWithout(t, "name"),
			wants: []string{"name"},
		},
		{
			name:  "no lot",
			level: levelWithout(t, "lot"),
			wants: []string{"lot", "positive"},
		},
		{
			name:  "no bay to park in",
			level: levelWithout(t, "bay"),
			wants: []string{"bay", "positive"},
		},
		{
			name:  "a bay of no size",
			level: strings.Replace(validLevel, `"width": 130`, `"width": 0`, 1),
			wants: []string{"bay", "positive"},
		},
		{
			name:  "a bay off the lot",
			level: strings.Replace(validLevel, `"x": 150`, `"x": 1500`, 1),
			wants: []string{"bay", "outside the lot"},
		},
		{
			name:  "no attempts",
			level: levelWithout(t, "attempts"),
			wants: []string{"attempts", "at least 1"},
		},
		{
			name:  "a car in a colour there is no sprite for",
			level: strings.Replace(validLevel, `"color": "yellow"`, `"color": "chartreuse"`, 1),
			wants: []string{"car", "chartreuse", "yellow"},
		},
		{
			name:  "a car starting off the lot",
			level: strings.Replace(validLevel, `"y": 300`, `"y": -20`, 1),
			wants: []string{"car", "outside the lot"},
		},
		{
			name:  "an obstacle of an unknown kind",
			level: strings.Replace(validLevel, `"type": "bush"`, `"type": "fountain"`, 1),
			wants: []string{"obstacle 0", "fountain", "bush", "car"},
		},
		{
			name:  "an obstacle with no kind",
			level: strings.Replace(validLevel, `"type": "bush"`, `"rotation": 0`, 1),
			wants: []string{"obstacle 0", "type is missing"},
		},
		{
			name:  "a parked car with no colour",
			level: strings.Replace(validLevel, `"type": "car", "color": "red"`, `"type": "car"`, 1),
			wants: []string{"obstacle 1", "colour"},
		},
		{
			name:  "a parked car in a colour there is no sprite for",
			level: strings.Replace(validLevel, `"color": "red"`, `"color": "puce"`, 1),
			wants: []string{"obstacle 1", "puce"},
		},
		{
			name:  "a bush painted a colour",
			level: strings.Replace(validLevel, `"type": "bush"`, `"type": "bush", "color": "red"`, 1),
			wants: []string{"obstacle 0", "bush has no colour"},
		},
		{
			name:  "an obstacle off the lot",
			level: strings.Replace(validLevel, `"x": 350`, `"x": 3500`, 1),
			wants: []string{"obstacle 1", "outside the lot"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseLevel([]byte(tc.level))
			if err == nil {
				t.Fatal("ParseLevel accepted the level, want an error")
			}

			for _, want := range tc.wants {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not mention %q", err, want)
				}
			}
		})
	}
}

// Levels are played in the order their files sort in, so the filenames are what
// decides the order rather than anything written inside them.
func TestLoadLevelsReadsEveryFileInOrder(t *testing.T) {
	files := fstest.MapFS{
		"level-02.json": {Data: []byte(strings.Replace(validLevel, "level-99", "level-02", 1))},
		"level-01.json": {Data: []byte(strings.Replace(validLevel, "level-99", "level-01", 1))},
		"README.md":     {Data: []byte("not a level")},
	}

	levels, err := LoadLevels(files)
	if err != nil {
		t.Fatalf("LoadLevels: %v", err)
	}

	if len(levels) != 2 {
		t.Fatalf("loaded %d levels, want the 2 .json files", len(levels))
	}
	if levels[0].ID != "level-01" || levels[1].ID != "level-02" {
		t.Errorf("levels came back as %q, %q, want them in filename order", levels[0].ID, levels[1].ID)
	}
}

// One unreadable file fails the whole load: a level that cannot be parsed is a
// level the player would otherwise be dropped into a broken world by, and the
// error has to say which file to go and look at.
func TestLoadLevelsFailsOnABadFile(t *testing.T) {
	files := fstest.MapFS{
		"level-01.json": {Data: []byte(validLevel)},
		"level-02.json": {Data: []byte("{ oh dear")},
	}

	_, err := LoadLevels(files)

	if err == nil {
		t.Fatal("LoadLevels accepted a broken file, want an error")
	}
	if !strings.Contains(err.Error(), "level-02.json") {
		t.Errorf("error %q does not name the file that failed", err)
	}
}

// Two levels with the same id would be two different puzzles under one name,
// which is how a level gets lost by being copied and half-edited.
func TestLoadLevelsRejectsDuplicateIDs(t *testing.T) {
	files := fstest.MapFS{
		"level-01.json": {Data: []byte(validLevel)},
		"level-02.json": {Data: []byte(validLevel)},
	}

	_, err := LoadLevels(files)

	if err == nil {
		t.Fatal("LoadLevels accepted two levels with one id, want an error")
	}
	if !strings.Contains(err.Error(), "level-99") {
		t.Errorf("error %q does not name the duplicated id", err)
	}
}

func TestLoadLevelsFailsWhenThereAreNone(t *testing.T) {
	if _, err := LoadLevels(fstest.MapFS{}); err == nil {
		t.Error("LoadLevels accepted an empty directory, want an error")
	}
}

// The levels the game ships with are compiled into the binary, so this is the
// test that they are all loadable at all.
func TestLevelsLoadsTheGamesOwnLevels(t *testing.T) {
	levels, err := Levels()
	if err != nil {
		t.Fatalf("Levels: %v", err)
	}

	if len(levels) == 0 {
		t.Fatal("no levels are embedded")
	}
}

// The first level is the prototype's world, moved into data field for field.
// This is what says the move changed the shape of the code and not the game.
func TestFirstLevelIsThePrototypesWorld(t *testing.T) {
	levels, err := Levels()
	if err != nil {
		t.Fatalf("Levels: %v", err)
	}

	want := Level{
		ID:   "level-01",
		Name: "First Steps",
		Lot:  Lot{Width: 640, Height: 480},
		Car:  CarStart{Color: "yellow", X: 400, Y: 300, Rotation: 0},
		Bay:  Bay{X: 150, Y: 330, Rotation: 0, Width: 130, Height: 210},
		Obstacles: []Obstacle{
			{Type: ObstacleBush, X: 0, Y: 0},
			{Type: ObstacleBush, X: 100, Y: 100},
			{Type: ObstacleCar, Color: "red", X: 350, Y: 100, Rotation: 90},
		},
		Attempts: 3,
	}
	if !levelsEqual(levels[0], want) {
		t.Errorf("level-01 = %+v, want the prototype's world %+v", levels[0], want)
	}
}

func TestBayCornersTurnWithTheBay(t *testing.T) {
	for _, tc := range []struct {
		name                                   string
		bay                                    Bay
		nearLeft, nearRight, farLeft, farRight Vector
	}{
		{
			name:      "unturned, with the entrance at the top",
			bay:       Bay{X: 100, Y: 200, Width: 40, Height: 100},
			nearLeft:  Vector{X: 80, Y: 150},
			nearRight: Vector{X: 120, Y: 150},
			farLeft:   Vector{X: 80, Y: 250},
			farRight:  Vector{X: 120, Y: 250},
		},
		{
			name:      "a quarter turn, with the entrance on the left",
			bay:       Bay{X: 100, Y: 200, Rotation: 90, Width: 40, Height: 100},
			nearLeft:  Vector{X: 150, Y: 180},
			nearRight: Vector{X: 150, Y: 220},
			farLeft:   Vector{X: 50, Y: 180},
			farRight:  Vector{X: 50, Y: 220},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nearLeft, nearRight, farLeft, farRight := tc.bay.Corners()

			for _, corner := range []struct {
				name      string
				got, want Vector
			}{
				{"near left", nearLeft, tc.nearLeft},
				{"near right", nearRight, tc.nearRight},
				{"far left", farLeft, tc.farLeft},
				{"far right", farRight, tc.farRight},
			} {
				if !nearlyEqual(corner.got, corner.want) {
					t.Errorf("%s corner = %v, want %v", corner.name, corner.got, corner.want)
				}
			}
		})
	}
}

// levelsEqual compares two levels field by field. reflect.DeepEqual would do it,
// but a level with no obstacles reads as an empty list in one and a nil one in
// the other, and that is not a difference worth failing a test over.
func levelsEqual(a, b Level) bool {
	if a.ID != b.ID || a.Name != b.Name || a.Lot != b.Lot || a.Car != b.Car || a.Bay != b.Bay || a.Attempts != b.Attempts {
		return false
	}
	if len(a.Obstacles) != len(b.Obstacles) {
		return false
	}
	for i := range a.Obstacles {
		if a.Obstacles[i] != b.Obstacles[i] {
			return false
		}
	}
	return true
}

// nearlyEqual compares two points loosely, because turning one goes through
// sine and cosine and lands a hair off the round number it is aiming at.
func nearlyEqual(a, b Vector) bool {
	const tolerance = 1e-9
	d := Vector{X: a.X - b.X, Y: a.Y - b.Y}
	return d.Length() < tolerance
}
