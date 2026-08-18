import UIKit

// The iOS app is a shell around the game and nothing more. Everything the
// player sees is drawn by Go: ebitenmobile binds the mobile package as a
// framework, and the view controller below is the platform's window onto it.
//
// There is no storyboard on purpose. One view controller filling one window is
// the whole interface, and saying that in six lines here is clearer than saying
// it in a nib nobody will open twice.
@main
final class AppDelegate: UIResponder, UIApplicationDelegate {
    var window: UIWindow?

    private var game: GameViewController? {
        window?.rootViewController as? GameViewController
    }

    func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
    ) -> Bool {
        let window = UIWindow(frame: UIScreen.main.bounds)
        window.rootViewController = GameViewController()
        window.makeKeyAndVisible()
        self.window = window
        return true
    }

    // The game is stopped while the app is not the one in front. Ebiten drives
    // its own loop, which iOS does not pause for us; left running it would go on
    // ticking the race in the background, draining the battery and — worse for a
    // game about parking — carrying the car on while nobody is watching.
    func applicationWillResignActive(_ application: UIApplication) {
        game?.suspendGame()
    }

    func applicationDidBecomeActive(_ application: UIApplication) {
        game?.resumeGame()
    }
}
