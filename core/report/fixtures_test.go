package report

import (
	"testing"
	"time"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
)

// Session A4 and pass A4 differ throughout these tests on purpose.
const passA4 = 435.0

// Merged-note lobes, deliberately unlike anything else so a test can detect them.
const (
	mergedLobeLow  = -37.3
	mergedLobeHigh = 41.7
)

var at = time.Date(2026, 7, 12, 10, 30, 0, 0, time.UTC)

func reeds(note tuning.Note, cents ...float64) session.Take {
	ref := note.Freq(passA4)
	t := session.Take{Note: note, At: at}
	for _, c := range cents {
		t.Reeds = append(t.Reeds, dsp.ReedMeasure{Freq: tuning.FreqAtCents(ref, c), DevCents: c})
	}
	return t
}

// mergedTake: two reeds the spectrum could not separate, so its Reeds are lobes and only the envelope beat is real.
func mergedTake(note tuning.Note) session.Take {
	t := reeds(note, mergedLobeLow, mergedLobeHigh, 0)
	t.ReedsMerged = true
	t.Beats = []dsp.BeatMeasure{{Pair: "1-2", Hz: 2.0, FromEnvelope: true, Depth: 0.8}}
	return t
}

// derivedTake: a merge with reeds recovered from the beat, which are measured and print.
func derivedTake(note tuning.Note, cents ...float64) session.Take {
	t := reeds(note, cents...)
	t.ReedsMerged = true
	t.ReedsFromBeat = true
	t.Beats = []dsp.BeatMeasure{{Pair: "1-2", Hz: 1.4, FromEnvelope: true, Depth: 0.7}}
	return t
}

// musette is the fixture: three reeds, a -8/0/+8 curve, and a pass with every kind of row (clean, merged, recovered, hand-edited, duplicated, one missing reed).
func musette(t *testing.T) *session.Session {
	t.Helper()

	s := session.New("Hohner Morino - Jan K.",
		session.Instrument{Make: "Hohner", Model: "Morino VI M", Serial: "SN-4471", ReedCount: 3},
		passA4)
	s.Notes = "Musette register, low block replaced."
	s.Curve = curve(t, 3, map[tuning.Note][]float64{60: {-8, 0, 8}, 84: {-8, 0, 8}})

	s.UpsertTake(reeds(60, -8.4, 0.3, 12.5)) // reed 3 out of tolerance
	s.UpsertTake(reeds(62, -7.9, 0.1, 8.2))
	s.UpsertTake(reeds(62, -7.6, 0.4, 8.1)) // the same note again: the last one shows
	s.UpsertTake(mergedTake(64))
	s.UpsertTake(derivedTake(65, -7.5, 0.2, 8.3))

	manual := reeds(67, -8.1, 0.2) // the third reed never sounded
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
			if err := c.Set(note, reed, v, passA4); err != nil {
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
