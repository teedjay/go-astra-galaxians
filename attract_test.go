package main

import (
	"math"
	"testing"
)

func TestAttractWaitsForIntroMusicAndTenSeconds(t *testing.T) {
	g := testGame()
	for i := 0; i < introMusicTicks+599; i++ {
		g.updateAttractSequence(false)
	}
	if g.fadePhase != 0 || g.started {
		t.Fatal("attract started before music and ten-second wait")
	}
	g.updateAttractSequence(false)
	if g.fadePhase != fadeToAttract {
		t.Fatal("intro did not fade toward attract")
	}
	for i := 0; i < fadeDuration; i++ {
		g.updateTransition(false)
	}
	if !g.attract || !g.started || g.fadePhase != fadeAttractIn || g.wave < 0 || g.wave > 7 {
		t.Fatal("random attract level failed to start")
	}
}
func TestAttractEndsWithFadeAndCleanIntro(t *testing.T) {
	g := testGame()
	g.best = 2345
	g.startAttract()
	g.fadePhase = 0
	g.attractTicks = attractMusicTicks - fadeDuration - 1
	g.updateAttractSequence(false)
	if g.fadePhase != fadeOut {
		t.Fatal("attract must fade during final second")
	}
	for i := 0; i < fadeDuration; i++ {
		g.updateTransition(false)
	}
	if g.attract || g.started || g.fadePhase != fadeIn || g.best != 2345 || g.introTicks != 0 {
		t.Fatal("attract must return to clean intro with best intact")
	}
}
func TestAnyInputCancelsAttractTransitionsSmoothly(t *testing.T) {
	for _, phase := range []int{0, fadeAttractIn, fadeToAttract} {
		g := testGame()
		g.attract = phase != fadeToAttract
		g.started = g.attract
		g.fadePhase = phase
		if phase != 0 {
			g.fadeTick = 20
		}
		before := g.fadeAlpha()
		g.updateAttractSequence(true)
		if g.fadePhase != fadeOut || math.Abs(before-g.fadeAlpha()) > .00001 {
			t.Fatal("interrupt must fade without an alpha jump")
		}
		for i := 0; i < fadeDuration; i++ {
			g.updateTransition(false)
		}
		if g.attract || g.started {
			t.Fatal("interrupt did not return to intro")
		}
	}
}
func TestAutopilotMovesAndSurvives(t *testing.T) {
	g := testGame()
	g.startAttract()
	g.fadePhase = 0
	g.tick = 123
	g.aliens = []*alien{{p: point{100, 300}, hp: 1}}
	before := g.px
	g.autopilot()
	if g.px >= before || before-g.px > 4 {
		t.Fatal("autopilot did not steer toward enemy")
	}
	lives := g.lives
	g.hit()
	if g.lives != lives || g.over {
		t.Fatal("attract interrupted by player death")
	}
}
func TestAttractMusicDuration(t *testing.T) {
	pcm := makeAttractTune()
	ticks := float64(len(pcm)) / 4 / soundRate * 60
	if math.Abs(ticks-attractMusicTicks) > .01 {
		t.Fatal("attract timeline does not match song length")
	}
}
