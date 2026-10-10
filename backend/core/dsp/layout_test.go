package dsp

import (
	"math"
	"slices"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

func TestEngineReedMaskNamesTheBoxes(t *testing.T) {
	m := lastMeasured(t, runSynth(t, EngineConfig{Reeds: []int{1, 3}, BeatsInHertz: true}, 6, 438.9, 441.8))
	if len(m.Reeds) != 2 || m.Reeds[0].Reed != 1 || m.Reeds[1].Reed != 3 || len(m.Beats) != 0 {
		t.Fatalf("boxes 1 and 3: reeds %+v beats %+v", m.Reeds, m.Beats)
	}
	if len(m.Panels) != 3 || !m.Panels[1].Absent || m.Panels[1].Shown != "-" || math.Abs(m.Reeds[1].Freq-441.8) > 0.05 {
		t.Fatalf("panels %+v", m.Panels)
	}
	m = lastMeasured(t, runSynth(t, EngineConfig{Reeds: []int{2, 3}, BeatsInHertz: true}, 6, 439.0, 441.5))
	if len(m.Reeds) != 2 || m.Reeds[0].Reed != 2 || len(m.Beats) != 1 || m.Beats[0].Pair != "1-2" || math.Abs(m.Beats[0].Hz-2.5) > 0.1 {
		t.Fatalf("boxes 2 and 3: reeds %+v beats %+v", m.Reeds, m.Beats)
	}
	if m.Panels[1].Absent || math.Abs(m.Panels[1].Value-m.Beats[0].Hz) > 1e-9 {
		t.Fatalf("pair 2-3 panel %+v", m.Panels[1])
	}
}

func TestReedMaskFillsTheCount(t *testing.T) {
	c := EngineConfig{ReedCount: 5, Reeds: []int{3, 1, 1, 4}}
	c.fill()
	if !slices.Equal(c.Reeds, []int{1, 3}) || c.ReedCount != 2 || !slices.Equal(c.meterSlots(), []int{0, 2}) {
		t.Fatalf("mask %v count %d slots %v", c.Reeds, c.ReedCount, c.meterSlots())
	}
	var curve meter.NoteCurve
	curve[57-4] = 7
	c.Targets = []meter.NoteCurve{{}, curve}
	if got := c.targets(); len(got.Slots) != 3 || got.Slots[2][57-4] != 7 || got.Slots[1][57-4] != meter.Missing {
		t.Fatalf("box 3's curve must sit on meter slot 3: %v", got.Slots)
	}
}

// T = 2: the instrument's A sounds as B; the band follows the sounding pitch, the label stays the key.
func TestEngineTransposeMovesTheBands(t *testing.T) {
	sounding := 440 * math.Pow(2, 2.0/12)
	m := lastMeasured(t, runSynth(t, EngineConfig{Transpose: 2}, 6, sounding+0.5))
	if m.Note != tuning.NoteA4 || m.NoteName != "A4" || math.Abs(m.ScalePitch-sounding) > 1e-9 {
		t.Fatalf("note %d %q at %v", m.Note, m.NoteName, m.ScalePitch)
	}
	if want := (0.5 / sounding) / 0.0006; math.Abs(m.Reeds[0].Dev-want) > 0.05 {
		t.Fatalf("deviation %v want %v", m.Reeds[0].Dev, want)
	}
	c := EngineConfig{Transpose: -3}
	c.fill()
	if c.Transpose != 0 || meter.EqualTemperament(440, 0) != c.table() {
		t.Fatal("the range is 0..11 and T = 0 is the plain table")
	}
}

func TestEngineChordNameHasNoOctave(t *testing.T) {
	c4, e4, g4 := 440*math.Pow(2, -9.0/12), 440*math.Pow(2, -5.0/12), 440*math.Pow(2, -2.0/12)
	m := lastMeasured(t, runSynth(t, EngineConfig{Chords: true, Scale: tuning.NamingPolish}, 6, c4, e4, g4))
	if m.Label != "C" || len(m.Beats) != 0 || len(m.Spectra) != 3 || len(m.Spectra[0]) != BandSpectrumColumns {
		t.Fatalf("label %q beats %+v spectra %d", m.Label, m.Beats, len(m.Spectra))
	}
	for i, p := range m.Panels {
		if i%2 == 1 && !p.Absent {
			t.Fatalf("chords have no pair readouts: %+v", p)
		}
	}
	want := map[meter.ChordKind]string{meter.ChordMajor: "", meter.ChordMinor: "m", meter.ChordDiminished: "dim",
		meter.ChordAugmented: "+", meter.ChordMajorFlatFifth: "-5", meter.ChordSuspendedSecond: "sus2", meter.ChordSuspendedFourth: "sus4"}
	for kind, suffix := range want {
		if chordSuffix[kind] != suffix {
			t.Errorf("kind %d: %q want %q", kind, chordSuffix[kind], suffix)
		}
	}
	polish := EngineConfig{Chords: true, Scale: tuning.NamingPolish}
	if got := polish.analyzer(48000).ChordLabel(meter.ChordIdentity{Root: 58, Kind: meter.ChordMinor}); got != "Aism" {
		t.Fatalf("A# minor in Polish naming: %q", got)
	}
}

// Octave beat readouts of the reference Bass mode show only in Hz, while the beats stay in the data;
// the skipped-octave layout names boxes 1 and 3.
func TestEngineBassOctaveBeatReadoutsOnlyInHertz(t *testing.T) {
	feet := []OctaveRequest{{0, 1}, {12, 1}, {24, 1}}
	units := lastMeasured(t, runSynth(t, EngineConfig{Bass: true, Octaves: feet}, 6, 110.1, 220.3, 440.5))
	hertz := lastMeasured(t, runSynth(t, EngineConfig{Bass: true, Octaves: feet, BeatsInHertz: true}, 6, 110.1, 220.3, 440.5))
	if len(units.Beats) != 2 || !units.Beats[0].Present || !units.Panels[1].Absent || !units.Panels[3].Absent {
		t.Fatalf("units: beats %+v panels %+v", units.Beats, units.Panels)
	}
	if len(hertz.Beats) != 2 || !hertz.Beats[0].Present || hertz.Panels[1].Absent || len(hertz.Spectra) != 3 {
		t.Fatalf("hertz: beats %+v panels %+v", hertz.Beats, hertz.Panels)
	}
	skip := lastMeasured(t, runSynth(t, EngineConfig{Bass: true, Octaves: []OctaveRequest{{0, 1}, {24, 1}}, BeatsInHertz: true}, 6, 110.1, 440.5))
	r := skip.Reeds
	if len(r) != 2 || r[0].Reed != 1 || r[1].Reed != 3 || r[1].Octave != 24 || !r[1].Present || len(skip.Beats) != 0 || len(skip.Spectra) != 2 {
		t.Fatalf("skipped octave: reeds %+v beats %+v", r, skip.Beats)
	}
	if math.Abs(r[1].Freq-440.5) > 0.05 || !skip.Panels[1].Absent {
		t.Fatalf("base+24 reads %+v, panel %+v", r[1], skip.Panels[1])
	}
}
