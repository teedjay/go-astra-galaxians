package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type effect struct {
	p, velocity         point
	radius, life, total float64
	kind                int
}

const (
	smokeEffect = iota
	flameEffect
	sparkEffect
	dustEffect
)
const maxEffects = 1800

func (g *game) addEffect(e effect) {
	if len(g.effects) < maxEffects {
		g.effects = append(g.effects, e)
	}
}
func (g *game) rocketSmoke(p point, angle, speed float64) {
	p.x -= math.Cos(angle) * 11
	p.y -= math.Sin(angle) * 11
	life := 30 + g.rng.Float64()*22
	g.addEffect(effect{p, point{-math.Cos(angle)*speed*.13 + (g.rng.Float64()-.5)*.6, -math.Sin(angle) * speed * .13}, 3 + g.rng.Float64()*3, life, life, smokeEffect})
}
func (g *game) laserSparks(p point) {
	for i := 0; i < 3; i++ {
		a := g.rng.Float64() * math.Pi * 2
		speed := 1.5 + g.rng.Float64()*4
		life := 10 + g.rng.Float64()*14
		g.addEffect(effect{p, point{math.Cos(a) * speed, math.Sin(a) * speed}, 1.5, life, life, sparkEffect})
	}
}
func (g *game) extremeExplosion(p point) {
	// A fast fireball opens into large, slower plumes and flying hot fragments.
	for i := 0; i < 65; i++ {
		a := g.rng.Float64() * math.Pi * 2
		speed := 1 + g.rng.Float64()*5
		kind := flameEffect
		radius := 12 + g.rng.Float64()*20
		life := 25 + g.rng.Float64()*32
		if i < 28 {
			kind = smokeEffect
			radius = 20 + g.rng.Float64()*25
			life = 65 + g.rng.Float64()*55
			speed *= .65
		}
		g.addEffect(effect{point{p.x + (g.rng.Float64()-.5)*16, p.y + (g.rng.Float64()-.5)*16}, point{math.Cos(a) * speed, math.Sin(a) * speed}, radius, life, life, kind})
	}
	for i := 0; i < 45; i++ {
		a := g.rng.Float64() * math.Pi * 2
		speed := 3 + g.rng.Float64()*8
		life := 20 + g.rng.Float64()*30
		g.addEffect(effect{p, point{math.Cos(a) * speed, math.Sin(a) * speed}, 2, life, life, sparkEffect})
	}
}
func (g *game) updateEffects() {
	kept := g.effects[:0]
	for _, e := range g.effects {
		e.life--
		if e.life <= 0 {
			continue
		}
		e.p.x += e.velocity.x
		e.p.y += e.velocity.y
		switch e.kind {
		case smokeEffect:
			e.velocity.x *= .97
			e.velocity.y = e.velocity.y*.97 - .018
			e.radius += .14
		case flameEffect:
			e.velocity.x *= .95
			e.velocity.y = e.velocity.y*.95 - .025
			e.radius += .12
		case sparkEffect, dustEffect:
			e.velocity.y += .035
		}
		kept = append(kept, e)
	}
	g.effects = kept
}

// Ebitengine expects premultiplied alpha for translucent particle colors.
func tint(c color.RGBA, alpha float64) color.RGBA {
	alpha = math.Max(0, math.Min(1, alpha))
	return color.RGBA{uint8(float64(c.R) * alpha), uint8(float64(c.G) * alpha), uint8(float64(c.B) * alpha), uint8(255 * alpha)}
}
func (g *game) drawEffects(s *ebiten.Image, hot bool) {
	for _, e := range g.effects {
		if (e.kind != smokeEffect) != hot {
			continue
		}
		t := e.life / e.total
		if e.kind == dustEffect {
			r := e.radius * (.6 + .4*math.Sin(e.life*.5))
			c := tint(palettes[2], t)
			stroke(s, point{e.p.x - r, e.p.y}, point{e.p.x + r, e.p.y}, 1.5, c)
			stroke(s, point{e.p.x, e.p.y - r}, point{e.p.x, e.p.y + r}, 1.5, c)
			continue
		}
		if e.kind == sparkEffect {
			stroke(s, point{e.p.x - e.velocity.x*2, e.p.y - e.velocity.y*2}, e.p, float32(e.radius), tint(color.RGBA{255, 211, 126, 255}, t))
			continue
		}
		c := color.RGBA{100, 108, 129, 255}
		alpha := math.Min(.42, t*.65)
		if e.kind == flameEffect {
			c = color.RGBA{255, uint8(45 + 150*t), uint8(12 + 90*t*t), 255}
			alpha = t * .85
		}
		// Overlapping square lobes keep large clouds consistent with the pixel art.
		r := float32(e.radius)
		x, y := float32(e.p.x), float32(e.p.y)
		vector.DrawFilledRect(s, x-r*.7, y-r, r*1.4, r*2, tint(c, alpha), false)
		vector.DrawFilledRect(s, x-r, y-r*.6, r*2, r*1.2, tint(c, alpha*.7), false)
		if e.kind == flameEffect {
			vector.DrawFilledRect(s, x-r*.3, y-r*.35, r*.6, r*.7, tint(color.RGBA{255, 239, 167, 255}, t*t*.9), false)
		}
	}
}

func (g *game) starDust(p point) {
	for i := 0; i < 45; i++ {
		a := g.rng.Float64() * math.Pi * 2
		speed := 1 + g.rng.Float64()*5
		life := 25 + g.rng.Float64()*30
		g.addEffect(effect{p, point{math.Cos(a) * speed, math.Sin(a) * speed}, 2 + g.rng.Float64()*4, life, life, dustEffect})
	}
}
