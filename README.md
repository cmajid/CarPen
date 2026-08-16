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
