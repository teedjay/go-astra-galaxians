package game

import (
	"bytes"
	"encoding/binary"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const soundRate = 44100

type soundSystem struct {
	attractMusic        *audio.Player
	attractLength       time.Duration
	interlude           *audio.Player
	interludeWave       int
	interludePaused     bool
	intro, ending, beam *audio.Player
	voices              map[string][]*audio.Player
	previousScene       int
	beamVolume          float64
}

func newSoundSystem() (*soundSystem, error) {
	ctx := audio.NewContext(soundRate)
	loop := func(data []byte) (*audio.Player, error) {
		return ctx.NewPlayer(audio.NewInfiniteLoop(bytes.NewReader(data), int64(len(data))))
	}
	intro := ctx.NewPlayerFromBytes(makeSIDTune())
	attractPCM := makeAttractTune()
	beam, err := loop(makeBeamHum())
	if err != nil {
		return nil, err
	}
	s := &soundSystem{attractMusic: ctx.NewPlayerFromBytes(attractPCM), attractLength: time.Duration(len(attractPCM)/4) * time.Second / soundRate, interlude: ctx.NewPlayerFromBytes(makeWaveJingle()), intro: intro, ending: ctx.NewPlayerFromBytes(makeTune(true)), beam: beam, voices: map[string][]*audio.Player{}, previousScene: -1}
	for _, name := range []string{"blaster", "rocket", "pulse", "explosion", "huge", "hit", "up", "down", "star"} {
		data := makeSound(name)
		for i := 0; i < 4; i++ {
			p := ctx.NewPlayerFromBytes(data)
			p.SetVolume(.28)
			s.voices[name] = append(s.voices[name], p)
		}
	}
	return s, nil
}
func (g *game) sfx(name string) {
	if g.sound == nil || g.over || g.paused || g.attract {
		return
	}
	for _, p := range g.sound.voices[name] {
		if !p.IsPlaying() {
			_ = p.Rewind()
			p.Play()
			return
		}
	}
}
func (g *game) syncAudio() {
	s := g.sound
	if s == nil {
		return
	}
	scene := 0
	if g.started {
		scene = 1
	}
	if g.over {
		scene = 3
		if g.gameOverVisible() {
			scene = 2
		}
	}
	if g.attract {
		scene = 4
	}
	if scene != s.previousScene {
		s.attractMusic.Pause()
		s.interlude.Pause()
		s.interludeWave = 0
		s.interludePaused = false
		s.intro.Pause()
		s.ending.Pause()
		s.beam.Pause()
		s.beamVolume = 0
		for _, pool := range s.voices {
			for _, p := range pool {
				p.Pause()
			}
		}
		switch scene {
		case 4:
			_ = s.attractMusic.Rewind()
			s.attractMusic.SetVolume(0)
			s.attractMusic.Play()
		case 0:
			_ = s.intro.Rewind()
			s.intro.SetVolume(0)
			s.intro.Play()
		case 3:
			boom := s.voices["huge"][0]
			_ = boom.Rewind()
			boom.SetVolume(.4)
			boom.Play()
		case 2:
			_ = s.ending.Rewind()
			s.ending.SetVolume(.5)
			s.ending.Play()
		}
		s.previousScene = scene
	}
	if scene == 0 {
		s.intro.SetVolume(.42 * (1 - g.fadeAlpha()))
	}
	if scene == 2 {
		s.ending.SetVolume(.5 * (1 - g.fadeAlpha()))
	}
	if scene == 4 {
		s.attractMusic.SetVolume(.48 * (1 - g.fadeAlpha()))
	}
	if scene != 1 {
		return
	}
	between := g.wave > 0 && len(g.aliens) == 0 && g.spawn == 0
	if between {
		if !g.paused && s.interludeWave != g.wave {
			s.interludeWave = g.wave
			_ = s.interlude.Rewind()
			s.interlude.SetVolume(.42)
			s.interlude.Play()
		}
		if g.paused && s.interlude.IsPlaying() {
			s.interlude.Pause()
			s.interludePaused = true
		}
		if !g.paused && s.interludePaused {
			s.interlude.Play()
			s.interludePaused = false
		}
	} else {
		s.interlude.Pause()
		s.interludePaused = false
	}
	active := false
	contact := false
	if !g.paused {
		for _, b := range g.beams {
			if b.age > 0 && b.retractAge == 0 && b.target != nil {
				active = true
				if b.dwell > 0 {
					contact = true
				}
			}
		}
	}
	target := 0.
	if active {
		target = .12
	}
	if contact {
		target = .21
	}
	s.beamVolume += (target - s.beamVolume) * .18
	if s.beamVolume < .001 {
		s.beam.Pause()
	} else {
		s.beam.SetVolume(s.beamVolume)
		if !s.beam.IsPlaying() {
			s.beam.Play()
		}
	}
	if g.paused {
		for _, pool := range s.voices {
			for _, p := range pool {
				p.Pause()
			}
		}
	}
}

// Original music and effects are synthesized as stereo signed 16-bit PCM.
// Pulse leads, triangle bass and noise percussion give a small-console sound.
func pcm(samples []float64) []byte {
	out := make([]byte, len(samples)*4)
	for i, v := range samples {
		v = math.Max(-1, math.Min(1, v))
		n := uint16(int16(v * 29000))
		binary.LittleEndian.PutUint16(out[i*4:], n)
		binary.LittleEndian.PutUint16(out[i*4+2:], n)
	}
	return out
}
func noteHz(note int) float64 { return 440 * math.Pow(2, float64(note-69)/12) }
func addNote(dst []float64, start, duration float64, note int, gain float64, triangle bool) {
	begin := int(start * soundRate)
	count := int(duration * soundRate)
	phase := 0.
	for i := 0; i < count && begin+i < len(dst); i++ {
		t := float64(i) / soundRate
		phase += noteHz(note) / soundRate
		phase -= math.Floor(phase)
		wave := -.5
		if phase < .25 {
			wave = 1.5
		}
		if triangle {
			wave = 1 - 4*math.Abs(phase-.5)
		}
		envelope := math.Min(1, t/.004) * math.Min(1, (duration-t)/.018) * math.Exp(-t/(duration*1.5))
		dst[begin+i] += wave * gain * envelope
	}
}
func addDrum(dst []float64, start float64, kind int) {
	duration := .12
	if kind == 1 {
		duration = .18
	}
	seed := uint32(9123 + kind)
	phase := 0.
	for i := 0; i < int(duration*soundRate); i++ {
		index := int(start*soundRate) + i
		if index >= len(dst) {
			break
		}
		t := float64(i) / soundRate
		seed = seed*1664525 + 1013904223
		noise := float64(seed>>16)/32768 - 1
		v := noise * .10 * math.Exp(-t*45)
		if kind == 0 {
			phase += (55 + 150*math.Exp(-t*45)) / soundRate
			v = math.Sin(phase*2*math.Pi) * .32 * math.Exp(-t*27)
		}
		if kind == 1 {
			v = noise * .18 * math.Exp(-t*22)
		}
		dst[index] += v * math.Min(1, t/.001)
	}
}
func makeTune(ending bool) []byte {
	if ending {
		step := .14
		dst := make([]float64, int(3.0*soundRate))
		melody := []int{84, 79, 76, 79, 83, 79, 74, 76, 79, 76, 72, 71, 72}
		for i, n := range melody {
			duration := step * .85
			if i == len(melody)-1 {
				duration = .9
			}
			addNote(dst, float64(i)*step, duration, n, .16, false)
			if i%4 == 0 {
				addNote(dst, float64(i)*step, .45, n-24, .2, true)
				addDrum(dst, float64(i)*step, 0)
			}
		}
		for _, n := range []int{48, 55, 60} {
			addNote(dst, 1.68, 1.1, n, .1, true)
		}
		return pcm(dst)
	}
	step := 60. / 168 / 4
	bars := 8
	dst := make([]float64, int(float64(bars*16)*step*soundRate))
	roots := []int{48, 44, 46, 43, 48, 44, 46, 43}
	motifs := [][]int{{24, 31, 27, 31, 36, 31, 27, 34, 31, 27, 24, 27, 31, 34, 36, 31}, {24, 27, 31, 34, 36, 34, 31, 27, 29, 31, 34, 31, 27, 26, 24, 19}}
	for bar, root := range roots {
		for j := 0; j < 16; j++ {
			at := float64(bar*16+j) * step
			addNote(dst, at, step*.82, root+motifs[(bar/4)%2][j], .085, false)
			if j%2 == 0 {
				bass := root
				if j%4 == 2 {
					bass += 12
				}
				addNote(dst, at, step*1.65, bass, .20, true)
			}
			if j%4 == 0 {
				kind := 0
				if j%8 == 4 {
					kind = 1
				}
				addDrum(dst, at, kind)
			}
			if j%2 == 1 {
				addDrum(dst, at, 2)
			}
		}
	}
	return pcm(dst)
}
func makeSound(name string) []byte {
	duration := .18
	switch name {
	case "rocket":
		duration = .38
	case "explosion":
		duration = .35
	case "huge":
		duration = 1.
	case "hit":
		duration = .5
	case "up", "down", "star":
		duration = .3
	}
	samples := make([]float64, int(duration*soundRate))
	phase := 0.
	seed := uint32(391)
	for i := range samples {
		t := float64(i) / soundRate
		u := t / duration
		seed = seed*1664525 + 1013904223
		noise := float64(seed>>16)/32768 - 1
		hz := 900 * math.Pow(.12, u)
		blend := 0.
		gain := .55
		switch name {
		case "pulse":
			hz = 1800 - 1200*u
			gain = .35
		case "rocket":
			hz = 90 + 380*u
			blend = .55
		case "explosion":
			hz = 100 - 70*u
			blend = .8
		case "huge":
			hz = 65 - 40*u
			blend = .75
			gain = .85
		case "hit":
			hz = 400 * math.Pow(.08, u)
			blend = .35
		case "up", "star":
			hz = noteHz(72 + int(u*4)*4)
		case "down":
			hz = noteHz(72 - int(u*4)*4)
		}
		phase += hz / soundRate
		phase -= math.Floor(phase)
		wave := 1.
		if phase > .25 {
			wave = -.333
		}
		if name == "huge" {
			wave = math.Sin(phase * 2 * math.Pi)
		}
		envelope := math.Min(1, t/.003) * math.Pow(1-u, 1.7)
		samples[i] = (wave*(1-blend) + noise*blend) * gain * envelope
	}
	return pcm(samples)
}
func makeBeamHum() []byte {
	dst := make([]float64, soundRate)
	for i := range dst {
		t := float64(i) / soundRate
		dst[i] = .22*math.Sin(2*math.Pi*110*t+1.8*math.Sin(2*math.Pi*7*t)) + .10*math.Sin(2*math.Pi*440*t)*(1+.35*math.Sin(2*math.Pi*23*t))
	}
	return pcm(dst)
}

// Fade over the last two beats of the first four-bar phrase, before it repeats.
func makeWaveJingle() []byte {
	full := makeTune(false)
	frames := len(full) / 8
	data := append([]byte(nil), full[:frames*4]...)
	fadeFrames := int(2. * 60 / 168 * soundRate)
	for i := frames - fadeFrames; i < frames; i++ {
		t := float64(i-(frames-fadeFrames)) / float64(fadeFrames-1)
		gain := 1 - t*t*(3-2*t)
		for channel := 0; channel < 2; channel++ {
			at := i*4 + channel*2
			sample := int16(binary.LittleEndian.Uint16(data[at:]))
			binary.LittleEndian.PutUint16(data[at:], uint16(int16(float64(sample)*gain)))
		}
	}
	return data
}

// Four 4/4 bars at 168 BPM, rounded up to ticks, plus two quiet seconds.
const waveBreakTicks = 463
