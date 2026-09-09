package game

import (
	"math"
	"math/rand"
	"testing"
)

func testGame() *game { return &game{rng: rand.New(rand.NewSource(7)), px: W / 2, lives: 3} }
func TestPickupProgressionAndReset(t *testing.T) {
	g := testGame()
	for want := 1; want <= 3; want++ {
		g.powerups = []powerup{{p: point{g.px, H - 76}}}
		g.updatePowerups()
		if g.weaponLevel != want || len(g.powerups) != 0 {
			t.Fatalf("pickup failed at level %d", want)
		}
	}
	g.powerups = []powerup{{p: point{g.px, H - 76}}}
	g.updatePowerups()
	if g.weaponLevel != 3 || g.weaponTime != 600 {
		t.Fatal("max-level pickup should refresh timer")
	}
	g.reset()
	if g.weaponLevel != 0 || len(g.powerups) != 0 || len(g.rockets) != 0 {
		t.Fatal("restart retained upgrades")
	}
}
func TestDropPityAndSingleKillReward(t *testing.T) {
	g := testGame()
	g.killsSinceDrop = 9
	a := &alien{p: point{300, 200}, hp: 1}
	g.destroy(a)
	g.destroy(a)
	if len(g.powerups) != 1 || g.score != 100 {
		t.Fatal("kill should reward once and guarantee a drop after ten kills")
	}
}
func TestRocketAccelerationAndRetarget(t *testing.T) {
	g := testGame()
	a := &alien{p: point{100, 170}, hp: 1}
	b := &alien{p: point{500, 180}, hp: 1}
	g.aliens = []*alien{a, b}
	g.weaponLevel = 1
	g.updateWeapons(true)
	if len(g.rockets) != 2 {
		t.Fatal("expected one rocket from each side")
	}
	speed := g.rockets[0].speed
	g.rockets[0].target = a
	a.hp = 0
	g.updateWeapons(false)
	if g.rockets[0].target != b || g.rockets[0].speed <= speed {
		t.Fatal("rocket must accelerate and retarget")
	}
	for i := 0; i < 300; i++ {
		g.updateWeapons(false)
	}
	if len(g.rockets) > 0 {
		t.Fatal("rockets must hit or expire")
	}
}
func TestFastPulseSweptCollision(t *testing.T) {
	g := testGame()
	a := &alien{p: point{200, 390}, hp: 1}
	g.aliens = []*alien{a}
	g.pulses = []pulse{{p: point{200, 405}}}
	g.updateWeapons(false)
	if a.hp != 0 || len(g.pulses) != 0 {
		t.Fatal("fast pulse skipped alien")
	}
}
func TestBeamLockDamageAndRelease(t *testing.T) {
	g := testGame()
	near := &alien{p: point{320, 400}, hp: 1}
	far := &alien{p: point{320, 150}, hp: 1}
	g.aliens = []*alien{far, near}
	g.weaponLevel = 3
	g.updateWeapons(true)
	if g.beams[0].target != near {
		t.Fatal("beam must select nearest alien")
	}
	for i := 0; i < 20; i++ {
		g.updateWeapons(true)
	}
	if near.hp == 0 {
		t.Fatal("beam damaged target before growing to it")
	}
	for i := 0; i < 240 && near.hp > 0; i++ {
		g.updateWeapons(true)
	}
	if near.hp != 0 {
		t.Fatal("beam must apply sustained damage")
	}
	g.updateWeapons(true)
	if g.beams[0].target != nil && g.beams[0].target == g.beams[1].target {
		t.Fatal("beams must not share a target after kill")
	}
	g.updateWeapons(false)
	for i := 0; i < 72; i++ {
		g.updateWeapons(false)
	}
	if len(g.beams[0].path) != 0 || g.beams[0].target != nil {
		t.Fatal("released trigger retained beam")
	}
}
func TestBeamPathHasLoopAndExactEndpoints(t *testing.T) {
	origin, target := point{320, 700}, point{200, 160}
	p := beamPath(origin, target, -1, 0)
	if p[0] != origin || p[len(p)-1] != target {
		t.Fatal("beam endpoints detached")
	}
	if math.Hypot(p[32].x-p[64].x, p[32].y-p[64].y) > 1e-8 {
		t.Fatal("beam does not cross itself")
	}
	if math.Hypot(p[48].x-p[32].x, p[48].y-p[32].y) < 20 {
		t.Fatal("loop collapsed")
	}
	for i := 1; i < len(p); i++ {
		if math.Hypot(p[i].x-p[i-1].x, p[i].y-p[i-1].y) > 30 {
			t.Fatal("beam is discontinuous")
		}
	}
}

