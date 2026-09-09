package game

import "testing"

func TestCatCrossesScreenWithoutGameplayEffects(t *testing.T) {
	g := testGame()
	g.cat.wait = 1
	g.updateCat()
	if !g.cat.active {
		t.Fatal("cat did not appear")
	}
	for i := 0; i < 700 && g.cat.active; i++ {
		g.updateCat()
	}
	if g.cat.active || g.cat.wait < 720 || g.lives != 3 || g.score != 0 || len(g.bullets) != 0 {
		t.Fatal("cat should leave safely and schedule next visit")
	}
	g.reset()
	if g.cat.active {
		t.Fatal("new game retained cat")
	}
}
func TestGameOverWaitsForDestruction(t *testing.T) {
	g := testGame()
	g.lives = 1
	g.hit()
	for i := 0; i < 60; i++ {
		g.updateTransition(true)
		g.updateEffects()
		g.updateShipExplosion()
		if g.gameOverVisible() || g.gameOverTicks != 0 || g.fadePhase != 0 {
			t.Fatal("game over interrupted explosion")
		}
	}
	for i := 0; i < 400 && !g.gameOverVisible(); i++ {
		g.updateEffects()
		g.updateShipExplosion()
	}
	if !g.gameOverVisible() {
		t.Fatal("game over never appeared after effects ended")
	}
	g.updateTransition(false)
	if g.gameOverTicks != 1 {
		t.Fatal("game-over timer must begin after destruction")
	}
}
