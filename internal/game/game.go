package game

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
	"math"
	"math/rand"
)

const (
	W                = 640
	H                = 800
	waveSize         = 40
	formationColumns = 10
)

type point struct{ x, y float64 }

func mix(a, b point, t float64) point { return point{a.x + (b.x-a.x)*t, a.y + (b.y-a.y)*t} }
func bezier(a, b, c, d point, t float64) point {
	return mix(mix(mix(a, b, t), mix(b, c, t), t), mix(mix(b, c, t), mix(c, d, t), t), t)
}

type alien struct {
	p, home, start  point
	kind            int
	age, wait, mode int
	angle           float64
	side            float64
	hp              int
}
type bullet struct {
	p      point
	vx, vy float64
	enemy  bool
}
type particle struct {
	p            point
	vx, vy, life float64
	c            color.RGBA
}
type star struct {
	p     point
	speed float64
}
type game struct {
	attract                                                    bool
	introTicks, attractTicks                                   int
	cat                                                        pixelCat
	catFrames                                                  []*ebiten.Image
	fragments                                                  []shipFragment
	deathTicks                                                 int
	deathOrigin                                                point
	sound                                                      *soundSystem
	fadePhase, fadeTick, gameOverTicks                         int
	weaponTime                                                 int
	pickupMessage                                              string
	starImage                                                  *ebiten.Image
	downgradeFlash                                             bool
	effects                                                    []effect
	weaponLevel, weaponCool, upgradeFlash, killsSinceDrop      int
	rockets                                                    []rocket
	pulses                                                     []pulse
	powerups                                                   []powerup
	beams                                                      [2]beam
	pods                                                       []*ebiten.Image
	aliens                                                     []*alien
	bullets                                                    []bullet
	particles                                                  []particle
	stars                                                      []star
	sprites                                                    []*ebiten.Image
	player                                                     *ebiten.Image
	tick, wave, score, best, lives, cool, inv, spawn, nextWave int
	px                                                         float64
	started, over, paused                                      bool
	rng                                                        *rand.Rand
}

var palettes = []color.RGBA{{61, 242, 194, 255}, {247, 91, 154, 255}, {255, 192, 74, 255}, {158, 135, 255, 255}}

