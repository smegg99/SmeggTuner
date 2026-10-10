// backend/core/dsp/build_test.go
package dsp

import (
	"math"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

func TestMeasurementRejectsKeysOutsideTheSupportedRange(t *testing.T) {
	for _, tc := range []struct {
		name   string
		offset int
		note   int
	}{
		{"reported crash", -12, 103},
		{"above highest key", -12, 97},
		{"below lowest key", 12, 15},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEngine(EngineConfig{Octaves: []OctaveRequest{{tc.offset, 1}, {tc.offset + 12, 2}}}, nil)
			r := newRunner(e.cfg, 48000, false, nil)
			for _, status := range []meter.FrameStatus{meter.FrameMeasured, meter.FrameIneligible, meter.FrameHeld, meter.FrameNoFine} {
				m := r.measurement(e, meter.Frame{Note: tc.note, Status: status, Taken: true})
				if m.Frame != FrameNoNote || m.Note != 0 || m.NoteName != "" || m.ScalePitch != 0 || m.Taken || len(m.Reeds) != 0 || len(m.Spectrum) != 0 {
					t.Fatalf("status %d published an unsupported key: %+v", status, m)
				}
			}
			m := r.measurement(e, meter.Frame{Note: 57, Status: meter.FrameMeasured})
			want := tuning.Note(69 - tc.offset)
			if m.Frame != FrameMeasured || m.Note != want || math.Abs(m.ScalePitch-want.Freq(440)) > 1e-9 {
				t.Fatalf("valid note after rejection: %+v", m)
			}
		})
	}
}

func TestMeasurementKeepsKeysAtTheSupportedRangeBoundaries(t *testing.T) {
	for _, tc := range []struct {
		offset int
		note   int
		want   tuning.Note
	}{
		{12, 16, tuning.MinNote},
		{-12, 96, tuning.MaxNote},
	} {
		e := NewEngine(EngineConfig{A4: 442, Transpose: 2, Octaves: []OctaveRequest{{tc.offset, 1}, {tc.offset + 12, 2}}}, nil)
		r := newRunner(e.cfg, 48000, false, nil)
		m := r.measurement(e, meter.Frame{Note: tc.note, Status: meter.FrameMeasured, Taken: true})
		if m.Frame != FrameMeasured || m.Note != tc.want || !m.Taken || math.Abs(m.ScalePitch-(tc.want+2).Freq(442)) > 1e-9 {
			t.Fatalf("boundary key %d: %+v", tc.want, m)
		}
	}
}

func TestEngineHighTrebleRegisterDoesNotPublishUnsupportedKeys(t *testing.T) {
	cfg := EngineConfig{A4: 442, Octaves: []OctaveRequest{{-12, 1}, {0, 1}, {12, 1}}}
	table := cfg.table()
	ms := runControlled(t, cfg, 6, func(e *Engine) { e.SetRecording(true); e.TakeNow() }, table[103])
	rejected := 0
	for _, m := range ms {
		if m.Frame == FrameNoNote {
			rejected++
			if m.Lock.Visible || m.Lock.Acquired || m.Lock.Recorded {
				t.Fatalf("rejected note kept a lock or recorded mark: %+v", m.Lock)
			}
			if !m.NeedlesDimmed {
				t.Fatal("rejected note left the needles undimmed")
			}
			for _, p := range m.Panels {
				if !p.Dimmed {
					t.Fatalf("rejected note left a panel undimmed: %+v", p)
				}
			}
		}
		if m.Note != 0 && !m.Note.Valid() || m.Taken {
			t.Fatalf("high register published an unsupported reading: %+v", m)
		}
	}
	if rejected == 0 {
		t.Fatal("high register never reached note rejection")
	}
}

func TestRegisterTargetsStayWithinTheSupportedKeyRange(t *testing.T) {
	var curve meter.NoteCurve
	var coverage meter.NoteCoverage
	for i := range curve {
		curve[i], coverage[i] = float64(i), true
	}
	for _, offset := range []int{-24, -12, 12, 24} {
		cfg := EngineConfig{Octaves: []OctaveRequest{{offset, 1}, {offset + 12, 2}},
			Targets: []meter.NoteCurve{curve, curve, curve}, TargetCoverage: coverage}
		got := cfg.targets()
		for i := range meter.ActiveNotes {
			from := i - offset
			valid := from >= 0 && from < meter.ActiveNotes
			want := meter.Missing
			if valid {
				want = float64(from)
			}
			if got.Coverage[i] != valid {
				t.Fatalf("offset %d row %d: coverage %v, want %v", offset, i, got.Coverage[i], valid)
			}
			for slot, target := range got.Slots {
				if target[i] != want {
					t.Fatalf("offset %d row %d slot %d: target %v, want %v", offset, i, slot, target[i], want)
				}
			}
		}
	}
}
