---
name: carpen-architecture
description: Map of CarPen's packages, physics model, and code conventions. Read this before exploring or editing game code instead of scanning files — it tells you which file owns what and which invariants must hold.
---

# CarPen architecture

Module `github.com/cmajid/carpen`, two packages: `main` (game loop, world setup) and `carpen` (entities, math). Ebitengine v2 is the only direct dependency.

## File map

| File | Owns |
|---|---|
| `main.go` | `ebiten.Game` impl. `Init()` hardcodes the world (2 cars, 2 bushes — becomes level data per issue #19). `Update()` maps keys to intent flags on the active car; Tab switches `ActiveCar`. `Layout` is fixed 640×480. |
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
