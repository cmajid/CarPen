---
name: carpen-architecture
description: Map of CarPen's packages, physics model, and code conventions. Read this before exploring or editing game code instead of scanning files — it tells you which file owns what and which invariants must hold.
---

# CarPen architecture

Module `github.com/cmajid/carpen`, three packages: `main` (window wiring only), `scene` (the screens and the manager that switches them) and `carpen` (entities, math). Direct dependencies are Ebitengine v2 and `golang.org/x/image` (the Go fonts the menus are set in — Ebitengine requires it anyway).

## File map

| File | Owns |
|---|---|
| `main.go` | Nothing but window setup: builds a `scene.Manager` on `scene.NewMenu` and hands it to `ebiten.RunGame`. Game logic does not belong here. |
| `scene/scene.go` | `Scene` interface (`Update() (next Scene, err error)`) and the `Input` seam (keyboard **and** mouse, so menus are testable without a window). |
| `scene/manager.go` | `Manager`: the `ebiten.Game` impl, fixed 640×480 `Layout`, and the scene switch. |
| `scene/ui.go` | The design system every screen draws from: palette, type scale (Go fonts via `text/v2`), panels, the bottom prompt bar, scene fade. Change a colour or size **here**, not in a screen. |
| `scene/menulist.go` | `menuList`, the one focusable list all menus are built from: wraps at both ends, keyboard + mouse, focus shown by shape and weight as well as colour. |
| `scene/gameplay.go` | The race. Builds the world (2 cars, 2 bushes — becomes level data per issue #19) and maps keys to intent flags on the active car; Tab switches `activeCar`. |
| `scene/menu.go`, `scene/pause.go`, `scene/results.go` | The three menu screens. Each is a heading, a `menuList`, and a row of prompts; pause holds the live `*Gameplay` so resuming keeps its state. |
| `scene/scene_test.go`, `scene/menulist_test.go` | Scene switching, pause behaviour, input mapping, and list navigation tests. |
| `carpen/car.go` | `Car` struct, physics, and drawing. The only file with nontrivial logic. |
| `carpen/vector.go` | `Vector` with `Length`/`Normalize`. `Normalize` returns zero for a zero vector — a deliberate NaN guard; a NaN here would spread to position and never leave. |
| `carpen/pivot.go`, `carpen/wheel.go` | Plain `{X, Y}` position structs (`Pivot`, `FrontPivot`, `RearPivot`, `RearPivotAbs`, `DirectionPivot`, `Direction`, `Wheel`). |
| `carpen/bush.go` | Static sprite obstacle. No collision yet (issue #20). |
| `carpen/assets.go` | `//go:embed assets/*.png`; entities load images via `ebitenutil.NewImageFromFileSystem`. Never load assets from disk paths. |
| `carpen/*_test.go` | Unit tests for car physics and vector math. |

## Physics model (car.go)

Kinematic front-wheel steering, all angles in **degrees** (converted with `math.Pi/180` at use sites):

- Key handlers only record intent (`Accelerate`, `RotateLeft`, …). **`Move()` is the one place `Speed` changes**, honouring `MaxSpeed` (6) and the reverse limit (−3); no input forces speed toward 0.
- `Steer()` steps `WheelAngle` by 2.4°/tick, clamped ±45°, and recomputes `DirectionPivot` → `Direction` (unit heading scaled by `Speed` in `UpdateDirection`).
- `Move()` advances `Pivot` along `Direction`, then derives body `Rotation` from the Pivot→RearPivotAbs vector (the "drift" block) and recomputes `RearPivotAbs`.
- `Update()` = `Move()` + `Steer()`, called from the game's `Update()` at Ebiten's fixed 60 ticks/s. **Never step physics from the draw path** — that ties speed to the display refresh rate.

## Rendering invariants

- No per-frame allocations: images are built once in `Init()` (`wheelImage` pattern) and placed with a fresh `GeoM` per draw. Keep it that way — issue #14 removed per-frame image churn.
- `GeoM` calls apply innermost-first (the reverse of a push/pop stack); see `wheelGeoM` for the canonical example.

## Conventions & CI gates

- `gofmt` is enforced (CI fails on any `gofmt -l` output).
- `js/wasm` and `windows/amd64` must keep compiling with `CGO_ENABLED=0`.
- Roadmap lives in epic #17; Phase 0–1 sub-issues are #18–#23.
