package main

import "testing"

func TestPowerdownsLowerOneLevelAndClamp(t *testing.T) {
	g := testGame()
	g.weaponLevel = 3
	for want := 2; want >= 0; want-- {
		g.powerups = []powerup{{p: point{g.px, H - 76}, down: true}}
		g.updatePowerups()
		if g.weaponLevel != want || len(g.powerups) != 0 || !g.downgradeFlash {
			t.Fatal("powerdown did not lower one level")
		}
	}
	g.powerups = []powerup{{p: point{g.px, H - 76}, down: true}}
	g.updatePowerups()
	if g.weaponLevel != 0 || g.score != 0 {
		t.Fatal("base-level powerdown changed score or underflowed")
	}
	g.powerups = []powerup{{p: point{g.px, H - 76}}}
	g.updatePowerups()
	if g.weaponLevel != 1 || g.downgradeFlash {
		t.Fatal("upgrade after downgrade failed")
	}
}
func TestDowngradeLetsExistingWeaponsFinish(t *testing.T) {
	g := testGame()
	g.weaponLevel = 3
	a := &alien{p: point{300, 200}, hp: 1}
	g.aliens = []*alien{a}
	g.rockets = []rocket{{p: point{200, 600}, angle: -1.57, speed: 1, target: a}}
	g.pulses = []pulse{{p: point{500, 600}}}
	b := &g.beams[0]
	b.age = 160
	b.target = a
	b.dwell = 10
	b.path = beamPath(point{g.px - 27, H - 96}, a.p, -1, 0)
	g.powerups = []powerup{{p: point{g.px, H - 76}, down: true}}
	g.updatePowerups()
	if len(g.rockets) != 1 || len(g.pulses) != 1 || len(b.path) == 0 || b.retractAge != 1 || b.target != nil || b.dwell != 0 {
		t.Fatal("downgrade deleted projectiles or failed to retract beam safely")
	}
	oldRocket := g.rockets[0].p
	oldPulse := g.pulses[0].p
	g.updateWeapons(false)
	if g.rockets[0].p == oldRocket || g.pulses[0].p == oldPulse {
		t.Fatal("old projectiles stopped moving")
	}
	for i := 0; i < 300; i++ {
		g.updateWeapons(false)
	}
	if len(g.rockets) > 0 || len(g.pulses) > 0 || len(b.path) > 0 {
		t.Fatal("old weapons did not finish normally")
	}
}
func TestDropsIncludeBothKinds(t *testing.T) {
	g := testGame()
	g.weaponLevel = 1
	down, up := 0, 0
	for i := 0; i < 100; i++ {
		g.killsSinceDrop = 9
		g.destroy(&alien{p: point{320, 200}, hp: 1})
	}
	for _, p := range g.powerups {
		if p.down {
			down++
		} else {
			up++
		}
	}
	if down == 0 || up == 0 || down >= up {
		t.Fatalf("expected occasional powerdowns: %d down, %d up", down, up)
	}
}
