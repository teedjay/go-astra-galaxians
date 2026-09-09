package main

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestSynthesizedAudioIsStereoAndHasHeadroom(t *testing.T) {
	sounds := map[string][]byte{"intro": makeTune(false), "sid": makeSIDTune(), "interlude": makeWaveJingle(), "ending": makeTune(true), "beam": makeBeamHum()}
	for _, name := range []string{"blaster", "rocket", "pulse", "explosion", "huge", "hit", "up", "down", "star"} {
		sounds[name] = makeSound(name)
	}
	for name, data := range sounds {
		if len(data) == 0 || len(data)%4 != 0 {
			t.Fatalf("%s invalid PCM", name)
		}
		peak := 0.
		for i := 0; i < len(data); i += 4 {
			left := int16(binary.LittleEndian.Uint16(data[i:]))
			right := int16(binary.LittleEndian.Uint16(data[i+2:]))
			if left != right {
				t.Fatal("invalid stereo layout")
			}
			peak = math.Max(peak, math.Abs(float64(left)))
		}
		if peak < 100 || peak >= 29000 {
			t.Fatalf("%s silent or clipped: %v", name, peak)
		}
	}
	if len(sounds["sid"]) < 60*soundRate*4 {
		t.Fatal("second intro must be a full arrangement over a minute")
	}
	if len(sounds["ending"]) > soundRate*4*4 {
		t.Fatal("game over jingle too long")
	}
	for _, name := range []string{"intro", "sid", "beam"} {
		data := sounds[name]
		first := int16(binary.LittleEndian.Uint16(data))
		last := int16(binary.LittleEndian.Uint16(data[len(data)-4:]))
		if math.Abs(float64(first)-float64(last)) > 1200 {
			t.Fatalf("%s loop has abrupt seam", name)
		}
	}
}
func TestLaunchFadesBeforeGameplay(t *testing.T) {
	g := testGame()
	g.fadePhase = fadeLaunch
	for i := 0; i < fadeDuration-1; i++ {
		g.updateTransition(false)
		if g.started {
			t.Fatal("game started before music fade finished")
		}
	}
	g.updateTransition(false)
	if !g.started || g.fadePhase != 0 || g.over {
		t.Fatal("launch fade did not start clean game")
	}
}

func TestWaveJingleFitsBreak(t *testing.T) {
	data := makeWaveJingle()
	duration := float64(len(data)) / 4 / soundRate
	if duration < 11 || duration >= 12 {
		t.Fatal("wave jingle must play the complete original song")
	}
	if float64(waveBreakTicks)/60-duration < 2 || float64(waveBreakTicks)/60-duration > 2.02 {
		t.Fatal("next wave must wait two seconds after jingle")
	}
}
