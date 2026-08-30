package report

import (
	"fmt"
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// Session A4 and pass A4 differ throughout these tests on purpose.
const passA4 = 435.0

// gone marks a slot that did not sound, as the engine reports it.
const gone = meter.Missing

var at = time.Date(2026, 7, 12, 10, 30, 0, 0, time.UTC)

// reeds is a take as the engine records one: a slot per deviation at the key's own octave, gone
// for a slot that did not sound, and the adjacent beats from the slots' frequencies.
func reeds(note tuning.Note, devs ...float64) session.Take {
	t := session.Take{Note: note, At: at}
	for _, d := range devs {
		t.Reeds = append(t.Reeds, slot(note, 0, d))
	}
	t.Beats = beatsOf(t.Reeds)
	return t
}

func slot(note tuning.Note, octave int, d float64) dsp.ReedMeasure {
	if d == gone {
		return dsp.ReedMeasure{Dev: gone}
	}
	ref := (note + tuning.Note(octave)).Freq(passA4)
	return dsp.ReedMeasure{Present: true, Dev: d, Octave: octave, Freq: meter.FrequencyHz(ref, d)}
}

func beatsOf(slots []dsp.ReedMeasure) []dsp.BeatMeasure {
	var out []dsp.BeatMeasure
	for i := 1; i < len(slots); i++ {
		b := dsp.BeatMeasure{Pair: fmt.Sprintf("%d-%d", i, i+1)}
		if slots[i-1].Present && slots[i].Present {
			b.Present, b.Hz = true, meter.HarmonicBeat(slots[i-1].Freq, slots[i].Freq, 0, true)
		}
		out = append(out, b)
	}
	return out
}

// musette is the fixture: three reeds, a -8/0/+8 curve, and a pass with every kind of row (clean,
// duplicated, a silent middle slot, hand-edited with a silent third slot).
func musette(t *testing.T) *session.Session {
	t.Helper()

	s := session.New("Hohner Morino - Jan K.",
		session.Instrument{Serial: "SN-4471", ReedCount: 3},
		passA4)
	s.Notes = "Musette register, low block replaced."
	s.Curve = curve(t, 3, map[tuning.Note][]float64{60: {-8, 0, 8}, 84: {-8, 0, 8}})

	s.UpsertTake(reeds(60, -8.4, 0.3, 12.5)) // slot 3 out of tolerance
	s.UpsertTake(reeds(62, -7.9, 0.1, 8.2))
	s.UpsertTake(reeds(62, -7.6, 0.4, 8.1)) // the same note again: the last one shows
	s.UpsertTake(reeds(64, -8.2, gone, 8.3))
	s.UpsertTake(reeds(65, -7.5, 0.2, 8.3))

	manual := reeds(67, -8.1, 0.2, gone) // the third reed never sounded
	manual.Manual = true
	s.UpsertTake(manual)

	if err := s.Validate(); err != nil {
		t.Fatalf("fixture session: %v", err)
	}
	return s
}

func curve(t *testing.T, reedCount int, anchors map[tuning.Note][]float64) *target.Curve {
	t.Helper()
	c := target.NewCurve("Musette", reedCount)
	for note, values := range anchors {
		for reed, v := range values {
			if err := c.Set(note, reed, v); err != nil {
				t.Fatalf("curve.Set(%d, %d): %v", note, reed, err)
			}
		}
	}
	return c
}

func sheet(t *testing.T, s *session.Session) *Report {
	t.Helper()
	r, err := Build(s, Options{Now: at})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return r
}

func row(t *testing.T, r *Report, note tuning.Note) Row {
	t.Helper()
	for _, row := range r.Rows {
		if row.Note == note {
			return row
		}
	}
	t.Fatalf("no row for note %d", note)
	return Row{}
}

func near(a, b, tol float64) bool { return a-b < tol && b-a < tol }

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