func TestBeamGrowsFromCannonAndRestartsOnRelease(t *testing.T) {
	g := testGame()
	g.aliens = []*alien{{p: point{320, 180}, hp: 1}}
	g.weaponLevel = 3
	g.updateWeapons(true)
	b := &g.beams[0]
	origin := point{g.px - 27, H - 96}
	if b.path[0] != origin {
		t.Fatal("beam detached from cannon")
	}
	if d := math.Hypot(b.path[len(b.path)-1].x-origin.x, b.path[len(b.path)-1].y-origin.y); d > 7.01 {
		t.Fatal("beam appeared fully extended")
	}
	for i := 0; i < 50; i++ {
		g.updateWeapons(true)
	}
	if len(b.path) < 20 {
		t.Fatal("beam did not grow toward target")
	}
	g.px += 5
	g.updateWeapons(true)
	if b.path[0].x != g.px-27 {
		t.Fatal("beam source did not follow ship")
	}
	g.updateWeapons(false)
	for i := 0; i < 72; i++ {
		g.updateWeapons(false)
	}
	g.updateWeapons(true)
	if b.age != 1 || b.reach != 7 {
		t.Fatal("new trigger press did not restart growth")
	}
}
func TestBeamRetargetSmoothlyWithoutEarlyDamage(t *testing.T) {
	g := testGame()
	g.weaponLevel = 3
	old := &alien{p: point{110, 200}, hp: 1}
	next := &alien{p: point{520, 220}, hp: 1}
	g.aliens = []*alien{old}
	// Grow without damage by advancing geometry directly.
	b := &g.beams[0]
	origin := point{g.px - 27, H - 96}
	b.aim = old.p
	b.target = old
	for j := range b.controls {
		b.controls[j] = mix(origin, old.p, float64(j)/9)
	}
	for i := 0; i < 180; i++ {
		b.advance(origin, -1)
	}
	previous := b.controls
	tip := b.path[len(b.path)-1]
	old.hp = 0
	g.aliens = append(g.aliens, next)
	g.updateWeapons(true)
	if b.target != next {
		t.Fatal("did not acquire replacement target")
	}
	if b.age != 181 {
		t.Fatal("retarget reset beam growth")
	}
	for i, p := range b.controls {
		if math.Hypot(p.x-previous[i].x, p.y-previous[i].y) > 12.001 {
			t.Fatal("control point snapped on retarget")
		}
	}
	newTip := b.path[len(b.path)-1]
	if math.Hypot(newTip.x-tip.x, newTip.y-tip.y) > 12.001 {
		t.Fatal("tip teleported")
	}
	if b.dwell != 0 || next.hp == 0 {
		t.Fatal("damage applied before visible beam reached replacement")
	}
	for i := 0; i < 240 && next.hp > 0; i++ {
		g.updateWeapons(true)
	}
	if next.hp != 0 {
		t.Fatal("beam failed to reach replacement target")
	}
}

