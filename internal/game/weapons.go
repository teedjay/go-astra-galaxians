package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var weaponNames = [...]string{"BLASTER", "SEEKER ROCKETS", "PULSE LASERS", "KNOT BEAMS"}

var powerdownColor = color.RGBA{255, 76, 86, 255}

type powerup struct {
	star bool
	down bool
	p    point
	age  int
}
type rocket struct {
	p                   point
	angle, speed, phase float64
	age                 int
	target              *alien
	trail               []point
}
type pulse struct {
	p   point
	age int
}
type beam struct {
	target        *alien
	dwell         int
	path          []point
	age           int
	reach         float64
	controls      [10]point
	anchors       [3]*alien
	retractAge    int
	retractPath   []point
	retractLength float64
	aim           point
}

func visible(a *alien) bool {
	return a != nil && a.hp > 0 && a.p.x > 10 && a.p.x < W-10 && a.p.y > 85 && a.p.y < H-105
}
func (g *game) target(from point, random bool) *alien {
	return g.targetExcept(from, random)
}
func (g *game) targetExcept(from point, random bool, excluded ...*alien) *alien {
	var chosen *alien
	count := 0
	best := math.Inf(1)
	for _, a := range g.aliens {
		if !visible(a) {
			continue
		}
		skip := false
		for _, other := range excluded {
			if a == other {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		count++
		d := math.Hypot(a.p.x-from.x, a.p.y-from.y)
		if random {
			if g.rng.Intn(count) == 0 {
				chosen = a
			}
		} else if d < best {
			chosen = a
			best = d
		}
	}
	return chosen
}
func (g *game) destroy(a *alien) {
	if a.hp <= 0 {
		return
	}
	a.hp = 0
	g.score += 100
	if a.mode == 2 {
		g.score += 150
	}
	g.emit(a.p, palettes[a.kind], 22)
	if g.rng.Float64() < .12 {
		g.extremeExplosion(a.p)
		g.sfx("huge")
	} else {
		g.sfx("explosion")
	}
	g.killsSinceDrop++
	// Random drops with a pity limit keep upgrades reachable even in unlucky runs.
	if a.p.y > 80 && a.p.y < H-100 && (g.rng.Float64() < .13 || g.killsSinceDrop >= 10) {
		g.powerups = append(g.powerups, g.rollPickup(a.p))
		g.killsSinceDrop = 0
	}
}

// Renormalize eligible types instead of creating useless pickups.
func (g *game) rollPickup(p point) powerup {
	if g.weaponLevel == 0 {
		return powerup{p: p}
	}
	roll := g.rng.Float64()
	if g.weaponLevel == 3 {
		return powerup{p: p, down: roll < .4, star: roll >= .4}
	}
	return powerup{p: p, down: roll < .2, star: roll >= .7}
}
func (g *game) updatePowerups() {
	if g.upgradeFlash > 0 {
		g.upgradeFlash--
	}
	kept := g.powerups[:0]
	for _, p := range g.powerups {
		p.age++
		p.p.y += 1.7
		if !g.over && math.Hypot(p.p.x-g.px, p.p.y-(H-75)) < 30 {
			if p.star {
				g.weaponTime += 5 * 60
				g.pickupMessage = "+5 SECONDS"
				g.downgradeFlash = false
				g.upgradeFlash = 120
				g.starDust(p.p)
				g.sfx("star")
			} else if !p.down && g.weaponLevel == 3 {
				g.pickupMessage = "MAX POWER"
				g.downgradeFlash = false
				g.upgradeFlash = 90
				g.emit(p.p, palettes[0], 20)
			} else {
				level := min(3, g.weaponLevel+1)
				if p.down {
					level = max(0, g.weaponLevel-1)
				}
				g.setWeaponLevel(level, p.down)
				if p.down {
					g.sfx("down")
				} else {
					g.sfx("up")
				}
				g.weaponTime = 10 * 60
				c := palettes[0]
				if p.down {
					c = powerdownColor
				}
				g.emit(p.p, c, 35)
			}
			continue
		}
		if p.p.y < H+20 {
			kept = append(kept, p)
		}
	}
	g.powerups = kept
}

// All level changes preserve fired projectiles and let beams retract normally.
func (g *game) setWeaponLevel(level int, down bool) {
	previous := g.weaponLevel
	g.weaponLevel = level
	g.weaponCool = 0
	g.upgradeFlash = 150
	g.downgradeFlash = down
	g.pickupMessage = ""
	if level < previous {
		for i := range g.beams {
			if g.beams[i].retractAge == 0 {
				g.beams[i].retract(point{g.px + float64(i*2-1)*27, H - 96})
			}
		}
	}
}
func (g *game) updateWeaponTimer() {
	if g.paused || g.over || g.weaponTime <= 0 || (g.wave > 0 && len(g.aliens) == 0 && g.spawn == 0) {
		return
	}
	g.weaponTime--
	if g.weaponTime == 0 && g.weaponLevel > 0 {
		g.setWeaponLevel(0, true)
		g.pickupMessage = "POWER EXPIRED"
	}
}
func segmentDistance(p, a, b point) float64 {
	dx, dy := b.x-a.x, b.y-a.y
	den := dx*dx + dy*dy
	if den == 0 {
		return math.Hypot(p.x-a.x, p.y-a.y)
	}
	t := math.Max(0, math.Min(1, ((p.x-a.x)*dx+(p.y-a.y)*dy)/den))
	return math.Hypot(p.x-a.x-t*dx, p.y-a.y-t*dy)
}
func (g *game) updateWeapons(firing bool) {
	if g.weaponCool > 0 {
		g.weaponCool--
	}
	if firing && g.weaponCool == 0 {
		switch g.weaponLevel {
		case 1:
			before := len(g.rockets)
			for _, side := range []float64{-1, 1} {
				p := point{g.px + side*27, H - 89}
				if a := g.target(p, true); a != nil {
					g.rockets = append(g.rockets, rocket{p: p, angle: -math.Pi/2 + side*.65, speed: 1.1, phase: g.rng.Float64() * 2 * math.Pi, target: a})
				}
			}
			if len(g.rockets) > before {
				g.sfx("rocket")
			}
			g.weaponCool = 44
		case 2:
			g.sfx("pulse")
			for _, side := range []float64{-1, 1} {
				g.pulses = append(g.pulses, pulse{p: point{g.px + side*27, H - 94}})
			}
			g.weaponCool = 10
		}
	}
	rs := g.rockets[:0]
	for _, r := range g.rockets {
		r.age++
		if !visible(r.target) {
			r.target = g.target(r.p, true)
		}
		prev := r.p
		if r.target != nil {
			desired := math.Atan2(r.target.p.y-r.p.y, r.target.p.x-r.p.x)
			wobble := math.Sin(float64(r.age)*.14+r.phase) * .24
			turn := math.Min(.15, .035+float64(r.age)*.0015)
			r.angle += math.Max(-turn, math.Min(turn, angleDelta(r.angle, desired+wobble)))
		}
		r.speed = math.Min(9, r.speed+.105)
		r.p.x += math.Cos(r.angle) * r.speed
		r.p.y += math.Sin(r.angle) * r.speed
		if r.age%2 == 0 {
			g.rocketSmoke(r.p, r.angle, r.speed)
		}
		r.trail = append(r.trail, prev)
		if len(r.trail) > 18 {
			r.trail = r.trail[1:]
		}
		hit := false
		for _, a := range g.aliens {
			if a.hp > 0 && segmentDistance(a.p, prev, r.p) < 18 {
				g.destroy(a)
				g.rocketBurst(r.p)
				hit = true
				break
			}
		}
		if !hit {
			if r.age > 240 || r.p.x < -60 || r.p.x > W+60 || r.p.y < -80 || r.p.y > H+80 {
				g.rocketBurst(r.p)
			} else {
				rs = append(rs, r)
			}
		}
	}
	g.rockets = rs
	ps := g.pulses[:0]
	for _, p := range g.pulses {
		prev := p.p
		p.p.y -= 23
		p.age++
		hit := false
		for _, a := range g.aliens {
			if a.hp > 0 && segmentDistance(a.p, prev, p.p) < 18 {
				g.destroy(a)
				g.emit(a.p, palettes[0], 14)
				hit = true
				break
			}
		}
		if !hit && p.p.y > 60 {
			ps = append(ps, p)
		}
	}
	g.pulses = ps
	for i := range g.beams {
		b := &g.beams[i]
		side := float64(i*2 - 1)
		origin := point{g.px + side*27, H - 96}
		if !firing || g.weaponLevel != 3 || b.retractAge > 0 {
			b.retract(origin)
			continue
		}
		// Keep each lock until the victim dies, then acquire the nearest living alien.
		other := g.beams[1-i].target
		if !visible(b.target) || b.target == other {
			b.target = g.targetExcept(origin, false, other)
			b.dwell = 0
		}
		if b.target == nil {
			b.retract(origin)
			continue
		}
		if b.age == 0 {
			b.aim = point{origin.x, origin.y - 450}
			if b.target != nil {
				b.aim = b.target.p
			}
			for j := range b.controls {
				b.controls[j] = mix(origin, b.aim, float64(j)/9)
			}
		}
		if b.target != nil {
			b.aim = b.target.p
		}
		g.assignAnchors(b, origin, side)
		b.advance(origin, side)
		tip := b.path[len(b.path)-1]
		// Damage and impact sparks follow the visible tip, including during retargeting.
		if b.target != nil && math.Hypot(tip.x-b.target.p.x, tip.y-b.target.p.y) < 16 {
			b.dwell++
			g.laserSparks(tip)
			if b.dwell >= 22 {
				g.destroy(b.target)
				b.dwell = 0
			}
		} else {
			b.dwell = 0
		}

	}
}
func (g *game) rocketBurst(p point) {
	g.sfx("explosion")
	g.emit(p, palettes[2], 35)
	g.emit(p, white, 16)
	for i := 0; i < 24; i++ {
		a := float64(i) * 2 * math.Pi / 24
		g.particles = append(g.particles, particle{p, math.Cos(a) * 5, math.Sin(a) * 5, 28, palettes[1]})
	}
}

// Each group of three new control points extends the previous cubic segment.
// Keeping these points alive across locks prevents the curve from jumping.
func beamControls(origin, target point, side, phase float64) [10]point {
	dx, dy := target.x-origin.x, target.y-origin.y
	length := math.Max(1, math.Hypot(dx, dy))
	u := point{dx / length, dy / length}
	v := point{-u.y * side, u.x * side}
	at := func(p point, along, across float64) point {
		return point{p.x + u.x*along + v.x*across, p.y + u.y*along + v.y*across}
	}
	cross := mix(origin, target, .52)
	radius := math.Min(80, length*.25) * (1 + .12*math.Sin(phase))
	handle := math.Min(90, length*.28)
	return [10]point{origin, at(origin, handle, 0), at(cross, 0, -radius), cross,
		at(cross, 0, radius*2.7), at(cross, -radius*2, 0), cross,
		at(cross, handle, 0), at(target, -handle, 0), target}
}

// Each beam keeps separate alien anchors for the crossing and the two loop
// handles. Retain live assignments so formation motion does not shuffle them.
func (g *game) assignAnchors(b *beam, origin point, side float64) {
	base := beamControls(origin, b.aim, side, float64(b.age)*.025)
	used := []*alien{b.target}
	for i, index := range []int{3, 4, 5} {
		a := b.anchors[i]
		valid := visible(a)
		for _, other := range used {
			if a == other {
				valid = false
			}
		}
		if !valid {
			a = g.targetExcept(base[index], false, used...)
			b.anchors[i] = a
		}
		if a != nil {
			used = append(used, a)
		}
	}
}
func (b *beam) livingControls(origin point, side float64) [10]point {
	c := beamControls(origin, b.aim, side, float64(b.age)*.025)
	if visible(b.anchors[0]) {
		cross := mix(c[3], b.anchors[0].p, .55)
		delta := point{cross.x - c[3].x, cross.y - c[3].y}
		for j := 2; j <= 7; j++ {
			c[j].x += delta.x
			c[j].y += delta.y
		}
	}
	for i, j := range []int{4, 5} {
		if visible(b.anchors[i+1]) {
			c[j] = mix(c[j], b.anchors[i+1].p, .6)
		}
	}
	// Share the crossing point and align its entering/exiting tangents while
	// the independently anchored handles stretch and turn the knot.
	c[6] = c[3]
	c[2] = point{2*c[3].x - c[4].x, 2*c[3].y - c[4].y}
	c[7] = point{2*c[3].x - c[5].x, 2*c[3].y - c[5].y}
	return c
}
func smoothPoint(current, desired point) point {
	d := math.Hypot(desired.x-current.x, desired.y-current.y)
	if d < .001 {
		return desired
	}
	return mix(current, desired, math.Min(.14, 12/d))
}
func (b *beam) advance(origin point, side float64) {
	b.age++
	desired := b.livingControls(origin, side)
	b.controls[0] = origin
	for j := 1; j < len(b.controls); j++ {
		// Successively bring the handles into the curve rather than creating a
		// completed knot on the first frame. Smoothstep eases each introduction.
		t := math.Max(0, math.Min(1, float64(b.age-(j-1)*7)/24))
		t = t * t * (3 - 2*t)
		straight := mix(origin, b.aim, float64(j)/9)
		b.controls[j] = smoothPoint(b.controls[j], mix(straight, desired[j], t))
	}
	b.reach += 7
	b.path = traceBeam(b.controls, b.reach)
}

// First unwind the visible curve, then draw its straight tip back into the
// cannon. Snapshot only the visible portion, never the accumulated reach.
func (b *beam) retract(origin point) {
	b.target = nil
	b.dwell = 0
	b.anchors = [3]*alien{}
	if len(b.path) < 2 {
		*b = beam{}
		return
	}
	if b.retractAge == 0 {
		start := b.path[0]
		b.retractPath = make([]point, len(b.path))
		for i, p := range b.path {
			b.retractPath[i] = point{p.x - start.x, p.y - start.y}
			if i > 0 {
				prev := b.path[i-1]
				b.retractLength += math.Hypot(p.x-prev.x, p.y-prev.y)
			}
		}
	}
	b.retractAge++
	if b.retractAge >= 72 {
		*b = beam{}
		return
	}
	ease := func(t float64) float64 { t = math.Max(0, math.Min(1, t)); return t * t * (3 - 2*t) }
	unwind := ease(float64(b.retractAge) / 36)
	remaining := 1 - ease(float64(b.retractAge-36)/36)
	length := math.Min(450, b.retractLength) * remaining
	traveled := 0.
	for i, p := range b.retractPath {
		if i > 0 {
			prev := b.retractPath[i-1]
			traveled += math.Hypot(p.x-prev.x, p.y-prev.y)
		}
		fraction := 0.
		if b.retractLength > 0 {
			fraction = traveled / b.retractLength
		}
		straight := point{0, -length * fraction}
		offset := mix(p, straight, unwind)
		b.path[i] = point{origin.x + offset.x, origin.y + offset.y}
	}
}

// Reveal the curve in travel order, clipping its final segment to the growing
// arc length. The source is always attached to the cannon, even while moving.
func traceBeam(c [10]point, reach float64) []point {
	path := make([]point, 0, 97)
	path = append(path, c[0])
	for segment := 0; segment < 3; segment++ {
		i := segment * 3
		for j := 1; j <= 32; j++ {
			p := bezier(c[i], c[i+1], c[i+2], c[i+3], float64(j)/32)
			prev := path[len(path)-1]
			distance := math.Hypot(p.x-prev.x, p.y-prev.y)
			if distance > reach {
				path = append(path, mix(prev, p, reach/distance))
				return path
			}
			path = append(path, p)
			reach -= distance
		}
	}
	return path
}
func beamPath(origin, target point, side, phase float64) []point {
	return traceBeam(beamControls(origin, target, side, phase), math.Inf(1))
}

func stroke(s *ebiten.Image, a, b point, width float32, c color.RGBA) {
	vector.StrokeLine(s, float32(a.x), float32(a.y), float32(b.x), float32(b.y), width, c, true)
}
func (g *game) drawWeapons(s *ebiten.Image) {
	for _, b := range g.beams {
		for layer := 0; layer < 3; layer++ {
			width := []float32{12, 5, 2}[layer] * float32(.3+.7*math.Min(1, float64(b.age)/100))
			c := []color.RGBA{{45, 24, 80, 255}, {170, 105, 255, 255}, {238, 228, 255, 255}}[layer]
			for j := 1; j < len(b.path); j++ {
				stroke(s, b.path[j-1], b.path[j], width, c)
			}
		}
	}
	for _, b := range g.beams {
		if b.dwell > 0 && b.retractAge == 0 && visible(b.target) && len(b.path) > 0 {
			tip := b.path[len(b.path)-1]
			radius := float32(12 + 3*math.Sin(float64(g.tick)*.9))
			vector.DrawFilledCircle(s, float32(tip.x), float32(tip.y), radius*1.6, tint(color.RGBA{170, 80, 255, 255}, .18), true)
			vector.DrawFilledCircle(s, float32(tip.x), float32(tip.y), radius, tint(color.RGBA{255, 155, 80, 255}, .4), true)
			vector.DrawFilledCircle(s, float32(tip.x), float32(tip.y), 5, white, true)
		}
	}
	for _, r := range g.rockets {
		for j := 1; j < len(r.trail); j++ {
			v := uint8(50 + 180*j/len(r.trail))
			stroke(s, r.trail[j-1], r.trail[j], float32(j)/float32(len(r.trail))*4, color.RGBA{v, v / 2, 30, 255})
		}
		back := point{r.p.x - math.Cos(r.angle)*10, r.p.y - math.Sin(r.angle)*10}
		stroke(s, back, r.p, 5, palettes[2])
		stroke(s, mix(back, r.p, .55), r.p, 3, white)
	}
	for _, p := range g.pulses {
		a := point{p.p.x, p.p.y + 24}
		stroke(s, a, p.p, 9, color.RGBA{15, 72, 80, 255})
		stroke(s, a, p.p, 4, palettes[0])
		stroke(s, a, p.p, 1.5, white)
	}
	for _, p := range g.powerups {
		if p.star {
			scale := 2.6 + .35*math.Sin(float64(p.age)*.09)
			vector.DrawFilledCircle(s, float32(p.p.x), float32(p.p.y), 22, tint(palettes[2], .15), true)
			drawSprite(s, g.starImage, p.p.x, p.p.y, scale, 0)
			continue
		}
		c := palettes[0]
		symbol := "P"
		if p.down {
			c = powerdownColor
			symbol = "-"
		}
		r := float32(14 + 2*math.Sin(float64(p.age)*.1))
		vector.DrawFilledCircle(s, float32(p.p.x), float32(p.p.y), r+7, tint(c, .17), true)
		vector.DrawFilledCircle(s, float32(p.p.x), float32(p.p.y), r+3, tint(c, .22), true)
		vector.StrokeCircle(s, float32(p.p.x), float32(p.p.y), r, 2, c, true)
		vector.DrawFilledRect(s, float32(p.p.x-9), float32(p.p.y-9), 18, 18, color.RGBA{20, 30, 56, 255}, false)
		label(s, symbol, int(p.p.x)-6, int(p.p.y)-10, c, 2)
	}
}
func (g *game) initPods() {
	g.starImage = sprite([]string{"....2....", "....1....", "...111...", "211111112", ".1111111.", "..11111..", "..11.11..", ".11...11.", ".2.....2."}, palettes[2])
	g.pods = []*ebiten.Image{
		sprite([]string{"..2..", ".121.", ".111.", "31113", "31113", ".131.", "..3.."}, palettes[2]),
		sprite([]string{".2.2.", ".212.", ".111.", "11111", "31113", ".131.", ".3.3."}, palettes[0]),
		sprite([]string{"2...2", "12.21", "11211", ".121.", "11111", "31113", ".131.", ".3.3."}, palettes[3]),
	}
}
func (g *game) drawPods(s *ebiten.Image) {
	for i, side := range []float64{-1, 1} {
		level := g.weaponLevel
		if g.beams[i].retractAge > 0 {
			level = 3
		}
		if level == 0 {
			continue
		}
		x := g.px + side*27
		vector.DrawFilledRect(s, float32(math.Min(g.px, x)), H-78, 27, 6, palettes[level], false)
		drawSprite(s, g.pods[level-1], x, H-82, 3, 0)
	}
}
