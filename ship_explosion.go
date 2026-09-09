package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type shipFragment struct {
	p, velocity point
	c           color.RGBA
	life, total float64
}

func (g *game) shatterSprite(sprite *ebiten.Image, origin point) {
	if sprite == nil {
		return
	}
	w, h := sprite.Bounds().Dx(), sprite.Bounds().Dy()
	pixels := make([]byte, w*h*4)
	sprite.ReadPixels(pixels)
	g.shatterPixels(pixels, w, h, origin)
}

// Every opaque source pixel becomes one original-size, original-color block.
func (g *game) shatterPixels(pixels []byte, w, h int, origin point) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y*w + x) * 4
			if pixels[i+3] == 0 {
				continue
			}
			p := point{origin.x + (float64(x)+.5-float64(w)/2)*3, origin.y + (float64(y)+.5-float64(h)/2)*3}
			a := g.rng.Float64() * math.Pi * 2
			speed := 2 + g.rng.Float64()*9
			life := 150 + g.rng.Float64()*150
			g.fragments = append(g.fragments, shipFragment{p, point{math.Cos(a)*speed + (p.x-g.px)*.06, math.Sin(a)*speed - 5}, color.RGBA{pixels[i], pixels[i+1], pixels[i+2], pixels[i+3]}, life, life})
		}
	}
}
func (g *game) explodeShip() {
	g.deathTicks = 1
	g.deathOrigin = point{g.px, H - 75}
	g.shatterSprite(g.player, g.deathOrigin)
	for i, side := range []float64{-1, 1} {
		level := g.weaponLevel
		if g.beams[i].retractAge > 0 {
			level = 3
		}
		if level > 0 && len(g.pods) >= level {
			g.shatterSprite(g.pods[level-1], point{g.px + side*27, H - 82})
			// The solid three-pixel connector joining each cannon to the fuselage.
			pixels := make([]byte, 9*2*4)
			c := palettes[level]
			for j := 0; j < len(pixels); j += 4 {
				pixels[j] = c.R
				pixels[j+1] = c.G
				pixels[j+2] = c.B
				pixels[j+3] = 255
			}
			g.shatterPixels(pixels, 9, 2, point{g.px + side*13.5, H - 75})
		}
	}
	// Reserve room even if the player dies during a particle-heavy battle.
	if len(g.effects) > maxEffects-500 {
		g.effects = g.effects[len(g.effects)-(maxEffects-500):]
	}
	g.extremeExplosion(g.deathOrigin)
	for _, dx := range []float64{-24, 24} {
		g.extremeExplosion(point{g.px + dx, H - 88})
	}
	g.deathPlume()
	// The destroyed ship can no longer sustain its beams.
	g.beams = [2]beam{}
}
func (g *game) deathPlume() {
	for i := 0; i < 24; i++ {
		life := 100 + g.rng.Float64()*90
		kind := smokeEffect
		if i%2 == 0 {
			kind = flameEffect
			life = 45 + g.rng.Float64()*35
		}
		p := point{g.deathOrigin.x + (g.rng.Float64()-.5)*50, g.deathOrigin.y - 10 - g.rng.Float64()*30}
		g.addEffect(effect{p, point{(g.rng.Float64() - .5) * 4, -2 - g.rng.Float64()*4}, 22 + g.rng.Float64()*28, life, life, kind})
	}
}
func (g *game) updateShipExplosion() {
	if g.deathTicks > 0 {
		g.deathTicks++
		if g.deathTicks <= 48 && g.deathTicks%8 == 0 {
			g.deathPlume()
		}
	}
	kept := g.fragments[:0]
	for _, f := range g.fragments {
		f.life--
		f.velocity.x *= .994
		f.velocity.y += .13
		f.p.x += f.velocity.x
		f.p.y += f.velocity.y
		if f.life > 0 && f.p.x >= -6 && f.p.x <= W+6 && f.p.y <= H+6 {
			kept = append(kept, f)
		}
	}
	g.fragments = kept
}
func (g *game) drawShipExplosion(s *ebiten.Image) {
	if g.deathTicks > 0 && g.deathTicks < 45 {
		t := float64(g.deathTicks) / 45
		vector.StrokeCircle(s, float32(g.deathOrigin.x), float32(g.deathOrigin.y), float32(10+160*t), float32(5*(1-t)+1), tint(palettes[2], .6*(1-t)), true)
	}
	for _, f := range g.fragments {
		alpha := math.Min(1, f.life/(f.total*.75))
		vector.DrawFilledRect(s, float32(f.p.x)-1.5, float32(f.p.y)-1.5, 3, 3, tint(f.c, alpha), false)
	}
}
