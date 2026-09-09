package game

import (
	"math"
	"math/rand/v2"
	"time"
)

const introMusicTicks = 4431   // ceil(48 bars * 4 beats * 60/156 * 60 updates)
const attractMusicTicks = 6144 // 64 bars at 150 BPM

func (g *game) updateAttractSequence(pressed bool) {
	if g.attract || g.fadePhase == fadeToAttract {
		if pressed && g.fadePhase != fadeOut {
			if g.fadePhase == fadeAttractIn {
				g.fadeTick = fadeDuration - g.fadeTick
			} else if g.fadePhase != fadeToAttract {
				g.fadeTick = 0
			}
			g.fadePhase = fadeOut
			return
		}
		if g.attract {
			g.attractTicks++
			ending := g.attractTicks >= attractMusicTicks-fadeDuration
			if g.sound != nil {
				ending = g.sound.previousScene == 4 && (g.sound.attractMusic.Position() >= g.sound.attractLength-time.Second || !g.sound.attractMusic.IsPlaying())
			}
			if ending && g.fadePhase == 0 {
				g.fadePhase = fadeOut
				g.fadeTick = 0
			}
		}
		return
	}
	if !g.started && (g.fadePhase == 0 || g.fadePhase == fadeIn) {
		if g.introTicks < introMusicTicks {
			if g.sound == nil {
				g.introTicks++
			} else if g.sound.previousScene == 0 && !g.sound.intro.IsPlaying() {
				g.introTicks = introMusicTicks
			}
		} else {
			g.introTicks++
			if g.introTicks >= introMusicTicks+600 && !pressed && g.fadePhase == 0 {
				g.fadePhase = fadeToAttract
				g.fadeTick = 0
			}
		}
	}
}
func (g *game) startAttract() {
	g.reset()
	g.attract = true
	g.wave = rand.IntN(8) // next spawn increments to a random wave in 1..8
	g.nextWave = 1
	g.weaponLevel = 1 + rand.IntN(3)
	g.weaponTime = 600
	g.inv = 0
	g.fadePhase = fadeAttractIn
}
func (g *game) autopilot() {
	desired := W/2 + math.Sin(float64(g.tick)*.013)*220
	if a := g.target(point{g.px, H - 75}, false); a != nil {
		desired = a.p.x + math.Sin(float64(g.tick)*.035)*18
	}
	for _, b := range g.bullets {
		if b.enemy && b.p.y > H-210 && math.Abs(b.p.x-g.px) < 45 {
			if g.px < W/2 {
				desired = g.px + 100
			} else {
				desired = g.px - 100
			}
			break
		}
	}
	g.px += math.Max(-4, math.Min(4, desired-g.px))
	// Cycle weapon demonstrations while keeping the unattended show alive.
	if g.tick%900 == 0 {
		g.weaponLevel = g.weaponLevel%3 + 1
	}
	g.weaponTime = 600
}
