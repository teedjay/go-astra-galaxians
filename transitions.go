package main

const (
	fadeLaunch    = 3
	fadeIn        = 1
	fadeOut       = 2
	fadeDuration  = 60
	gameOverDelay = 10 * 60
)

// Only a fresh press can dismiss game over. Holding fire through death cannot
// bypass the result screen or automatically launch the next game.
func (g *game) updateTransition(pressed bool) bool {
	if g.fadePhase != 0 {
		g.fadeTick++
		if g.fadeTick >= fadeDuration {
			if g.fadePhase == fadeLaunch {
				g.reset()
			} else if g.fadePhase == fadeOut {
				g.reset()
				g.started = false
				g.fadePhase = fadeIn
			} else {
				g.fadePhase = 0
				g.fadeTick = 0
			}
		}
		return true
	}
	if g.over && !g.gameOverVisible() {
		return true
	}
	if g.over {
		g.gameOverTicks++
		if pressed || g.gameOverTicks >= gameOverDelay {
			g.fadePhase = fadeOut
			g.fadeTick = 0
		}
		return true
	}
	return false
}
func (g *game) fadeAlpha() float64 {
	t := float64(g.fadeTick) / fadeDuration
	t = t * t * (3 - 2*t)
	switch g.fadePhase {
	case fadeIn:
		return 1 - t
	case fadeOut, fadeLaunch:
		return t
	}
	return 0
}

func (g *game) gameOverVisible() bool {
	return g.over && (g.deathTicks == 0 || (g.deathTicks > 60 && len(g.fragments) == 0 && len(g.effects) == 0))
}
