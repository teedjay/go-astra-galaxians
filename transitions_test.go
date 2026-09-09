package main

import "testing"

func TestIntroFadeFromBlack(t *testing.T) {
	g := testGame()
	g.fadePhase = fadeIn
	if g.fadeAlpha() != 1 {
		t.Fatal("intro must start black")
	}
	previous := 1.
	for i := 0; i < fadeDuration; i++ {
		if !g.updateTransition(true) {
			t.Fatal("input must be blocked during fade")
		}
		a := g.fadeAlpha()
		if a > previous {
			t.Fatal("intro fade reversed")
		}
		previous = a
	}
	if g.fadeAlpha() != 0 || g.started || g.fadePhase != 0 {
		t.Fatal("intro must finish visible and wait for launch")
	}
}
func TestGameOverTimesOutAndReturnsToIntro(t *testing.T) {
	g := testGame()
	g.started = true
	g.over = true
	g.score = 1200
	g.best = 1200
	g.weaponLevel = 3
	for i := 0; i < gameOverDelay-1; i++ {
		g.updateTransition(false)
	}
	if g.fadePhase != 0 {
		t.Fatal("game over timed out early")
	}
	g.updateTransition(false)
	if g.fadePhase != fadeOut {
		t.Fatal("game over must fade after ten seconds")
	}
	previous := 0.
	for i := 0; i < fadeDuration; i++ {
		g.updateTransition(false)
		a := g.fadeAlpha()
		if a < previous {
			t.Fatal("out fade reversed")
		}
		previous = a
	}
	if g.fadeAlpha() != 1 || g.fadePhase != fadeIn || g.started || g.over || g.score != 0 || g.weaponLevel != 0 || g.best != 1200 {
		t.Fatal("scene must reset at black and preserve high score")
	}
	for i := 0; i < fadeDuration; i++ {
		g.updateTransition(true)
	}
	if g.started || g.fadeAlpha() != 0 {
		t.Fatal("transition input started a new game")
	}
}
func TestGameOverPressStartsFadeImmediately(t *testing.T) {
	g := testGame()
	g.started = true
	g.over = true
	g.updateTransition(true)
	if g.fadePhase != fadeOut || !g.over || g.score != 0 {
		t.Fatal("press should begin fade, not immediately restart")
	}
}
