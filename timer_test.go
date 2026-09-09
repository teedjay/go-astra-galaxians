package main

import "testing"

func TestTimedPickups(t *testing.T) {
	g := testGame()
	pick := func(p powerup) { p.p = point{g.px, H - 76}; g.powerups = []powerup{p}; g.updatePowerups() }
	pick(powerup{})
	if g.weaponLevel != 1 || g.weaponTime != 600 {
		t.Fatal("upgrade must grant ten seconds")
	}
	for i := 0; i < 120; i++ {
		g.updateWeaponTimer()
	}
	pick(powerup{star: true})
	if g.weaponLevel != 1 || g.weaponTime != 780 {
		t.Fatal("star must add five seconds without changing level")
	}
	found := false
	for _, e := range g.effects {
		if e.kind == dustEffect {
			found = true
		}
	}
	if !found {
		t.Fatal("star pickup must emit stardust")
	}
	pick(powerup{})
	if g.weaponLevel != 2 || g.weaponTime != 600 {
		t.Fatal("upgrade must reset timer")
	}
	pick(powerup{down: true})
	if g.weaponLevel != 1 || g.weaponTime != 600 {
		t.Fatal("downgrade must reset timer")
	}
	for i := 0; i < 599; i++ {
		g.updateWeaponTimer()
	}
	if g.weaponLevel != 1 || g.weaponTime != 1 {
		t.Fatal("weapon expired early")
	}
	g.updateWeaponTimer()
	if g.weaponLevel != 0 || g.weaponTime != 0 {
		t.Fatal("weapon did not expire to blaster")
	}
	pick(powerup{star: true})
	if g.weaponLevel != 0 || g.weaponTime != 300 {
		t.Fatal("star must not grant a weapon")
	}
	g.reset()
	if g.weaponTime != 0 {
		t.Fatal("restart retained timer")
	}
}
func TestTimerPauseAndSafeExpiration(t *testing.T) {
	g := testGame()
	g.weaponLevel = 3
	g.weaponTime = 1
	g.paused = true
	g.updateWeaponTimer()
	if g.weaponTime != 1 {
		t.Fatal("timer ran during pause")
	}
	g.paused = false
	g.over = true
	g.updateWeaponTimer()
	if g.weaponTime != 1 {
		t.Fatal("timer ran after game over")
	}
	g.over = false
	b := &g.beams[0]
	b.age = 100
	b.path = beamPath(point{g.px - 27, H - 96}, point{200, 150}, -1, 0)
	g.rockets = []rocket{{p: point{200, 500}}}
	g.pulses = []pulse{{p: point{300, 500}}}
	g.updateWeaponTimer()
	if g.weaponLevel != 0 || b.retractAge != 1 || len(b.path) == 0 || len(g.rockets) != 1 || len(g.pulses) != 1 {
		t.Fatal("expiry must preserve outgoing weapons and animate retraction")
	}
	for i := 0; i < 72; i++ {
		g.updateWeapons(false)
	}
	if len(b.path) != 0 {
		t.Fatal("expired beam did not retract")
	}
}

func TestMaxLevelPickupPreservesRemainingTime(t *testing.T) {
	for _, remaining := range []int{120, 600, 1200} {
		g := testGame()
		g.weaponLevel = 3
		g.weaponTime = remaining
		g.weaponCool = 9
		g.powerups = []powerup{{p: point{g.px, H - 76}}}
		g.updatePowerups()
		if g.weaponTime != remaining || g.weaponLevel != 3 || g.weaponCool != 9 {
			t.Fatal("max-level pickup changed active weapon or timer")
		}
	}
}
func TestPickupEligibilityByLevel(t *testing.T) {
	for level := 0; level <= 3; level++ {
		g := testGame()
		g.weaponLevel = level
		up, down, stars := 0, 0, 0
		for i := 0; i < 200; i++ {
			p := g.rollPickup(point{100, 200})
			switch {
			case p.down:
				down++
			case p.star:
				stars++
			default:
				up++
			}
		}
		if level == 0 && (down != 0 || stars != 0 || up != 200) {
			t.Fatal("base level should only generate upgrades")
		}
		if level == 3 && (up != 0 || down == 0 || stars == 0) {
			t.Fatal("max level should only generate downs and stars")
		}
		if (level == 1 || level == 2) && (up == 0 || down == 0 || stars == 0) {
			t.Fatal("intermediate levels should retain all pickups")
		}
	}
}

func TestWeaponTimerPausesForEntireWaveBreak(t *testing.T) {
	g := testGame()
	g.wave = 1
	g.weaponLevel = 3
	g.weaponTime = 1
	for i := 0; i < waveBreakTicks; i++ {
		g.updateWeaponTimer()
	}
	if g.weaponTime != 1 || g.weaponLevel != 3 {
		t.Fatal("weapon expired during interlude or extra wait")
	}
	g.powerups = []powerup{{p: point{g.px, H - 76}, star: true}}
	g.updatePowerups()
	g.updateWeaponTimer()
	if g.weaponTime != 301 {
		t.Fatal("pickup time should also be preserved during break")
	}
	g.wave++
	g.spawn = waveSize
	g.updateWeaponTimer()
	if g.weaponTime != 300 {
		t.Fatal("timer must resume when next wave starts spawning")
	}
	g.spawn = 0
	g.aliens = []*alien{{hp: 1}}
	g.updateWeaponTimer()
	if g.weaponTime != 299 {
		t.Fatal("timer must continue while aliens remain")
	}
}
