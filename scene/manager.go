package scene

import "github.com/hajimehoshi/ebiten/v2"

// Manager runs one scene at a time. It is the ebiten.Game the window drives, and
// the only part of the game Ebiten knows about.
type Manager struct {
	current Scene
	width,
	height int
}

var _ ebiten.Game = (*Manager)(nil)

// NewManager starts the game on first, in a screen width by height pixels.
func NewManager(width, height int, first Scene) *Manager {
	return &Manager{current: first, width: width, height: height}
}

// Update ticks the running scene and hands over to whatever scene it returns.
// The scene taking over is first updated on the next tick, so the key press that
// ended one scene is never read a second time by the one that follows it.
func (m *Manager) Update() error {
	next, err := m.current.Update()
	if err != nil {
		return err
	}
	if next != nil {
		m.current = next
	}

	return nil
}

func (m *Manager) Draw(screen *ebiten.Image) {
	m.current.Draw(screen)
}

// Layout holds the game at a fixed size, whatever the window is doing.
func (m *Manager) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return m.width, m.height
}
