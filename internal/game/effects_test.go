package game

import "testing"

func TestEffectsExpireAndStayBounded(t *testing.T) {
	g := testGame()
	g.extremeExplosion(point{300, 200})
	counts := [3]int{}
	for _, e := range g.effects {
		counts[e.kind]++
	}
	for _, n := range counts {
		if n == 0 {
			t.Fatal("extreme explosion missing smoke, flames or sparks")
		}
	}
	for i := 0; i < 40; i++ {
		g.extremeExplosion(point{300, 200})
	}
	if len(g.effects) > maxEffects {
		t.Fatal("particle budget exceeded")
	}
	for i := 0; i < 150; i++ {
		g.updateEffects()
	}
	if len(g.effects) != 0 {
		t.Fatal("effects did not expire")
	}
}
func TestRocketExhaustAndLaserContactSparks(t *testing.T) {
	g := testGame()
	a := &alien{p: point{300, 200}, hp: 1}
	g.aliens = []*alien{a}
	g.rockets = []rocket{{p: point{300, 600}, angle: -1.57, speed: 1, target: a}}
	g.updateWeapons(false)
	g.updateWeapons(false)
	found := false
	for _, e := range g.effects {
		if e.kind == smokeEffect {
			found = true
		}
	}
	if !found {
		t.Fatal("rocket did not emit smoke")
	}
	g.rockets = nil
	g.effects = nil
	g.weaponLevel = 3
	b := &g.beams[0]
	b.age = 180
	b.reach = 2000
	b.aim = a.p
	b.target = a
	b.controls = beamControls(point{g.px - 27, H - 96}, a.p, -1, 0)
	for i := 0; i < 5; i++ {
		before := len(g.effects)
		g.updateWeapons(true)
		if b.dwell != i+1 || len(g.effects) < before+3 {
			t.Fatal("laser contact must spark every frame until destruction")
		}
	}
	g.updateWeapons(false)
	if b.dwell != 0 {
		t.Fatal("released beam retained contact glow")
	}
}
