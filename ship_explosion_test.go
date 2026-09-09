package main

import (
	"image/color"
	"testing"
)

func TestShatterPreservesOpaquePixels(t *testing.T) {
	g := testGame()
	pixels := []byte{80, 180, 255, 255, 0, 0, 0, 0, 255, 200, 90, 255, 230, 253, 255, 255}
	origin := point{320, 400}
	g.shatterPixels(pixels, 2, 2, origin)
	if len(g.fragments) != 3 {
		t.Fatal("must create one fragment per opaque pixel")
	}
	f := g.fragments[0]
	if f.p != (point{318.5, 398.5}) || f.c != (color.RGBA{80, 180, 255, 255}) {
		t.Fatal("fragment lost source position or color")
	}
	old := f.velocity.y
	g.updateShipExplosion()
	if g.fragments[0].velocity.y <= old || g.fragments[0].life >= f.life {
		t.Fatal("fragment needs gravity and fading lifetime")
	}
	for i := 0; i < 400; i++ {
		g.updateShipExplosion()
	}
	if len(g.fragments) != 0 {
		t.Fatal("debris did not expire or leave screen")
	}
}
func TestFinalLifeCreatesInfernoOnlyOnce(t *testing.T) {
	g := testGame()
	g.lives = 2
	g.hit()
	if g.deathTicks != 0 || g.over {
		t.Fatal("nonfinal hit triggered ship destruction")
	}
	g.hit()
	if !g.over || g.deathTicks != 1 || len(g.effects) < 300 {
		t.Fatal("final hit missing inferno")
	}
	count := len(g.effects)
	g.hit()
	if g.lives != 0 || len(g.effects) != count {
		t.Fatal("final explosion was retriggered")
	}
	for i := 0; i < 240; i++ {
		g.updateEffects()
		g.updateShipExplosion()
	}
	if len(g.effects) != 0 {
		t.Fatal("inferno effects leaked after game over")
	}
	g.reset()
	if g.deathTicks != 0 || len(g.fragments) != 0 {
		t.Fatal("restart retained destruction")
	}
}
func TestDebrisOffscreenCleanup(t *testing.T) {
	g := testGame()
	g.fragments = []shipFragment{{p: point{W + 10, 200}, life: 100, total: 100}, {p: point{320, H + 10}, life: 100, total: 100}}
	g.updateShipExplosion()
	if len(g.fragments) != 0 {
		t.Fatal("offscreen fragments were retained")
	}
}
