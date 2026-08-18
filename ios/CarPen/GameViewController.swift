import UIKit
import Mobile

// The game's view controller. MobileEbitenViewController comes out of the
// generated framework and does all of the work — it owns the Metal view, the
// run loop and the touch handling, and hands touches to the Go side where
// scene/touch.go reads them.
//
// The module is called Mobile rather than CarPen because gomobile names it after
// the Go package it bound; the framework is named to match so that Xcode can
// find it (see ios/build.sh).
final class GameViewController: MobileEbitenViewController {
    // Errors from inside the game's own update reach the platform here. Ebiten
    // calls this on the main thread, and the default does nothing at all, which
    // on a device means the game quietly stops with no way to tell why.
    override func onError(onGameUpdate err: (any Error)!) {
        NSLog("CarPen stopped: %@", err?.localizedDescription ?? "unknown error")
    }

    // The race is drawn edge to edge; the status bar sits over the HUD strip and
    // the top row of touch controls.
    override var prefersStatusBarHidden: Bool { true }

    // The gesture that opens Control Centre and the app switcher runs along the
    // bottom of the screen, which on this layout is where the steering stick and
    // the pedals are. Asking for it to take a second swipe is the difference
    // between steering hard and being thrown out to the home screen.
    override var prefersHomeIndicatorAutoHidden: Bool { true }

    override var preferredScreenEdgesDeferringSystemGestures: UIRectEdge { .bottom }
}
