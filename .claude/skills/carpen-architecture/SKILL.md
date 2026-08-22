---
name: carpen-architecture
description: Map of CarPen's packages, physics model, and code conventions. Read this before exploring or editing game code instead of scanning files — it tells you which file owns what and which invariants must hold.
---

# CarPen architecture

Module `github.com/cmajid/carpen`, three packages: `main` (window wiring only), `scene` (the screens and the manager that switches them) and `carpen` (entities, math). Direct dependencies are Ebitengine v2, `golang.org/x/image` (the Go fonts the menus are set in — Ebitengine requires it anyway), and `github.com/solarlune/resolv` (SAT collision — imported **only** by `carpen/collision.go`; see the `resolv-collision` skill before touching collision code).

## File map

| File | Owns |
|---|---|
| `main.go` | Nothing but window setup: builds a `scene.Manager` on `scene.NewMenu`, opens a resizable window at the largest whole multiple of the design size the monitor holds, and hands it to `ebiten.RunGame`. Game logic does not belong here. |
| `scene/scene.go` | `Scene` interface (`Update() (next Scene, err error)`) and the `Input` seam — keyboard, mouse and pad, as edges (`…JustPressed`) **and** as levels (`IsKeyPressed`, `GamepadButtonValue`, `GamepadAxisValue`), so every screen is testable without a window. |
| `scene/manager.go` | `Manager`: the `ebiten.Game` impl, the scene switch, the F11 fullscreen toggle, and `Layout` — which holds the height at 480, follows the device's aspect ratio for the width (a phone gets ~1041, a tablet ~640–731), and records `pointsPerPixel`, what one game pixel is in the device's own units. Scenes receive all three as a `viewport`; anything a thumb has to hit is sized through `viewport.size`, never in game pixels alone. Also the one home of the design size, `DesignWidth`/`DesignHeight` (640×480) — never write those numbers anywhere else. |
| `scene/ui.go` | The design system every screen draws from: palette, type scale (Go fonts via `text/v2`), panels, the bottom prompt bar, scene fade. Change a colour or size **here**, not in a screen. |
| `scene/menulist.go` | `menuList`, the one focusable list all menus are built from: wraps at both ends, keyboard + mouse, focus shown by shape and weight as well as colour. |
| `scene/controls.go` | The binding table: `action` (what the player *means*) mapped to keys, pad buttons and — for steering — a stick axis. `justPressed`/`justReleased` read edges; `analog` reads how far a control is being asked for, in 0..1, past `analogDeadzone`. Rebind **here**, never at a use site. |
| `scene/gamepad.go` | Finding a pad in Ebiten's standard layout and reading it: `buttonValue`/`axisValue` for analogue travel, plus the left stick standing in for the d-pad's four buttons so menus get stick presses for free. |
| `scene/gameplay.go` | The race. Builds the world from a `carpen.Level`, reads the controls onto the active car every tick (Tab switches `activeCar`; `releaseControls` first, or the car handed over drives on forever), builds the lot's four wall OBBs (`lotWalls`), checks the active car for collisions each tick and raises `CollisionEvent`s through `OnCollision` (rising-edge, seeded with the spawn state), and draws the F3 OBB overlay. |
| `scene/menu.go`, `scene/pause.go`, `scene/results.go` | The three menu screens. Each is a heading, a `menuList`, and a row of prompts; pause holds the live `*Gameplay` so resuming keeps its state. |
| `scene/scene_test.go`, `scene/menulist_test.go`, `scene/gameplay_test.go`, `scene/controls_test.go` | Scene switching, pause and swap behaviour, input mapping (keys, pad buttons, trigger and stick travel), list navigation, and collision-event tests. `fakeInput` in `scene_test.go` is the device stub every screen test drives. |
| `carpen/car.go` | `Car` struct, physics, and drawing. The only file with nontrivial logic. |
| `carpen/vector.go` | `Vector` with `Length`/`Normalize`. `Normalize` returns zero for a zero vector — a deliberate NaN guard; a NaN here would spread to position and never leave. |
| `carpen/pivot.go`, `carpen/wheel.go` | Plain `{X, Y}` position structs (`Pivot`, `FrontPivot`, `RearPivot`, `RearPivotAbs`, `DirectionPivot`, `Direction`, `Wheel`). |
| `carpen/bush.go` | Static sprite obstacle with a round `Collider()` derived from its position and sprite size. |
| `carpen/collision.go` | `OBB` (plain or corner-beveled, degrees clockwise), `Circle`, and the SAT tests `Intersects` / `IntersectsOBB` — the **only** file that imports resolv. Details and the library's traps: `resolv-collision` skill. |
| `carpen/level.go` | Level format and loader (issue #19): `Level`, `Lot`, `Bay`, `Obstacle`, JSON files embedded from `carpen/levels/`, validated on load. |
| `carpen/assets.go` | `//go:embed assets/*.png`; entities load images via `ebitenutil.NewImageFromFileSystem`. Never load assets from disk paths. |
| `carpen/*_test.go` | Unit tests for car physics, vector math, level loading, and collision (`collision_test.go` pins the rotation direction and the containment case). |

## Physics model (car.go)

Kinematic front-wheel steering, all angles in **degrees** (converted with `math.Pi/180` at use sites):

- The controls are **analogue**: `Throttle` and `Brake` in 0..1, `Steering` in −1..1 (negative is left). A key is only ever 0 or 1; a trigger and a stick fill in between. They record intent only — **`Move()` is the one place `Speed` changes**, honouring `MaxSpeed` (6) and the reverse limit (−3); no input forces speed toward 0.
- `Move()` scales acceleration by the pedal (`Acceleration * Throttle`). Coasting and the crawl back to a stop are **unscaled** — the car slowing itself is not the player asking for anything.
- `Steer()` treats `Steering` as a **position**: it picks a target of `Steering × WheelMaxAngle` and walks `WheelAngle` to it at 2.4°/tick, clamped ±45°. `Steering == 0` means *nothing asked*, not *straight ahead* — the wheels stay where they were left, so there is **no self-centring**. Then it recomputes `DirectionPivot` → `Direction` (unit heading scaled by `Speed` in `UpdateDirection`).
- `Move()` advances `Pivot` along `Direction`, then derives body `Rotation` from the Pivot→RearPivotAbs vector (the "drift" block) and recomputes `RearPivotAbs`.
- `Update()` = `Move()` + `Steer()`, called from the game's `Update()` at Ebiten's fixed 60 ticks/s. **Never step physics from the draw path** — that ties speed to the display refresh rate.

## Collision (collision.go + gameplay.go)

SAT via the resolv library, wrapped so the game only ever speaks degrees-clockwise (full rules and resolv's traps: the `resolv-collision` skill):

- Entities build their collider once and re-place it per call: `Car.OBB()` is a corner-beveled octagon, `Bush.Collider()` a circle, both inset a few px from the sprite for forgiveness (tuning constants in car.go/bush.go). `Intersects` is strict, so edge-touching is not a hit.
- Detection raises `carpen.CollisionEvent`; `Gameplay.OnCollision` alone decides what a crash means. Keep consequences out of detection code.
- Only the active car is checked — against walls, bushes, and the other cars — after all cars have stepped, so a hit lands the same tick.

## Rendering invariants

- No per-frame allocations: images are built once in `Init()` (`wheelImage` pattern) and placed with a fresh `GeoM` per draw. Keep it that way — issue #14 removed per-frame image churn.
- `GeoM` calls apply innermost-first (the reverse of a push/pop stack); see `wheelGeoM` for the canonical example.

## Conventions & CI gates

- `gofmt` is enforced (CI fails on any `gofmt -l` output).
- `js/wasm` and `windows/amd64` must keep compiling with `CGO_ENABLED=0`.
- Roadmap lives in epic #17; Phase 0–1 sub-issues are #18–#23.
