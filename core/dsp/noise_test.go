package dsp

import "testing"

func flat(v float64) CoarseResult {
	var r CoarseResult
	for i := range r.NoteEnergy {
		r.NoteEnergy[i] = v
	}
	return r
}

func TestNoiseCalibrationThenSignal(t *testing.T) {
	nf := NewNoiseFloor(5)
	for i := 0; i < 50; i++ { // 5 s of 0.1 s ticks with ambient 0.001
		nf.Update(flat(0.001), 0.1)
	}
	if nf.Calibrating() {
		t.Fatal("still calibrating after 5s")
	}
	floorBefore := nf.Floor()[50]
	// 10 s of loud playing must not fold the tone into the floor
	for i := 0; i < 100; i++ {
		nf.Update(flat(0.5), 0.1)
	}
	if f := nf.Floor()[50]; f > 0.15 {
		t.Fatalf("floor rose too fast during playing: %v (was %v)", f, floorBefore)
	}
}

func TestNoiseAbsorbsSustainedBackground(t *testing.T) {
	nf := NewNoiseFloor(1)
	for i := 0; i < 10; i++ {
		nf.Update(flat(0.001), 0.1)
	}
	// 3 minutes of constant 0.2 background (e.g. a fan turned on)
	for i := 0; i < 1800; i++ {
		nf.Update(flat(0.2), 0.1)
	}
	if f := nf.Floor()[50]; f < 0.1 {
		t.Fatalf("sustained background not absorbed: floor %v", f)
	}
}
