package tests

// Properties that must hold of the engine on real audio whatever the recording is. Unlike the rest
// of this package these assert nothing about pitch: each compares the engine against itself, so
// there is no hand-measured expectation to go stale, and each covers something the synthetic golden
// tests cannot - a real spectrum with a real noise floor, real partials and a bellows that moves.

import (
	"context"
	"math"
	"testing"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/dsp"
)

// One of each shape the engine handles: a lone rank, three 8' ranks through the single band with
// its pair splitter, and an octave-spanning register through the compound stage.
var propertyTakes = []struct {
	file    string
	reeds   int
	a4      float64
	octaves []dsp.OctaveRequest
}{
	{file: "A '8.wav", reeds: 1, a4: loneRankA4},
	{file: "hohner8+8+8A4.wav", reeds: 3, a4: 440},
	{file: "hohner16+8A4.wav", reeds: 2, a4: 440,
		octaves: []dsp.OctaveRequest{{Offset: -12, Reeds: 1}, {Offset: 0, Reeds: 1}}},
}

func runFile(t *testing.T, file string, reeds int, a4, ppm float64, octaves []dsp.OctaveRequest) []dsp.Measurement {
	t.Helper()
	src, err := audio.NewFileSource("../sounds/"+file, false, false)
	if err != nil {
		t.Skipf("recording not present: %v", err)
	}
	var all []dsp.Measurement
	eng := dsp.NewEngine(dsp.EngineConfig{
		A4:         a4,
		ReedCount:  reeds,
		ClockPPM:   ppm,
		Octaves:    octaves,
		FineWindow: fineWindow,
	}, func(m dsp.Measurement) { all = append(all, m) })
	if err := eng.Run(context.Background(), src); err != nil {
		t.Fatalf("engine: %v", err)
	}
	return all
}

// The same recording twice must measure the same. The app builds a new engine per pass but the
// process is long-lived, and a reading that depended on what had been played before it would be
// invisible at a bench: every number would look plausible and only the second opinion would differ.
func TestTheSameRecordingMeasuresTheSameTwice(t *testing.T) {
	if testing.Short() {
		t.Skip("loads multi-megabyte recordings")
	}
	for _, take := range propertyTakes {
		t.Run(take.file, func(t *testing.T) {
			first := runFile(t, take.file, take.reeds, take.a4, 0, take.octaves)
			second := runFile(t, take.file, take.reeds, take.a4, 0, take.octaves)

			if len(first) != len(second) {
				t.Fatalf("%d ticks then %d: the engine's cadence is not reproducible",
					len(first), len(second))
			}
			for i := range first {
				a, b := first[i], second[i]
				if a.Note != b.Note || a.State != b.State || len(a.Reeds) != len(b.Reeds) {
					t.Fatalf("tick %d differs: %s/%s %d reeds vs %s/%s %d reeds",
						i, a.Note.Name(0), a.State, len(a.Reeds),
						b.Note.Name(0), b.State, len(b.Reeds))
				}
				// Bit for bit: the arithmetic is the same arithmetic over the same samples, so
				// anything less would mean state carried across the two runs.
				for r := range a.Reeds {
					if a.Reeds[r].Freq != b.Reeds[r].Freq {
						t.Fatalf("tick %d reed %d: %.12f then %.12f",
							i, r+1, a.Reeds[r].Freq, b.Reeds[r].Freq)
					}
				}
			}
		})
	}
}

