package report

import (
	"bytes"
	"strings"
	"testing"

	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// transposed is a one-reed pass of an instrument whose C sounds as D: its A4 key sounded at B4.
func transposed(t *testing.T) *session.Session {
	t.Helper()
	s := session.New("Bench", session.Instrument{ReedCount: 1}, passA4)
	s.Transpose = 2
	sounding := tuning.Note(71).Freq(passA4)
	s.UpsertTake(session.Take{Note: 69, At: at, Reeds: []dsp.ReedMeasure{
		{Present: true, Dev: 3, Freq: meter.FrequencyHz(sounding, 3)},
	}})
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	return s
}

// The sheet states the transposition the pass was taken with, in the sheet's own words and names.
func TestTheSheetStatesTheTransposition(t *testing.T) {
	defer i18n.SetLocale("en")
	i18n.SetLocale("en")

	r := sheet(t, transposed(t))
	if r.Session.Transpose != 2 || r.Session.Key != "C" || r.Session.Sounds != "D" {
		t.Fatalf("session info = %+v, want C sounding as D", r.Session)
	}
	if out := text(t, r); !strings.Contains(out, "C sounds as D") {
		t.Error("the English sheet does not state the transposition")
	}
	if out := text(t, sheet(t, musette(t))); strings.Contains(out, "Transposition") {
		t.Error("an untransposed sheet prints a transposition")
	}

	i18n.SetLocale("pl")
	if out := text(t, sheet(t, transposed(t))); !strings.Contains(out, "C brzmi jako D") {
		t.Error("the Polish sheet does not state the transposition")
	}
}

// A transposed take is scored at the pitch it sounded, and the CSV says what it was measured with.
func TestATransposedPassIsScoredAtItsSoundingPitch(t *testing.T) {
	r := sheet(t, transposed(t))
	cell := row(t, r, 69).Reeds[0]
	if !cell.Present || !near(cell.Curr, 3, 1e-9) {
		t.Fatalf("cell = %+v, want the 3 units the engine read", cell)
	}

	var buf bytes.Buffer
	if err := CSV(&buf, r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "transpose_semitones,2\n") {
		t.Error("the CSV does not carry the transposition")
	}
}

// A note past the curve has no goal: the CSV leaves the goal cell empty rather than writing a zero.
func TestTheCSVLeavesAMissingGoalEmpty(t *testing.T) {
	s := musette(t)
	s.UpsertTake(reeds(90, -1, 0, 1))
	r := sheet(t, s)
	if c := row(t, r, 90).Reeds[0]; c.HasGoal || c.Goal != 0 {
		t.Fatalf("cell past the curve = %+v, want no goal", c)
	}
	if c := row(t, r, 62).Reeds[0]; !c.HasGoal {
		t.Fatalf("cell inside the curve = %+v, want its goal", c)
	}

	var buf bytes.Buffer
	if err := CSV(&buf, r); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, line := range strings.Split(buf.String(), "\n") {
		if !strings.HasPrefix(line, "90,") {
			continue
		}
		found = true
		if !strings.Contains(line, ",-1.00,,-1.00,") {
			t.Fatalf("row 90 = %q, want reed 1 with an empty goal between curr and error", line)
		}
	}
	if !found {
		t.Fatalf("no row for note 90 in:\n%s", buf.String())
	}
}
