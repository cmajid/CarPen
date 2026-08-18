#!/bin/sh
# Rebuilds the Go half of the app: everything the player sees lives in the
# mobile package, and Xcode cannot compile Go. Run this after changing anything
# under carpen/ or scene/, then build in Xcode as usual — the project links
# whatever this last left in dist/.
#
# The framework has to be called Mobile because gomobile names the Swift module
# after the Go package it bound, and Xcode will only find a module in a
# framework of the same name.
set -eu

cd "$(dirname "$0")/.."

if ! command -v ebitenmobile >/dev/null 2>&1; then
	echo "installing ebitenmobile" >&2
	go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
	PATH="$PATH:$(go env GOPATH)/bin"
	export PATH
fi

echo "binding ./mobile — this takes a couple of minutes" >&2
ebitenmobile bind -target ios -o dist/Mobile.xcframework ./mobile
echo "built dist/Mobile.xcframework" >&2
