# CarPen
Mini Game: Park your car carefully!
CarPen wrote with GoLang using [Ebiten](https://github.com/hajimehoshi/ebiten)


<p align="center">
  <img width="350" height="300" src="./Screen_Shot.png">
</p>

## Instructions

Requires Go 1.25 or newer.

1. Clone the project: `git clone https://github.com/cmajid/CarPen.git`
2. Change directory to `CarPen/`: `cd CarPen/`
3. Run `go run .` or run `go build` 

The window opens as large as the monitor allows in whole multiples of the
game's size, and can be resized to any shape; `[F11]` toggles fullscreen.

Once the window is open, one can do the following:

In the menus:

* `[Up]` / `[Down]` - Move between choices
* `[Enter]` or `[Space]` - Take the highlighted choice
* `[Esc]` - Go back a step, and quit from the main menu
* The mouse works too: hover to highlight, click to choose

In a race:

* `[Arrow-Keys]` - Move the car
* `[Tab]` - Switch between the yellow and the red car
* `[Esc]` - Pause, for resume / restart / quit to menu
* `[Enter]` - Finish the race (standing in for the win condition, which is still to come)

Enjoy!

## iOS

Requires macOS with Xcode, alongside Go.

Xcode cannot compile Go, so the game is built in two halves: everything under
[`carpen/`](carpen/), [`scene/`](scene/) and [`mobile/`](mobile/) is bound into a
framework, and the Xcode project is a thin shell that links it. Only the first
half is generated — the Swift files in [`ios/CarPen/`](ios/CarPen/) and the Xcode
project are hand-written and nothing overwrites them.

**Generate the framework**, which is also how you update it:

```sh
./ios/build.sh
```

It installs `ebitenmobile` if it is not already on your `PATH`, then binds the
`mobile` package into `dist/Mobile.xcframework`. It takes a couple of minutes.
The same script does both jobs — there is no separate update command.

Then open the project and run it on a simulator or a device:

```sh
open ios/CarPen.xcodeproj
```

**Re-run `./ios/build.sh` after any change under `carpen/`, `scene/` or
`mobile/`.** Nothing does it for you: the Xcode project has no script build
phase, and simply links whatever the script last left in `dist/`, so building
without re-running it quietly ships the previous version of the game. Changes to
the Swift files alone need only an Xcode build.

`dist/` is not in the repository, so a fresh clone has to run the script once
before Xcode will build at all.

If you would rather call the tool directly, that script is only wrapping:

```sh
go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
ebitenmobile bind -target ios -o dist/Mobile.xcframework ./mobile
```

The framework has to be called `Mobile`: gomobile names the Swift module after
the Go package it bound, and Xcode will only find a module inside a framework of
the same name.

On a phone or a tablet the game is landscape-only and drives itself by touch —
a steering stick under the left thumb, the pedals under the right, and the rest
along the top. The controls appear the first time the screen is touched, and
size themselves to the device, so a pad paired to the phone still works.

## Levels

A level is data, not code. Each one is a JSON file in [`carpen/levels/`](carpen/levels/),
compiled into the binary, so adding a level is adding a file — nothing in the
game has to change:

```json
{
  "id": "level-02",
  "name": "Tight Squeeze",
  "lot": { "width": 640, "height": 480 },
  "car": { "color": "yellow", "x": 400, "y": 300, "rotation": 0 },
  "bay": { "x": 150, "y": 330, "rotation": 0, "width": 130, "height": 210 },
  "obstacles": [
    { "type": "bush", "x": 100, "y": 100 },
    { "type": "car", "color": "red", "x": 350, "y": 100, "rotation": 90 }
  ],
  "attempts": 3
}
```

* `id` is unique across levels, and levels are played in filename order
* `lot` is the ground the level is played on, in pixels
* `car` is where the player starts; `color` is optional and is yellow by default
* `bay` is the space to park in, given by its **middle**, turned about that
  point by `rotation` degrees, and entered over the edge that faces up at 0
* `obstacles` are `bush`es and parked `car`s
* a car's `color` is any colour there is a `carpen/assets/car-<colour>.png` for

The game refuses to start on a level file it cannot make sense of, and says
which file and which field is at fault.
