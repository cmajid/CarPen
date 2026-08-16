---
name: carpen-architecture
description: Map of CarPen's packages, physics model, and code conventions. Read this before exploring or editing game code instead of scanning files — it tells you which file owns what and which invariants must hold.
---

# CarPen architecture

Module `github.com/cmajid/carpen`, three packages: `main` (window wiring only), `scene` (the screens and the manager that switches them) and `carpen` (entities, math). Direct dependencies are Ebitengine v2, `golang.org/x/image` (the Go fonts the menus are set in — Ebitengine requires it anyway), and `github.com/solarlune/resolv` (SAT collision — imported **only** by `carpen/collision.go`; see the `resolv-collision` skill before touching collision code).

## File map

| File | Owns |
|---|---|
| `main.go` | Nothing but window setup: builds a `scene.Manager` on `scene.NewMenu` and hands it to `ebiten.RunGame`. Game logic does not belong here. |
| `scene/scene.go` | `Scene` interface (`Update() (next Scene, err error)`) and the `Input` seam (keyboard **and** mouse, so menus are testable without a window). |
| `scene/manager.go` | `Manager`: the `ebiten.Game` impl, fixed 640×480 `Layout`, and the scene switch. |
| `scene/ui.go` | The design system every screen draws from: palette, type scale (Go fonts via `text/v2`), panels, the bottom prompt bar, scene fade. Change a colour or size **here**, not in a screen. |
| `scene/menulist.go` | `menuList`, the one focusable list all menus are built from: wraps at both ends, keyboard + mouse, focus shown by shape and weight as well as colour. |
| `scene/gameplay.go` | The race. Builds the world from a `carpen.Level`, maps keys to intent flags on the active car (Tab switches `activeCar`), builds the lot's four wall OBBs (`lotWalls`), checks the active car for collisions each tick and raises `CollisionEvent`s through `OnCollision` (rising-edge, seeded with the spawn state), and draws the F3 OBB overlay. |
| `scene/menu.go`, `scene/pause.go`, `scene/results.go` | The three menu screens. Each is a heading, a `menuList`, and a row of prompts; pause holds the live `*Gameplay` so resuming keeps its state. |
| `scene/scene_test.go`, `scene/menulist_test.go`, `scene/gameplay_test.go` | Scene switching, pause behaviour, input mapping, list navigation, and collision-event tests. |
| `carpen/car.go` | `Car` struct, physics, and drawing. The only file with nontrivial logic. |
| `carpen/vector.go` | `Vector` with `Length`/`Normalize`. `Normalize` returns zero for a zero vector — a deliberate NaN guard; a NaN here would spread to position and never leave. |
| `carpen/pivot.go`, `carpen/wheel.go` | Plain `{X, Y}` position structs (`Pivot`, `FrontPivot`, `RearPivot`, `RearPivotAbs`, `DirectionPivot`, `Direction`, `Wheel`). |
| `carpen/bush.go` | Static sprite obstacle with an `OBB()` derived from its position and sprite size. |
| `carpen/collision.go` | `OBB` (centre, size, degrees clockwise) and `Intersects` — the SAT test, and the **only** file that imports resolv. Details and the library's traps: `resolv-collision` skill. |
| `carpen/level.go` | Level format and loader (issue #19): `Level`, `Lot`, `Bay`, `Obstacle`, JSON files embedded from `carpen/levels/`, validated on load. |
| `carpen/assets.go` | `//go:embed assets/*.png`; entities load images via `ebitenutil.NewImageFromFileSystem`. Never load assets from disk paths. |
| `carpen/*_test.go` | Unit tests for car physics, vector math, level loading, and collision (`collision_test.go` pins the rotation direction and the containment case). |

## Physics model (car.go)

Kinematic front-wheel steering, all angles in **degrees** (converted with `math.Pi/180` at use sites):

- Key handlers only record intent (`Accelerate`, `RotateLeft`, …). **`Move()` is the one place `Speed` changes**, honouring `MaxSpeed` (6) and the reverse limit (−3); no input forces speed toward 0.
- `Steer()` steps `WheelAngle` by 2.4°/tick, clamped ±45°, and recomputes `DirectionPivot` → `Direction` (unit heading scaled by `Speed` in `UpdateDirection`).
- `Move()` advances `Pivot` along `Direction`, then derives body `Rotation` from the Pivot→RearPivotAbs vector (the "drift" block) and recomputes `RearPivotAbs`.
- `Update()` = `Move()` + `Steer()`, called from the game's `Update()` at Ebiten's fixed 60 ticks/s. **Never step physics from the draw path** — that ties speed to the display refresh rate.

## Collision (collision.go + gameplay.go)

OBB/SAT via the resolv library, wrapped so the game only ever speaks degrees-clockwise (full rules and resolv's traps: the `resolv-collision` skill):

- Entities build their `*OBB` once and re-place it per call (`Car.OBB()`, `Bush.OBB()`); `Intersects` is strict, so edge-touching is not a hit.
- Detection raises `carpen.CollisionEvent`; `Gameplay.OnCollision` alone decides what a crash means. Keep consequences out of detection code.
- Only the active car is checked — against walls, bushes, and the other cars — after all cars have stepped, so a hit lands the same tick.

## Rendering invariants

- No per-frame allocations: images are built once in `Init()` (`wheelImage` pattern) and placed with a fresh `GeoM` per draw. Keep it that way — issue #14 removed per-frame image churn.
- `GeoM` calls apply innermost-first (the reverse of a push/pop stack); see `wheelGeoM` for the canonical example.

## Conventions & CI gates

- `gofmt` is enforced (CI fails on any `gofmt -l` output).
- `js/wasm` and `windows/amd64` must keep compiling with `CGO_ENABLED=0`.
- Roadmap lives in epic #17; Phase 0–1 sub-issues are #18–#23.