// The clock correction is the one setting a technician turns when they do not trust their sound
// card, and it has to be a pure scaling: every reed of every reading moved by exactly the parts per
// million asked for, or the beats between them change and the correction has invented a tuning
// error. Tested on a real spectrum rather than a synthesised tone, because the correction is applied
// to a number that came out of peak interpolation, not out of the oscillator that made the signal.
func TestTheClockCorrectionScalesEveryReedByExactlyThePPM(t *testing.T) {
	if testing.Short() {
		t.Skip("loads multi-megabyte recordings")
	}
	const ppm = 250.0

	for _, take := range propertyTakes {
		t.Run(take.file, func(t *testing.T) {
			plain := runFile(t, take.file, take.reeds, take.a4, 0, take.octaves)
			shifted := runFile(t, take.file, take.reeds, take.a4, ppm, take.octaves)

			if len(plain) != len(shifted) {
				t.Fatalf("%d ticks uncorrected, %d corrected", len(plain), len(shifted))
			}

			want := 1 - ppm*1e-6
			compared := 0
			for i := range plain {
				a, b := plain[i], shifted[i]
				if len(a.Reeds) == 0 || len(a.Reeds) != len(b.Reeds) {
					continue // the correction may move a tick across the detector's own thresholds
				}
				for r := range a.Reeds {
					compared++
					got := b.Reeds[r].Freq / a.Reeds[r].Freq
					// A part in ten million, on a ratio of two doubles.
					if math.Abs(got-want) > 1e-7 {
						t.Fatalf("tick %d reed %d: %.6f Hz became %.6f Hz, a factor of %.9f, want %.9f",
							i, r+1, a.Reeds[r].Freq, b.Reeds[r].Freq, got, want)
					}
				}
			}
			if compared == 0 {
				t.Fatal("no reed was measured under both settings")
			}
			t.Logf("%d reed readings scaled by exactly %g ppm", compared, ppm)
		})
	}
}

// dsp.Aggregate is what turns a run into the one reading that gets filed and printed. It votes on
// the note and takes per-index MEDIANS of the reeds, so every number it hands back has to be one the
// engine actually measured - a median never leaves the range of its inputs. A reading interpolated
// into existence between two ticks would be a frequency nobody heard, filed as though somebody had.
func TestTheFiledReadingIsOneTheEngineActuallyMeasured(t *testing.T) {
	if testing.Short() {
		t.Skip("loads multi-megabyte recordings")
	}
	for _, take := range propertyTakes {
		t.Run(take.file, func(t *testing.T) {
			all := runFile(t, take.file, take.reeds, take.a4, 0, take.octaves)
			m, ok := dsp.Aggregate(all, take.a4, 0.5)
			if !ok {
				t.Fatalf("no usable reading")
			}
			if len(m.Reeds) == 0 {
				t.Fatalf("the filed reading carries no reed")
			}

			// The ticks the aggregate drew from: its own note, its own reed count. Aggregate also
			// filters by state and level, so this is a superset - a median of the subset still
			// cannot leave the range of the superset.
			var contributing []dsp.Measurement
			for _, tick := range all {
				if tick.Note == m.Note && len(tick.Reeds) == len(m.Reeds) {
					contributing = append(contributing, tick)
				}
			}
			if len(contributing) == 0 {
				t.Fatalf("the filed reading names %s with %d reeds, which no tick ever reported",
					m.Note.Name(0), len(m.Reeds))
			}

			for r := range m.Reeds {
				lo, hi := math.Inf(1), math.Inf(-1)
				for _, tick := range contributing {
					lo = math.Min(lo, tick.Reeds[r].Freq)
					hi = math.Max(hi, tick.Reeds[r].Freq)
				}
				if f := m.Reeds[r].Freq; f < lo || f > hi {
					t.Errorf("reed %d was filed at %.4f Hz, but across %d ticks the engine only ever "+
						"measured %.4f to %.4f", r+1, f, len(contributing), lo, hi)
				}
			}

			if m.Locked {
				locked := false
				for _, tick := range contributing {
					if tick.Locked {
						locked = true
						break
					}
				}
				if !locked {
					t.Error("the filed reading claims a lock no tick reported")
				}
			}
			t.Logf("%s: %d reeds, each within the range of %d ticks",
				m.Note.Name(0), len(m.Reeds), len(contributing))
		})
	}
}
