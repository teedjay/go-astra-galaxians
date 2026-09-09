package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
)

type pixelCat struct {
	x         float64
	direction float64
	age, wait int
	active    bool
}

func (g *game) initCat() {
	body := []string{"....1...1.......", "....11.11.......", "....11211.......", "....11111.......", ".....11111111..1", "....111111111111", "....11111111111.", ".....111111111.."}
	for _, legs := range [][]string{{".....11...11....", ".....1.....1...."}, {"......11.11.....", ".......1.1......"}} {
		rows := append(append([]string(nil), body...), legs...)
		g.catFrames = append(g.catFrames, sprite(rows, color.RGBA{243, 173, 83, 255}))
	}
}
func (g *game) updateCat() {
	c := &g.cat
	if !c.active {
		c.wait--
		if c.wait > 0 {
			return
		}
		c.active = true
		c.age = 0
		c.direction = 1
		c.x = -30
		if g.rng.Intn(2) == 0 {
			c.direction = -1
			c.x = W + 30
		}
	}
	c.age++
	c.x += c.direction * 1.2
	if c.x < -40 || c.x > W+40 {
		c.active = false
		c.wait = 720 + g.rng.Intn(1080)
	}
}
func (g *game) drawCat(s *ebiten.Image) {
	c := g.cat
	if !c.active || len(g.catFrames) == 0 {
		return
	}
	im := g.catFrames[(c.age/9)%len(g.catFrames)]
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(im.Bounds().Dx())/2, -float64(im.Bounds().Dy())/2)
	// Sprite faces left, so mirror it when walking right.
	op.GeoM.Scale(-c.direction*2.5, 2.5)
	op.GeoM.Translate(c.x, H-49)
	s.DrawImage(im, op)
}
