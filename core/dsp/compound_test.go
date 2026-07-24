package dsp

import (
	"math"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/tuning"
)

// lmConfig declares a 16'+8' register on the A4 key.
func lmConfig() EngineConfig {
	return EngineConfig{
		A4:         440,
		ReedCount:  2,
		FineWindow: 3 * time.Second,
		Octaves:    []OctaveRequest{{Offset: -12, Reeds: 1}, {Offset: 0, Reeds: 1}},
	}
}

// A 16'+8' register: the 16' at 219 lays its second partial at 438, the genuine 8' sits at 440.5.
// The engine must name the key A4, read each rank against its own octave, and put the audible beat -
// the 8' against the 16's partial - in front of the technician.
func TestCompoundEngineReadsA16Plus8(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 8 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 219, Amp: 0.40, Harmonics: []float64{0.30, 0.15}},
			{Freq: 440.5, Amp: 0.45, Harmonics: []float64{0.20}},
		},
		NoiseAmp: 1e-4,
	}
	cap := runEngine(t, spec, lmConfig())
	m, ok := Aggregate(cap.all, 440, 0.5)
	if !ok {
		t.Fatal("no usable measurements")
	}

	if m.Note != tuning.Note(69) {
		t.Errorf("note %d, want A4 (69): the key, not the 16' pitch", int(m.Note))
	}
	if len(m.Reeds) != 2 {
		t.Fatalf("resolved %d reeds, want 2: %+v", len(m.Reeds), m.Reeds)
	}
	if m.Reeds[0].Octave != -12 || math.Abs(m.Reeds[0].Freq-219) > 0.1 {
		t.Errorf("16' reed: want ~219 at octave -12, got %+v", m.Reeds[0])
	}
	if m.Reeds[1].Octave != 0 || math.Abs(m.Reeds[1].Freq-440.5) > 0.1 {
		t.Errorf("8' reed: want ~440.5 at octave 0, got %+v", m.Reeds[1])
	}
	// DevCents against each rank's own pitch: the 16' at 219 is ~7.9 cents flat of 220.
	if math.Abs(m.Reeds[0].DevCents-tuning.Cents(219, 220)) > 0.5 {
		t.Errorf("16' DevCents %v, want vs its own octave (220)", m.Reeds[0].DevCents)
	}
	if len(m.Beats) != 1 || math.Abs(m.Beats[0].Hz-2.5) > 0.2 {
		t.Errorf("want the octave beat 8' vs 2*16' = +2.5 Hz, got %+v", m.Beats)
	}
	if len(m.Bands) != 2 || m.Bands[0].Found != 1 || m.Bands[1].Found != 1 {
		t.Errorf("bands: want both ranks found, got %+v", m.Bands)
	}
}

// A blocked 8': only the 16' sounds, its second partial standing exactly where the 8' would. The
// partial is phase-locked to its fundamental, so the engine must refuse it - one reed, a GhostOnly
// 8' band, and no phantom second voice. This is the case the detector used to get wrong.
func TestCompoundEngineRefusesABlockedRank(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 8 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 219, Amp: 0.40, Harmonics: []float64{0.50, 0.15}},
		},
		NoiseAmp: 1e-4,
	}
	cap := runEngine(t, spec, lmConfig())
	m, ok := Aggregate(cap.all, 440, 0.5)
	if !ok {
		t.Fatal("no usable measurements")
	}

	if len(m.Reeds) != 1 || m.Reeds[0].Octave != -12 {
		t.Fatalf("want only the 16' reed, got %+v", m.Reeds)
	}
	if len(m.Beats) != 0 {
		t.Errorf("no second voice, no beat; got %+v", m.Beats)
	}
	if len(m.Bands) != 2 || !m.Bands[1].GhostOnly || m.Bands[1].Found != 0 {
		t.Errorf("8' band must report GhostOnly with nothing found, got %+v", m.Bands)
	}
}

// An 8' tuned within a lobe of the 16's partial: one merged line in the band, but the reed's own
// phase unlocks it from the fundamental, so it is kept - and the beat, too slow to split, is read
// off the band's envelope.
func TestCompoundEngineKeepsANearCoincidentRank(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 8 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 219, Amp: 0.40, Harmonics: []float64{0.30, 0.15}},
			{Freq: 438.8, Amp: 0.45},
		},
		NoiseAmp: 1e-4,
	}
	cap := runEngine(t, spec, lmConfig())
	m, ok := Aggregate(cap.all, 440, 0.5)
	if !ok {
		t.Fatal("no usable measurements")
	}

	if len(m.Reeds) != 2 {
		t.Fatalf("the 8' must survive sitting near the 16's partial, got %+v", m.Reeds)
	}
	if len(m.Beats) != 1 || !m.Beats[0].FromEnvelope || math.Abs(m.Beats[0].Hz-0.8) > 0.2 {
		t.Errorf("want the ~0.8 Hz octave beat off the envelope, got %+v", m.Beats)
	}
}

