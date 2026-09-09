package game

import "github.com/hajimehoshi/ebiten/v2"

// Run starts the desktop game. Demo skips the intro for a playable session.
func Run(demo bool) error {
	ebiten.SetWindowSize(W, H)
	ebiten.SetWindowTitle("ASTRA GALAXIANS")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	g := newGame()
	var err error
	g.sound, err = newSoundSystem()
	if err != nil {
		return err
	}
	if demo {
		g.reset()
	}
	return ebiten.RunGame(g)
}
