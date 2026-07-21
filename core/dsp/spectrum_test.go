package dsp

import (
	"math"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
)

// centsOfColumn is the deviation a spectrum column stands for: the inverse of spectrumFor's binning.
func centsOfColumn(c int) float64 {
	const step = 2 * SpectrumCents / SpectrumColumns
	return -SpectrumCents + (float64(c)+0.5)*step
}

func peakColumn(spec []float32) int {
	best := 0
	for c, v := range spec {
		if v > spec[best] {
			best = c
		}
	}
	return best
}

// The spectrum the frontend draws is built here, so a picture whose peak is not where the reed is
// would silently send a technician to file by the wrong line.
func TestSpectrumPeakSitsOnTheReed(t *testing.T) {
	// 22 cents sharp of A4, well inside the +-50 view and not on a column boundary.
	const freq = 445.6

	spec := audio.SynthSpec{
		Duration: 6 * time.Second,
		Reeds:    []audio.ReedSpec{{Freq: freq, Amp: 0.4, Harmonics: []float64{0.3, 0.15}}},
	}
	m := runEngine(t, spec, defaultCfg()).lastFull()

	if len(m.Spectrum) != SpectrumColumns {
		t.Fatalf("spectrum = %d columns, want %d", len(m.Spectrum), SpectrumColumns)
	}
	if len(m.Reeds) != 1 {
		t.Fatalf("reeds = %+v", m.Reeds)
	}

	// One column is 0.39 cents; the smoothing kernel and the reed's lobe are both wider than that.
	want := m.Reeds[0].DevCents
	got := centsOfColumn(peakColumn(m.Spectrum))
	if math.Abs(got-want) > 1.5 {
		t.Fatalf("spectrum peak at %.2f cents, reed at %.2f: the picture disagrees with the reading", got, want)
	}
}

// Heights are what the display multiplies its frame by. Outside 0..1 it draws off the card.
func TestSpectrumHeightsAreNormalised(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 6 * time.Second,
		Reeds:    []audio.ReedSpec{{Freq: 442.6, Amp: 0.4}},
	}
	m := runEngine(t, spec, defaultCfg()).lastFull()

	var top float32
	for c, v := range m.Spectrum {
		if v < 0 || v > 1 {
			t.Fatalf("column %d = %v, outside 0..1", c, v)
		}
		if v > top {
			top = v
		}
	}

	// A tone this clean reaches the reference height: the loudest line in the band IS the reed.
	if top < 0.9 {
		t.Fatalf("tallest column = %v, want the reed at full height", top)
	}
}

// A merged pair is one lobe, and the view must still be drawn around the note (the ruler underneath is
// on the same axis), not around whichever line the peak picker found.
func TestSpectrumCentredOnTheNote(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 6 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 440.0, Amp: 0.4},
			{Freq: 442.0, Amp: 0.4},
		},
	}
	cfg := defaultCfg()
	cfg.ReedCount = 2

	m := runEngine(t, spec, cfg).lastFull()
	if m.ScalePitch <= 0 {
		t.Fatalf("scale pitch = %v", m.ScalePitch)
	}

	peak := centsOfColumn(peakColumn(m.Spectrum))
	if math.Abs(peak) > SpectrumCents {
		t.Fatalf("peak at %.2f cents, outside the +-%v view", peak, SpectrumCents)
	}
}

// The band is drawn long before it is resolved: a short window draws it with fat lobes (a blurred
// picture, not a wrong answer), so it goes out as soon as there is half a second to draw it over.
func TestSpectrumArrivesBeforeTheReading(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 8 * time.Second,
		Reeds:    []audio.ReedSpec{{Freq: 442.6, Amp: 0.4, Harmonics: []float64{0.3, 0.15}}},
	}
	cap := runEngine(t, spec, defaultCfg())

	cap.mu.Lock()
	defer cap.mu.Unlock()

	firstSpectrum, firstReed := -1, -1
	for i, m := range cap.all {
		if firstSpectrum < 0 && len(m.Spectrum) > 0 {
			firstSpectrum = i
		}
		if firstReed < 0 && len(m.Reeds) > 0 {
			firstReed = i
		}
	}

	if firstSpectrum < 0 {
		t.Fatal("no spectrum was ever sent")
	}
	if firstReed < 0 {
		t.Fatal("no reed was ever measured")
	}
	if firstSpectrum >= firstReed {
		t.Fatalf("spectrum at tick %d, reed at tick %d: the picture must not wait for the reading",
			firstSpectrum, firstReed)
	}

	// And a picture never carries numbers.
	for i, m := range cap.all {
		if len(m.Spectrum) > 0 && len(m.Reeds) == 0 {
			if len(m.Beats) > 0 || m.Locked {
				t.Fatalf("tick %d: a spectrum-only frame carried a reading (beats=%d locked=%v)",
					i, len(m.Beats), m.Locked)
			}
			if m.ScalePitch <= 0 {
				t.Fatalf("tick %d: a spectrum with no pitch to draw it around", i)
			}
		}
	}
}
