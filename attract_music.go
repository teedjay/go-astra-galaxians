package main

import "math"

// "Orbit Runner": a 102-second original with a sliding bass ostinato,
// portamento hooks, PWM arpeggios, a sparse bridge and a double final chorus.
func makeAttractTune() []byte {
	step := 60. / 150 / 4
	dst := make([]float64, int(64*16*step*soundRate))
	roots := []int{38, 34, 41, 36, 38, 34, 43, 36}
	hook := []int{74, 77, 81, 79, 77, 74, 72, 69, 74, 77, 81, 84, 81, 79, 77, 76}
	answer := []int{77, 81, 86, 84, 81, 79, 77, 74, 76, 79, 84, 81, 79, 76, 72, 73}
	bassPattern := []int{0, 0, 12, 7, 0, 12, 10, 7}
	prevBass, prevLead := 38, 74
	for bar := 0; bar < 64; bar++ {
		root := roots[bar%8]
		chorus := bar >= 16 && bar < 32 || bar >= 48
		bridge := bar >= 32 && bar < 40
		for j := 0; j < 16; j++ {
			at := float64(bar*16+j) * step
			if j%2 == 0 {
				n := root + bassPattern[j/2]
				addGlideVoice(dst, at, step*1.95, prevBass, n, .28, true)
				prevBass = n
				if bar >= 8 && !bridge {
					line := hook
					if bar%4 >= 2 {
						line = answer
					}
					n := line[(bar%2)*8+j/2]
					if !chorus {
						n -= 12
					}
					addGlideVoice(dst, at, step*1.9, prevLead, n, .17, false)
					prevLead = n
				}
			}
			if j%4 == 0 && (chorus || bridge || bar < 8) {
				chord := []int{0, 3, 7}
				if bar%8 != 0 && bar%8 != 4 {
					chord = []int{0, 4, 7}
				}
				gain := .065
				if bridge {
					gain = .13
				}
				addSIDArp(dst, at, step*3.8, root+24, chord, gain)
			}
			if j%4 == 0 {
				kind := 0
				if j%8 == 4 {
					kind = 1
				}
				addDrum(dst, at, kind)
			}
			if j%2 == 1 && !bridge {
				addDrum(dst, at, 2)
			}
			if bar%8 == 7 && j >= 12 {
				addDrum(dst, at, 1)
			}
		}
	}
	peak := .01
	for _, v := range dst {
		peak = math.Max(peak, math.Abs(v))
	}
	for i := range dst {
		edge := math.Min(1, math.Min(float64(i)/800, float64(len(dst)-1-i)/soundRate))
		dst[i] *= .84 / peak * edge
	}
	return pcm(dst)
}
func addGlideVoice(dst []float64, start, duration float64, from, to int, gain float64, bass bool) {
	begin := int(start * soundRate)
	phase, sub, filter := 0., 0., 0.
	for i := 0; i < int(duration*soundRate) && begin+i < len(dst); i++ {
		t := float64(i) / soundRate
		glide := math.Min(1, t/.095)
		glide = glide * glide * (3 - 2*glide)
		pitch := float64(from) + (float64(to-from))*glide
		hz := 440 * math.Pow(2, (pitch-69)/12) * (1 + .004*math.Sin(t*34))
		dt := hz / soundRate
		phase = math.Mod(phase+dt, 1)
		sub = math.Mod(sub+dt*.5, 1)
		width := .27 + .20*math.Sin((start+t)*3.1)
		wave := sidPulse(phase, dt, width)
		if bass {
			saw := 2*phase - 1 - blep(phase, dt)
			cutoff := .035 + .3*math.Exp(-t*15)
			filter += cutoff * (.6*saw + .4*wave - filter)
			wave = .75*filter + .5*(1-4*math.Abs(sub-.5))
		} else {
			wave = .8*wave + .12*math.Sin(2*math.Pi*sub)
		}
		env := math.Min(1, t/.003) * math.Min(1, (duration-t)/.012)
		dst[begin+i] += wave * gain * env
	}
}
