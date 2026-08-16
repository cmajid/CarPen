---
name: resolv-collision
description: How CarPen does SAT/OBB collision with the resolv library — the wrapper API, the rotation-sign trap, and resolv's containment pitfall. Read before touching collision code, adding anything collidable, or upgrading resolv.
---

# Collision with resolv

CarPen's collision detection is the resolv library (`github.com/solarlune/resolv` v0.8.1 — pure Go, wasm-safe, detection-only) behind the wrapper in `carpen/collision.go`. **That file is the only place resolv may be imported.** Everything else speaks `carpen.OBB` in the game's own convention: degrees, clockwise, Y growing downward. resolv's full physics cousins (jakecoffman/cp, box2d ports) were considered and rejected: CarPen has its own kinematic car model in `carpen/car.go`, and a rigid-body engine would fight it.

## The wrapper API

| Call | Use |
|---|---|
| `NewOBB(cx, cy, w, h, rotationDeg)` | Build a plain box **once** per entity. |
| `NewBeveledOBB(cx, cy, w, h, bevel, rotationDeg)` | A box with 45° cut corners — the octagon a rounded rectangle flattens to. The car uses this. |
| `(*OBB).SetTransform(cx, cy, rotationDeg)` | Re-place it every tick — mutates in place, no allocation. |
| `Intersects(a, b *OBB) bool` | The SAT test, generic over any convex outline. Strict: edge-touching is not a hit. |
| `NewCircle(cx, cy, r)` / `(*Circle).SetPosition` | Round collider — the bushes, whose square boxes crashed cars into empty corners. |
| `(*Circle).IntersectsOBB(o *OBB) bool` | Circle-vs-polygon SAT: the polygon's edge normals **plus the centre-to-nearest-vertex axis** (without it, corners over-report — the exact bug circles were added to fix). |
| `(*OBB).Outline() []Vector` (4 or 8 points), `Center()`; `(*Circle).Center()`, `Radius()` | World-space geometry for the F3 overlay and tests. |
| `CollisionEvent{Obstruction}` | What detection raises; `ObstructionWall`, `ObstructionBush`, `ObstructionCar`. |

Entities own their shapes: `Car.OBB()` (beveled) and `Bush.Collider()` (circle) derive centre/rotation from live state plus sprite size (kept as plain numbers so tests need no image) and return the same re-placed collider every call. Follow that pattern for anything new that collides.

## Game-feel tuning constants

Collision shapes are deliberately a little smaller than the sprites, so a visible near miss is a miss. All hand-tuned, all in one place each: `carBodyInset` (4) and `carCornerBevel` (16) in `carpen/car.go`, `bushInset` (4) in `carpen/bush.go`. Scene tests measure the car's `Outline()` rather than hardcoding sprite arithmetic, so retuning these does not break them. Performance is irrelevant at this object count — never trade shape accuracy for it here.

## Two resolv traps the wrapper guards

1. **Rotation sign.** resolv's `SetRotation` takes radians **counter-clockwise** (the math.Atan2 convention); the lot turns clockwise in degrees because Y grows downward. `SetTransform` flips the sign, and nothing else may call resolv's rotation API with a game angle.
2. **`IsIntersecting` misses containment.** resolv v0.8's convex-convex test looks for *crossing edges*, so a box wholly inside another reports no intersection. `Intersects` therefore runs SAT projections itself over resolv's public `SATAxes()` / `Project()` / `IsOverlapping()`. Do not "simplify" it back to `shape.IsIntersecting`.

Both traps are pinned by `carpen/collision_test.go` (cases `clockwise reaches down-right` / `clockwise misses up-right` and `contained`). After any resolv upgrade, re-read `Transformed()` (the `p.Rotate(-cp.rotation)` line) and `convexConvexTest` in the new version, then rely on those tests to catch drift.

## Rules of the seam

- **Detection never decides consequences.** Collision code raises `CollisionEvent`; whatever is plugged into `Gameplay.OnCollision` decides what a crash means. No game-over logic near detection.
- **Events fire on the rising edge only**, and the tracker is seeded with the spawn state — level-01 legitimately spawns the car overlapping the bottom wall, and that is not a crash the player made.
- **Walls** are four 100px-thick OBBs pressed against the outside of the lot (`lotWalls` in `scene/gameplay.go`). Thickness must stay far above `MaxSpeed` (6 px/tick) or a fast car could step across one between checks.
- **F3** toggles the debug overlay in gameplay (hinted on the HUD): every collider outlined (octagons for cars, circles for bushes, boxes for walls), the active car's shape **red for exactly the ticks it overlaps**, and a bottom strip with the live pivot/rotation/speed and `touching wall|bush|car` vs `clear` (`Gameplay.touching`). The crash notice under the HUD is different on purpose: it is the *first* crash remembered, the overlay strip is the truth of *now* — a "crashed while merely close" report is usually the 219px car's far corner clipping something the player wasn't watching. Only the active car is checked against obstacles (parked things don't collide with each other).
- resolv's `Space`/broad-phase is unused on purpose at this object count. Shapes are never added to a `Space`; resolv's `SetPosition` is nil-space safe, so that is fine.
