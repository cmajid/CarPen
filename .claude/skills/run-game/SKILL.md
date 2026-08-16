---
name: run-game
description: Build, test, and launch CarPen to verify a change. Use before claiming any change works, or when asked to run or screenshot the game. Orders checks from cheapest to most expensive so no tokens or time are wasted on unnecessary steps.
---

# Verifying CarPen changes

Run the cheapest check that can catch the mistake. Do not launch the game window for pure logic changes — tests cover the `carpen` package headlessly.

## 1. Static checks and tests (always)

```sh
gofmt -l .        # any output fails CI — fix with gofmt -w
go vet ./...
go test ./...
```

Tests are display-free and run directly on macOS (CI wraps them in xvfb only because Linux Ebitengine needs a display at import time).

## 2. Cross-target compile check (when touching rendering, input, or platform code)

```sh
CGO_ENABLED=0 GOOS=js GOARCH=wasm go build ./...
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...
```

CI enforces both (see [.github/workflows/ci.yml](../../../.github/workflows/ci.yml)); catching a js/wasm break locally is much cheaper than a failed PR.

## 3. Launch the window (only for visual or game-feel verification)

```sh
go run .
```

Opens a 640×480 window titled "Car Pen" and blocks until closed — run it in the background and kill it when done.

- Arrow keys: accelerate / brake–reverse / steer the active car
- Tab: switch between the yellow and red car
- TPS counter is drawn top-left (should hold ~60)

There is no headless screenshot path; anything visual needs the real window.
