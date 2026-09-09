package main

import (
	"math"
	"testing"
)

func TestBezierEndpoints(t *testing.T) {
	a, b, c, d := point{10, 20}, point{-200, 400}, point{600, 900}, point{300, 100}
	if bezier(a, b, c, d, 0) != a || bezier(a, b, c, d, 1) != d {
		t.Fatal("curve must join its endpoints exactly")
	}
	prev := a
	for i := 1; i <= 1000; i++ {
		p := bezier(a, b, c, d, float64(i)/1000)
		if math.Hypot(p.x-prev.x, p.y-prev.y) > 4 {
			t.Fatal("discontinuous flight path")
		}
		prev = p
	}
}
func TestRotationTakesShortestPath(t *testing.T) {
	d := angleDelta(math.Pi-.01, -math.Pi+.01)
	if math.Abs(d-.02) > 1e-9 {
		t.Fatalf("rotation crosses wrap incorrectly: %v", d)
	}
}
func TestFormationsKeepShipsApart(t *testing.T) {
	for w := 0; w < 3; w++ {
		for i := 0; i < waveSize; i++ {
			p := formation(w, i)
			if p.x < 25 || p.x > W-25 || p.y < 90 || p.y > 400 {
				t.Fatal("formation outside playfield")
			}
			for j := 0; j < i; j++ {
				q := formation(w, j)
				if math.Hypot(p.x-q.x, p.y-q.y) < 35 {
					t.Fatal("overlapping ships")
				}
			}
		}
	}
}
