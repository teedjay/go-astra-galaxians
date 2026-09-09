package game

import "math"

// Original 48-bar arrangement (~74 s): opening, verse, hook, breakdown,
// variation and double chorus. SID-inspired synthesis, not a chip emulator.
func makeSIDTune() []byte {
	const bars = 48
	step := 60. / 156 / 4
	dst := make([]float64, int(bars*16*step*soundRate))
	roots := []int{40, 36, 43, 38} // E minor / C / G / D
	hook := []int{76, 76, 79, 83, 81, 79, 78, 74, 76, 79, 83, 86, 83, 81, 79, 78}
	verse := []int{64, 71, 67, 74, 71, 67, 66, 62, 64, 67, 71, 74, 72, 71, 67, 66}
	bass := []int{0, -1, 12, 0, -1, 7, 0, 12, 0, -1, 12, 7, -1, 0, 10, 12}
	for bar := 0; bar < bars; bar++ {
		root := roots[bar%4]
		chorus := (bar >= 12 && bar < 20) || bar >= 32
		breakdown := bar >= 20 && bar < 24
		for j := 0; j < 16; j++ {
			at := float64(bar*16+j) * step
			if bass[j] >= 0 && (!breakdown || j%4 == 0) {
				addSID(dst, at, step*1.2, root+bass[j], .28, 0)
			}
			if bar >= 4 && !breakdown {
				chord := []int{0, 3, 7}
				if bar%4 != 0 {
					chord = []int{0, 4, 7}
				}
				if j%4 == 0 {
					addSIDArp(dst, at, step*3.7, root+24, chord, .075)
				}
			}
			if bar >= 4 && j%2 == 0 && !breakdown {
				line := verse
				if chorus {
					line = hook
				}
				n := line[(bar%2)*8+j/2]
				if bar >= 40 && j >= 12 {
					n += 12
				}
				gain := .13
				if chorus {
					gain = .19
				}
				addSID(dst, at, step*1.8, n, gain, 1)
			}
			if breakdown && j%4 == 0 {
				addSIDArp(dst, at, step*3.8, root+24, []int{0, 7, 12}, .17)
			}
			if j%4 == 0 {
				kind := 0
				if j%8 == 4 {
					kind = 1
				}
				addDrum(dst, at, kind)
			}
			if j%2 == 1 && !breakdown {
				addDrum(dst, at, 2)
			}
			if bar%8 == 7 && j >= 12 {
				addDrum(dst, at, 1)
			}
		}
	}
	// Preserve mix headroom and feather the loop seam without changing its beat.
	peak := .01
	for _, v := range dst {
		peak = math.Max(peak, math.Abs(v))
	}
	for i := range dst {
		edge := math.Min(1, math.Min(float64(i), float64(len(dst)-1-i))/220)
		dst[i] *= .82 / peak * edge
	}
	return pcm(dst)
}

// PolyBLEP softens oscillator discontinuities while retaining pulse/saw bite.
func blep(phase, dt float64) float64 {
	if phase < dt {
		x := phase / dt
		return 2*x - x*x - 1
	}
	if phase > 1-dt {
		x := (phase - 1) / dt
		return x*x + 2*x + 1
	}
	return 0
}
func sidPulse(phase, dt, width float64) float64 {
	v := -1.
	if phase < width {
		v = 1
	}
	edge := phase - width
	if edge < 0 {
		edge++
	}
	return v + blep(phase, dt) - blep(edge, dt)
}
func addSID(dst []float64, start, duration float64, note int, gain float64, voice int) {
	begin := int(start * soundRate)
	phase, subPhase, filter := 0., 0., 0.
	hz := noteHz(note)
	for i := 0; i < int(duration*soundRate) && begin+i < len(dst); i++ {
		t := float64(i) / soundRate
		global := start + t
		freq := hz * (1 + .006*math.Sin(2*math.Pi*5.5*t)*math.Min(1, t*20))
		if voice == 0 {
			freq *= 1 + .8*math.Exp(-t*110)
		} // percussive bass pitch snap
		dt := freq / soundRate
		phase = math.Mod(phase+dt, 1)
		subPhase = math.Mod(subPhase+dt*.5, 1)
		width := .28 + .19*math.Sin(2*math.Pi*.7*global)
		wave := sidPulse(phase, dt, width)
		if voice == 0 {
			saw := 2*phase - 1 - blep(phase, dt)
			sub := 1 - 4*math.Abs(subPhase-.5)
			filter += (.025 + .35*math.Exp(-t*22)) * (.55*saw + .45*wave - filter)
			wave = .7*filter + .5*sub
		} else {
			// Rapid opening slide, pulse-width motion and ring-colored lead accents.
			wave = .85*wave + .15*wave*math.Sin(2*math.Pi*hz*2*t)
		}
		envelope := math.Min(1, t/.002) * math.Min(1, (duration-t)/.015) * (.65 + .35*math.Exp(-t*12))
		dst[begin+i] += gain * wave * envelope
	}
}
func addSIDArp(dst []float64, start, duration float64, root int, chord []int, gain float64) {
	begin := int(start * soundRate)
	phase := 0.
	for i := 0; i < int(duration*soundRate) && begin+i < len(dst); i++ {
		t := float64(i) / soundRate
		note := root + chord[int(t*50)%len(chord)]
		dt := noteHz(note) / soundRate
		phase = math.Mod(phase+dt, 1)
		envelope := math.Min(1, t/.003) * math.Min(1, (duration-t)/.012)
		dst[begin+i] += sidPulse(phase, dt, .16+.08*math.Sin(t*14)) * gain * envelope
	}
}