// A musette cluster too close to split, with partials outweighing the fundamentals: the partials
// resolve into clean lines an octave up and once read as phantom voices of the wrong key. The
// cluster's own envelope beat gives them away - partials of reeds beating at env sit 2*env apart -
// so the engine must keep the true key, report the merged cluster, and read its beat off the envelope.
func TestCompoundEngineRefusesAPartialImage(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 8 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 146.0, Amp: 0.10, Harmonics: []float64{3.0}},
			{Freq: 147.0, Amp: 0.10, Harmonics: []float64{3.0}},
			{Freq: 148.0, Amp: 0.10, Harmonics: []float64{3.0}},
		},
		NoiseAmp: 1e-4,
	}
	cfg := EngineConfig{
		A4:         440,
		ReedCount:  4,
		FineWindow: 3 * time.Second,
		Octaves:    []OctaveRequest{{Offset: -12, Reeds: 1}, {Offset: 0, Reeds: 3}},
	}
	cap := runEngine(t, spec, cfg)
	m, ok := Aggregate(cap.all, 440, 0.5)
	if !ok {
		t.Fatal("no usable measurements")
	}

	if m.Note != tuning.Note(50) {
		t.Errorf("note %s, want D3: the partials an octave up are not the key", m.Note.Name(tuning.NamingCDEFGAB))
	}
	for _, r := range m.Reeds {
		if r.Octave == 0 && r.Freq > 200 {
			t.Errorf("a partial reported as a voice: %+v", r)
		}
	}
	env := false
	for _, b := range m.Beats {
		if b.FromEnvelope && math.Abs(b.Hz-1.0) < 0.2 {
			env = true
		}
	}
	if !env {
		t.Errorf("the cluster's ~1 Hz beat must be read off the envelope, got %+v", m.Beats)
	}
}

// A five-voice bass machine: one rank per octave from F1 up, the way a Stradella bass stacks them.
// The engine must resolve every voice against its own octave and lay an octave beat between each
// neighbouring pair - the same measurement as 16+8+4, only taller and five times lower.
func TestCompoundEngineReadsAFiveVoiceBass(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 8 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 43.7, Amp: 0.07, Harmonics: []float64{0.30}},
			{Freq: 87.6, Amp: 0.10, Harmonics: []float64{0.25}},
			{Freq: 174.2, Amp: 0.13, Harmonics: []float64{0.20}},
			{Freq: 349.9, Amp: 0.15},
			{Freq: 698.0, Amp: 0.12},
		},
		NoiseAmp: 1e-4,
	}
	cfg := EngineConfig{
		A4:         440,
		ReedCount:  5,
		FineWindow: 3 * time.Second,
		Octaves: []OctaveRequest{
			{Offset: 0, Reeds: 1}, {Offset: 12, Reeds: 1}, {Offset: 24, Reeds: 1},
			{Offset: 36, Reeds: 1}, {Offset: 48, Reeds: 1},
		},
	}
	cap := runEngine(t, spec, cfg)
	m, ok := Aggregate(cap.all, 440, 0.5)
	if !ok {
		t.Fatal("no usable measurements")
	}

	if m.Note != tuning.Note(29) {
		t.Errorf("note %s, want F1: the lowest sounding rank", m.Note.Name(tuning.NamingCDEFGAB))
	}
	if len(m.Reeds) != 5 {
		t.Fatalf("resolved %d voices, the machine sounds 5: %+v", len(m.Reeds), m.Reeds)
	}
	wantOcts := []int{0, 12, 24, 36, 48}
	wantFreqs := []float64{43.7, 87.6, 174.2, 349.9, 698.0}
	for i, r := range m.Reeds {
		if r.Octave != wantOcts[i] || math.Abs(r.Freq-wantFreqs[i]) > 0.2 {
			t.Errorf("voice %d: got %+v, want ~%.1f at +%d", i+1, r, wantFreqs[i], wantOcts[i])
		}
	}
	if len(m.Beats) != 4 {
		t.Errorf("want an octave beat between each neighbouring pair, got %+v", m.Beats)
	}
}

// HarmonicPLV itself: a rank's own partial locks near 1, an independent reed does not.
func TestHarmonicPLVSeparatesPartialFromReed(t *testing.T) {
	const sr = 48000
	window := 3 * time.Second
	analyze := func(spec audio.SynthSpec) (ZoomResult, ZoomResult, float64) {
		spec.SampleRate = sr
		spec.Duration = 4 * time.Second
		ring := synthRing(t, spec, window)
		z := NewZoom(sr)
		lo := z.Analyze(ring, 220, 16, window)
		hi := z.Analyze(ring, 440, 16, window)
		if !lo.Valid || !hi.Valid {
			t.Fatal("bands not valid")
		}
		f := FindPeaks(lo, 1, 1.0)
		if len(f) != 1 {
			t.Fatalf("want the 16' line, got %+v", f)
		}
		return lo, hi, f[0].Freq
	}

	lo, hi, f := analyze(audio.SynthSpec{
		Reeds:    []audio.ReedSpec{{Freq: 219, Amp: 0.4, Harmonics: []float64{0.3}}},
		NoiseAmp: 1e-4,
	})
	if plv := HarmonicPLV(lo, f, 2, hi); plv < plvLocked {
		t.Errorf("a lone partial must lock, plv %.3f", plv)
	}

	lo, hi, f = analyze(audio.SynthSpec{
		Reeds: []audio.ReedSpec{
			{Freq: 219, Amp: 0.4, Harmonics: []float64{0.3}},
			{Freq: 440.5, Amp: 0.45},
		},
		NoiseAmp: 1e-4,
	})
	if plv := HarmonicPLV(lo, f, 2, hi); plv >= plvLocked {
		t.Errorf("an independent reed must unlock the band, plv %.3f", plv)
	}
}
