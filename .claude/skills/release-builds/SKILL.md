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

Uses `ebitenmobile bind`, which wraps a library package, not `package main`:

```sh
go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
ebitenmobile bind -target android -o dist/carpen.aar ./mobile
ebitenmobile bind -target ios -o dist/Carpen.xcframework ./mobile
```

The `./mobile` package (exposing the game via `github.com/hajimehoshi/ebiten/v2/mobile.SetGame`) does not exist yet — it is Phase 3 of epic #17. Android needs the SDK/NDK; iOS needs Xcode.

## Xbox

Ebitengine supports Xbox only through the Microsoft GDK, gated on ID@Xbox registration — there is no public toolchain. Tracked in epic #17; start the registration early, it has a long lead time.
