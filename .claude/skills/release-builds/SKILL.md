---
name: release-builds
description: Cross-platform build matrix for CarPen — macOS, Windows, web/WASM bundle, Android/iOS via ebitenmobile, Xbox status. Use when producing distributable builds, adding a platform target, or extending CI's build jobs.
---

# Release builds

All commands verified from macOS (arm64). Output goes in `dist/` (gitignored — add it if not).

## Desktop

- **macOS** (native, needs Xcode CLT for cgo):
  `go build -o dist/carpen .`
- **Windows** (cross-compiles cleanly — Ebitengine's Windows backend is pure Go):
  `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags -H=windowsgui -o dist/carpen.exe .`
- **Linux**: needs cgo plus X11/GL headers — build on Linux (apt package list is in `.github/workflows/ci.yml`); do not cross-compile from macOS.

## Web (WASM)

```sh
mkdir -p dist/web
CGO_ENABLED=0 GOOS=js GOARCH=wasm go build -o dist/web/carpen.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/web/
```

Minimal `dist/web/index.html`:

```html
<!DOCTYPE html>
<script src="wasm_exec.js"></script>
<script>
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("carpen.wasm"), go.importObject)
    .then((r) => go.run(r.instance));
</script>
```

Serve over HTTP (e.g. `python3 -m http.server -d dist/web`) — `file://` will not load WASM.

## Mobile (Android / iOS)

`ebitenmobile bind` wraps a library package, not `package main` — that package is `./mobile`, which does in `init()` what `main.go` does in `main`.

```sh
go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
ebitenmobile bind -target ios -o dist/Mobile.xcframework ./mobile      # or ios/build.sh
ebitenmobile bind -target android -o dist/carpen.aar ./mobile
```

The iOS framework **must** be named `Mobile`: gomobile names the Swift module after the Go package it bound, and Xcode only finds a module inside a framework of the same name. Android needs the SDK/NDK; iOS needs Xcode.

### The iOS app

`ios/CarPen.xcodeproj` is the app shell around that framework — an `AppDelegate` and a `GameViewController` subclassing the generated `MobileEbitenViewController`, in landscape, with the status bar and home indicator hidden. Xcode cannot compile Go, so after changing anything under `carpen/` or `scene/`, run `ios/build.sh` and then build as usual.

Two settings are load-bearing and easy to lose:

- `OTHER_LDFLAGS = -framework GameController`. gomobile frameworks do not declare their system dependencies, and without this the link fails on `GCController` — the pad support has no other way in on iOS.
- The `Embed Frameworks` phase must carry `Mobile.xcframework` with `CodeSignOnCopy`, or the app builds and then dies at launch on a missing dylib.

Simulator builds need no signing. A device build needs a development team set on the target (`DEVELOPMENT_TEAM`), which needs an Apple ID — that is the only thing standing between the project and a device.

## Xbox

Ebitengine supports Xbox only through the Microsoft GDK, gated on ID@Xbox registration — there is no public toolchain. Tracked in epic #17; start the registration early, it has a long lead time.