func sprite(rows []string, c color.RGBA) *ebiten.Image {
	im := ebiten.NewImage(len(rows[0]), len(rows))
	for y, r := range rows {
		for x, v := range r {
			switch v {
			case '1':
				im.Set(x, y, c)
			case '2':
				im.Set(x, y, color.RGBA{230, 253, 255, 255})
			case '3':
				im.Set(x, y, color.RGBA{42, 59, 103, 255})
			}
		}
	}
	return im
}
func newGame() *game {
	g := &game{rng: rand.New(rand.NewSource(42)), px: W / 2}
	g.player = sprite([]string{"......2......", "......2......", ".....212.....", ".....212.....", "..1..111..1..", "..1.11111.1..", ".11111111111.", "1112111112111", "1111111111111", "...11...11...", "...3.....3..."}, color.RGBA{87, 187, 255, 255})
	patterns := [][]string{
		{"..1.....1..", "...1...1...", "..1111111..", ".112111211.", "11111111111", "1.1111111.1", "1.1.....1.1", "...11.11..."},
		{"....111....", "..1111111..", ".111212111.", "11111111111", "11111111111", "..11.11.1..", ".11...11...", "11.....11.."},
		{"1...111...1", "11.11111.11", "11121112111", ".111111111.", "..1111111..", ".111...111.", "11..1.1..11", "1.......1.."},
		{"....1.1....", "..1111111..", ".112111211.", "11111111111", "1.1111111.1", "..1111111..", "...1.1.1...", "..1.....1.."}}
	for i, p := range patterns {
		g.sprites = append(g.sprites, sprite(p, palettes[i]))
	}
	for i := 0; i < 120; i++ {
		g.stars = append(g.stars, star{point{g.rng.Float64() * W, g.rng.Float64() * H}, .3 + g.rng.Float64()*1.7})
	}
	g.fadePhase = fadeIn
	g.initCat()
	g.initPods()
	return g
}
func (g *game) reset() {
	g.attract = false
	g.introTicks = 0
	g.attractTicks = 0
	g.cat = pixelCat{wait: 300}
	g.fragments = nil
	g.deathTicks = 0
	g.fadePhase = 0
	g.fadeTick = 0
	g.gameOverTicks = 0
	g.aliens = nil
	g.bullets = nil
	g.particles = nil
	g.effects = nil
	g.score = 0
	g.weaponLevel = 0
	g.weaponTime = 0
	g.pickupMessage = ""
	g.weaponCool = 0
	g.upgradeFlash = 0
	g.downgradeFlash = false
	g.killsSinceDrop = 0
	g.cool = 0
	g.rockets = nil
	g.pulses = nil
	g.powerups = nil
	g.beams = [2]beam{}
	g.lives = 3
	g.wave = 0
	g.tick = 0
	g.spawn = 0
	g.nextWave = 60
	g.px = W / 2
	g.inv = 120
	g.started = true
	g.over = false
	g.paused = false
}
func formation(w, i int) point {
	row, col := i/formationColumns, i%formationColumns
	x := float64(col) - 4.5
	y := float64(row)
	switch w % 3 {
	case 1:
		return point{W/2 + x*50, 135 + y*48 + math.Abs(x)*12}
	case 2:
		return point{W/2 + x*50, 165 + y*49 + math.Sin(float64(col)*math.Pi/4.5)*32}
	default:
		return point{W/2 + x*50, 140 + y*50}
	}
}
func (g *game) emit(p point, c color.RGBA, n int) {
	for i := 0; i < n; i++ {
		a := g.rng.Float64() * math.Pi * 2
		s := 1 + g.rng.Float64()*4
		g.particles = append(g.particles, particle{p, math.Cos(a) * s, math.Sin(a) * s, 20 + g.rng.Float64()*25, c})
	}
}
func (g *game) Update() error {
	defer g.syncAudio()
	if !g.over && !g.attract && g.fadePhase != fadeToAttract && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	pressed := len(inpututil.AppendJustPressedKeys(nil)) > 0 || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle)
	g.updateAttractSequence(pressed)
	transitioning := g.updateTransition(pressed)
	if !g.started && !transitioning {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.fadePhase = fadeLaunch
			g.fadeTick = 0
		}
	} else if g.started && !g.attract && !g.over && !transitioning && inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.paused = !g.paused
	}
	if g.paused {
		return nil
	}
	for i := range g.stars {
		s := &g.stars[i]
		s.p.y += s.speed
		if s.p.y > H {
			s.p.y = 0
		}
	}
	g.updateEffects()
	g.updateShipExplosion()
	if g.started && !g.over {
		g.updateCat()
	}
	ps := g.particles[:0]
	for _, p := range g.particles {
		p.p.x += p.vx
		p.p.y += p.vy
		p.vx *= .97
		p.vy *= .97
		p.life--
		if p.life > 0 {
			ps = append(ps, p)
		}
	}
	g.particles = ps
	if !g.started || g.over {
		return nil
	}
	g.tick++
	g.updateWeaponTimer()
	if g.inv > 0 {
		g.inv--
	}
	if g.cool > 0 {
		g.cool--
	}
	if g.attract {
		g.autopilot()
	} else {
		if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
			g.px -= 5
		}
		if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
			g.px += 5
		}
	}
	g.px = math.Max(40, math.Min(W-40, g.px))
	if (g.attract || ebiten.IsKeyPressed(ebiten.KeySpace)) && g.cool == 0 {
		g.bullets = append(g.bullets, bullet{point{g.px, H - 91}, 0, -9, false})
		g.sfx("blaster")
		g.cool = 9
	}
	if len(g.aliens) == 0 && g.spawn == 0 {
		g.nextWave--
		if g.nextWave <= 0 {
			g.wave++
			g.spawn = waveSize
			g.nextWave = waveBreakTicks
			if g.attract {
				g.nextWave = 60
			}
			g.bullets = nil
		}
	}
	if g.spawn > 0 && g.tick%8 == 0 {
		i := waveSize - g.spawn
		side := 1.
		if i%2 == 0 {
			side = -1
		}
		start := point{W/2 + side*400, -50}
		g.aliens = append(g.aliens, &alien{p: start, start: start, home: formation(g.wave, i), kind: (i/formationColumns + g.wave - 1) % 4, side: side, hp: 1})
		g.spawn--
	}
	for _, a := range g.aliens {
		prev := a.p
		a.age++
		switch a.mode {
		case 0:
			t := math.Min(1, float64(a.age)/160)
			a.p = bezier(a.start, point{W/2 - a.side*340, 430}, point{a.home.x - a.side*170, 20}, a.home, t)
			if t >= 1 {
				a.mode = 1
				a.age = 0
				a.wait = 180 + g.rng.Intn(480)
			}
		case 1:
			a.p = point{a.home.x + math.Sin(float64(g.tick)*.018)*16, a.home.y + math.Sin(float64(g.tick)*.027+a.home.x)*4}
			a.angle += angleDelta(a.angle, 0) * .07
			if a.age > a.wait {
				a.mode = 2
				a.age = 0
				a.start = a.p
			}
		case 2:
			t := math.Min(1, float64(a.age)/float64(max(180, 290-g.wave*7)))
			a.p = bezier(a.start, point{a.start.x + a.side*310, 390}, point{W/2 - a.side*280, H + 240}, point{W/2 - a.side*230, -65}, t)
			if a.age%65 == 0 {
				dx := g.px - a.p.x
				dy := H - 75 - a.p.y
				dist := math.Hypot(dx, dy)
				if dist > 1 {
					speed := math.Min(4.5, 2.4+float64(g.wave)*.15)
					g.bullets = append(g.bullets, bullet{a.p, dx / dist * speed, dy / dist * speed, true})
				}
			}
			if t >= 1 {
				a.mode = 0
				a.age = 0
				a.start = a.p
			}
		}
		if a.mode != 1 {
			d := point{a.p.x - prev.x, a.p.y - prev.y}
			if math.Hypot(d.x, d.y) > .01 {
				target := math.Atan2(d.y, d.x) - math.Pi/2
				a.angle += angleDelta(a.angle, target) * .16
			}
		}
		if g.inv == 0 && math.Hypot(a.p.x-g.px, a.p.y-(H-75)) < 25 {
			g.hit()
		}
	}
	bs := g.bullets[:0]
	for _, b := range g.bullets {
		b.p.x += b.vx
		b.p.y += b.vy
		dead := b.p.y < 0 || b.p.y > H || b.p.x < 0 || b.p.x > W
		if b.enemy {
			if g.inv == 0 && math.Hypot(b.p.x-g.px, b.p.y-(H-75)) < 15 {
				g.hit()
				dead = true
			}
		} else {
			for _, a := range g.aliens {
				if a.hp > 0 && math.Abs(b.p.x-a.p.x) < 17 && math.Abs(b.p.y-a.p.y) < 15 {
					g.destroy(a)
					dead = true
					break
				}
			}
		}
		if !dead {
			bs = append(bs, b)
		}
	}
	g.bullets = bs
	g.updateWeapons((g.attract || ebiten.IsKeyPressed(ebiten.KeySpace)) && !g.over)
	g.updatePowerups()
	as := g.aliens[:0]
	for _, a := range g.aliens {
		if a.hp > 0 {
			as = append(as, a)
		}
	}
	g.aliens = as
	if !g.attract && g.score > g.best {
		g.best = g.score
	}
	return nil
}
func angleDelta(a, b float64) float64 { return math.Atan2(math.Sin(b-a), math.Cos(b-a)) }
func (g *game) hit() {
	if g.over || g.attract {
		return
	}
	g.sfx("hit")
	g.lives--
	g.inv = 150
	g.emit(point{g.px, H - 75}, color.RGBA{100, 195, 255, 255}, 40)
	if g.lives <= 0 {
		g.explodeShip()
		g.over = true
	}
}
func drawSprite(dst, src *ebiten.Image, x, y, scale, angle float64) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(src.Bounds().Dx())/2, -float64(src.Bounds().Dy())/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Rotate(angle)
	op.GeoM.Translate(x, y)
	dst.DrawImage(src, op)
}
func label(dst *ebiten.Image, s string, x, y int, c color.RGBA, scale float64) {
	im := ebiten.NewImage(max(1, len(s)*6), 16)
	ebitenutil.DebugPrint(im, s)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	dst.DrawImage(im, op)
}
func center(dst *ebiten.Image, s string, y int, scale float64, c color.RGBA) {
	label(dst, s, int(W/2-float64(len(s)*6)*scale/2), y, c, scale)
}

