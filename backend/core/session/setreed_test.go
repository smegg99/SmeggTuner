package session

import (
	"math"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// A hand edit rewrites the slot at its own pitch and re-reads the beats it touches.
func TestSetReedKeepsTheTakeConsistent(t *testing.T) {
	tk := Take{Note: 69, Reeds: []dsp.ReedMeasure{
		{Present: true, Dev: -6, Octave: -12, Freq: 219.2}, {Present: true, Dev: 0, Freq: 440}, {Dev: -99},
	}}
	if !tk.SetReed(0, 2, 440, 0) {
		t.Fatal("a sounding slot was refused")
	}
	if want := 220 * (1 + 0.0006*2); math.Abs(tk.Reeds[0].Freq-want) > 1e-9 || tk.Reeds[0].Dev != 2 {
		t.Fatalf("slot 1 = %+v, want 2 units of the 16' pitch (%v Hz)", tk.Reeds[0], want)
	}
	if !tk.Manual {
		t.Fatal("a hand edit is not marked manual")
	}
	if len(tk.Beats) != 2 || !tk.Beats[0].Present || tk.Beats[1].Present {
		t.Fatalf("beats %+v, want 1-2 re-read and 2-3 absent around the silent slot", tk.Beats)
	}
	if want := math.Abs(440 - 2*tk.Reeds[0].Freq); math.Abs(tk.Beats[0].Hz-want) > 1e-9 {
		t.Fatalf("octave beat %v Hz, want %v", tk.Beats[0].Hz, want)
	}

	for _, slot := range []int{-1, 2, 3} {
		if tk.SetReed(slot, 1, 440, 0) {
			t.Fatalf("slot %d was written: out of range or never heard", slot)
		}
	}
}

// A transposed take is rewritten at the pitch it sounded, not at its key.
func TestSetReedOnATransposedTake(t *testing.T) {
	sounding := tuning.Note(72).Freq(440)
	tk := Take{Note: 69, Reeds: []dsp.ReedMeasure{{Present: true, Freq: sounding}}}
	if !tk.SetReed(0, 5, 440, 3) {
		t.Fatal("a sounding slot was refused")
	}
	if want := sounding * (1 + 0.0006*5); math.Abs(tk.Reeds[0].Freq-want) > 1e-9 {
		t.Fatalf("freq = %v, want %v: 5 units of C5, the A4 key three semitones up", tk.Reeds[0].Freq, want)
	}
	if tk.Note != 69 {
		t.Fatalf("note = %d, the take keeps its key", tk.Note)
	}
}