func TestBeamsReserveDifferentTargets(t *testing.T) {
	g := testGame()
	g.weaponLevel = 3
	a := &alien{p: point{320, 350}, hp: 1}
	b := &alien{p: point{320, 200}, hp: 1}
	c := &alien{p: point{400, 170}, hp: 1}
	g.aliens = []*alien{a, b, c}
	g.updateWeapons(true)
	if g.beams[0].target == nil || g.beams[1].target == nil || g.beams[0].target == g.beams[1].target {
		t.Fatal("beams did not select separate targets")
	}
	a.hp = 0
	g.updateWeapons(true)
	if g.beams[0].target != c || g.beams[1].target != b {
		t.Fatal("retarget stole the other beam's lock")
	}
	c.hp = 0
	g.updateWeapons(true)
	if g.beams[0].target != nil || g.beams[1].target != b {
		t.Fatal("single remaining alien should have only one beam assigned")
	}
}
func TestLivingKnotFollowsDistinctAnchorsSmoothly(t *testing.T) {
	g := testGame()
	origin := point{293, 704}
	for i := 0; i < 5; i++ {
		g.aliens = append(g.aliens, &alien{p: point{100 + float64(i)*90, 200 + float64(i%2)*70}, hp: 1})
	}
	b := &beam{target: g.aliens[0], aim: g.aliens[0].p, age: 180}
	g.assignAnchors(b, origin, -1)
	seen := map[*alien]bool{b.target: true}
	for _, a := range b.anchors {
		if a == nil || seen[a] {
			t.Fatal("anchors must use distinct aliens")
		}
		seen[a] = true
	}
	before := b.livingControls(origin, -1)
	b.anchors[0].p.x += 80
	b.anchors[1].p.y += 90
	after := b.livingControls(origin, -1)
	if before[3] == after[3] || before[4] == after[4] {
		t.Fatal("knot ignored anchor motion")
	}
	if after[3] != after[6] {
		t.Fatal("knot crossing split apart")
	}
	b.controls = before
	b.anchors[1].hp = 0
	g.assignAnchors(b, origin, -1)
	if !visible(b.anchors[1]) {
		t.Fatal("dead anchor was not replaced")
	}
	b.advance(origin, -1)
	for j := 1; j < 10; j++ {
		if math.Hypot(b.controls[j].x-before[j].x, b.controls[j].y-before[j].y) > 12.001 {
			t.Fatal("anchor replacement snapped curve")
		}
	}
}

func TestBeamUnwindsAndRetractsIntoMovingCannon(t *testing.T) {
	g := testGame()
	g.weaponLevel = 3
	origin := point{g.px - 27, H - 96}
	b := &g.beams[0]
	b.age = 150
	b.path = beamPath(origin, point{150, 180}, -1, 0)
	initial := append([]point(nil), b.path...)
	g.updateWeapons(false)
	if len(b.path) != len(initial) || b.target != nil || b.retractAge != 1 {
		t.Fatal("release must animate without retaining damage lock")
	}
	tip := b.path[len(b.path)-1]
	oldTip := initial[len(initial)-1]
	if math.Hypot(tip.x-oldTip.x, tip.y-oldTip.y) > 2 {
		t.Fatal("release snapped beam")
	}
	for i := 1; i < 36; i++ {
		g.px += 1
		g.updateWeapons(false)
	}
	for _, p := range b.path {
		if math.Abs(p.x-(g.px-27)) > 1e-8 {
			t.Fatal("beam did not straighten above moving cannon")
		}
	}
	previous := math.Inf(1)
	for i := 36; i < 72; i++ {
		g.updateWeapons(false)
		if len(b.path) > 0 {
			tip = b.path[len(b.path)-1]
			length := H - 96 - tip.y
			if length > previous {
				t.Fatal("retracting beam grew")
			}
			previous = length
		}
	}
	if len(b.path) != 0 || b.age != 0 {
		t.Fatal("beam did not disappear inside cannon")
	}
}
func TestMissingTargetRetractsOnlyUnassignedBeam(t *testing.T) {
	g := testGame()
	g.weaponLevel = 3
	a := &alien{p: point{150, 180}, hp: 1}
	b := &alien{p: point{500, 180}, hp: 1}
	g.aliens = []*alien{a, b}
	for i := 0; i < 35; i++ {
		g.updateWeapons(true)
	}
	lost := g.beams[0].target
	lost.hp = 0
	g.updateWeapons(true)
	if g.beams[0].retractAge != 1 || len(g.beams[0].path) == 0 {
		t.Fatal("unassigned beam must retract visibly")
	}
	if g.beams[1].target == nil || g.beams[1].retractAge != 0 {
		t.Fatal("remaining valid beam should continue firing")
	}
	empty := testGame()
	empty.weaponLevel = 3
	for i := 0; i < 100; i++ {
		empty.updateWeapons(true)
	}
	if len(empty.beams[0].path) > 0 || len(empty.beams[1].path) > 0 {
		t.Fatal("beams grew without targets")
	}
}