var white = color.RGBA{220, 235, 255, 255}

func (g *game) Draw(s *ebiten.Image) {
	s.Fill(color.RGBA{6, 9, 23, 255})
	for _, st := range g.stars {
		v := uint8(55 + st.speed*65)
		vector.DrawFilledRect(s, float32(st.p.x), float32(st.p.y), 1, float32(1+st.speed/2), color.RGBA{v, v, 255, 255}, false)
	}
	vector.StrokeLine(s, 24, 73, W-24, 73, 1, color.RGBA{36, 52, 77, 255}, false)
	label(s, "SCORE", 24, 18, palettes[0], 1)
	label(s, fmt.Sprintf("%06d", g.score), 24, 35, white, 2)
	label(s, "BEST", 272, 18, palettes[1], 1)
	label(s, fmt.Sprintf("%06d", g.best), 272, 35, white, 2)
	label(s, fmt.Sprintf("WAVE %02d", g.wave), 405, 18, white, 1)
	if g.weaponLevel > 0 {
		label(s, "POWER", 535, 6, palettes[0], 1)
		timerColor := palettes[0]
		if g.weaponTime <= 180 {
			timerColor = powerdownColor
		}
		label(s, fmt.Sprintf("%02d", (g.weaponTime+59)/60), 530, 23, timerColor, 3)
	}
	g.drawEffects(s, false)
	for _, a := range g.aliens {
		drawSprite(s, g.sprites[a.kind], a.p.x, a.p.y, 3, a.angle)
	}
	for _, b := range g.bullets {
		c := palettes[0]
		if b.enemy {
			c = palettes[1]
		}
		vector.DrawFilledRect(s, float32(b.p.x-2), float32(b.p.y-5), 4, 10, c, false)
	}
	g.drawEffects(s, true)
	g.drawWeapons(s)
	for _, p := range g.particles {
		c := p.c
		c.A = uint8(math.Min(255, p.life*12))
		vector.DrawFilledRect(s, float32(p.p.x), float32(p.p.y), 3, 3, c, false)
	}
	g.drawShipExplosion(s)
	if g.started && !g.over && (g.inv == 0 || g.inv%12 < 6) {
		drawSprite(s, g.player, g.px, H-75, 3, 0)
		g.drawPods(s)
		flame := float32(8 + g.tick%7)
		vector.DrawFilledRect(s, float32(g.px-4), H-58, 8, flame, palettes[2], false)
	}
	if g.started {
		label(s, weaponNames[g.weaponLevel], 150, H-32, palettes[g.weaponLevel], 1)
		if g.upgradeFlash > 0 {
			message := "UPGRADE: "
			c := palettes[g.weaponLevel]
			if g.downgradeFlash {
				message = "DOWNGRADE: "
				c = powerdownColor
			}
			if g.pickupMessage != "" {
				message = g.pickupMessage
			} else {
				message += weaponNames[g.weaponLevel]
			}
			center(s, message, H-145, 1.5, c)
		}
		for i := 0; i < g.lives; i++ {
			drawSprite(s, g.player, float64(35+i*28), H-25, 1.5, 0)
		}
		label(s, "P PAUSE  /  ESC EXIT", 440, H-32, color.RGBA{108, 128, 164, 255}, 1)
		if len(g.aliens) == 0 && g.spawn == 0 && !g.over {
			center(s, fmt.Sprintf("SECTOR %02d CLEAR", g.wave), 365, 2, palettes[0])
		}
	}
	if !g.started {
		vector.DrawFilledRect(s, 45, 190, 550, 405, color.RGBA{9, 15, 34, 240}, false)
		center(s, "A S T R A", 220, 5, white)
		center(s, "G A L A X I A N S", 296, 2, palettes[0])
		for i := 0; i < 4; i++ {
			drawSprite(s, g.sprites[i], float64(200+i*80), 382, 4, 0)
		}
		center(s, "ARROWS / A D  TO MOVE", 455, 1.5, white)
		center(s, "HOLD SPACE TO FIRE", 485, 1.5, white)
		center(s, "PRESS ENTER TO LAUNCH", 546, 2, palettes[2])
		center(s, "40 HOSTILES  /  THREE FORMATIONS  /  ENDLESS WAVES", 657, 1, color.RGBA{108, 128, 164, 255})
	}
	if g.attract {
		center(s, "ATTRACT MODE - PRESS ANY KEY", 96, 1.5, palettes[2])
	}
	g.drawCat(s)
	if g.gameOverVisible() || g.paused {
		vector.DrawFilledRect(s, 75, 290, 490, 200, color.RGBA{8, 13, 31, 240}, false)
		title := "PAUSED"
		sub := "PRESS P TO RESUME"
		if g.over {
			title = "GAME OVER"
			sub = "ANY KEY / CLICK TO RETURN"
		}
		center(s, title, 326, 3, white)
		center(s, fmt.Sprintf("SCORE %06d", g.score), 386, 2, palettes[0])
		center(s, sub, 444, 1.5, palettes[2])
	}
	alpha := g.fadeAlpha()
	if alpha > 0 {
		vector.DrawFilledRect(s, 0, 0, W, H, color.RGBA{0, 0, 0, uint8(math.Round(alpha * 255))}, false)
	}
}
func (g *game) Layout(int, int) (int, int) { return W, H }
