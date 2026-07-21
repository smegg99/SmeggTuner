package dsp

import (
	"math"
	"testing"

	"smegg.me/smeggtuner/core/tuning"
)

func fillRing(freqs []float64, amps []float64, sr int, seconds float64) *Ring {
	n := int(float64(sr) * seconds)
	r := NewRing(n)
	buf := make([]float32, n)
	for i := 0; i < n; i++ {
		var v float64
		for j, f := range freqs {
			v += amps[j] * math.Sin(2*math.Pi*f*float64(i)/float64(sr))
		}
		buf[i] = float32(v)
	}
	r.Write(buf)
	return r
}

func TestCoarseFindsA4(t *testing.T) {
	r := fillRing([]float64{440}, []float64{0.5}, 48000, 3)
	c := NewCoarse(48000, 440)
	res := c.Analyze(r)
	idxA4 := int(tuning.NoteA4 - tuning.MinNote)
	if res.NoteEnergy[idxA4] < 0.3 {
		t.Fatalf("A4 energy %v too low", res.NoteEnergy[idxA4])
	}
	if res.NoteEnergy[idxA4-1] > res.NoteEnergy[idxA4]/3 ||
		res.NoteEnergy[idxA4+1] > res.NoteEnergy[idxA4]/3 {
		t.Fatalf("neighbors too high: %v %v %v",
			res.NoteEnergy[idxA4-1], res.NoteEnergy[idxA4], res.NoteEnergy[idxA4+1])
	}
}

func TestCoarseLowNote(t *testing.T) {
	f := tuning.Note(28).Freq(440) // E1 = 41.2 Hz
	r := fillRing([]float64{f}, []float64{0.5}, 48000, 3)
	c := NewCoarse(48000, 440)
	res := c.Analyze(r)
	idx := 28 - int(tuning.MinNote)
	if res.NoteEnergy[idx] < 0.3 {
		t.Fatalf("E1 energy %v too low", res.NoteEnergy[idx])
	}
	if res.NoteEnergy[idx-1] > res.NoteEnergy[idx]/2 || res.NoteEnergy[idx+1] > res.NoteEnergy[idx]/2 {
		t.Fatalf("E1 neighbors not separated: %v %v %v",
			res.NoteEnergy[idx-1], res.NoteEnergy[idx], res.NoteEnergy[idx+1])
	}
}
