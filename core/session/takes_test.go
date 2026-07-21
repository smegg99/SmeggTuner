package session

import (
	"testing"
	"time"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

// Capture order is the order the technician worked in, not the display order, and must not be sorted.
func TestTakesStayInCaptureOrder(t *testing.T) {
	s := newTestSession(t, 3)
	now := time.Now()
	for i, n := range []tuning.Note{69, 60, 72, 60} {
		s.UpsertTake(take(n, 3, now.Add(time.Duration(i)*time.Second)))
	}
	// 60 was played twice: the second REPLACES the first, in place.
	want := []tuning.Note{69, 60, 72}
	if len(s.Takes) != len(want) {
		t.Fatalf("readings = %d, want %d", len(s.Takes), len(want))
	}
	for i, n := range want {
		if s.Takes[i].Note != n {
			t.Fatalf("reading %d = note %d, want %d", i, s.Takes[i].Note, n)
		}
	}
}

// Playing a note again replaces its reading in place: one job, one set of readings.
func TestReplayingANoteReplacesIt(t *testing.T) {
	s := newTestSession(t, 3)
	now := time.Now()
	first := take(60, 3, now)
	first.Reeds[0].DevCents = -4
	second := take(60, 3, now.Add(time.Second))
	second.Reeds[0].DevCents = 7
	s.UpsertTake(first)
	s.UpsertTake(take(62, 3, now.Add(2*time.Second)))
	s.UpsertTake(second)

	if len(s.Takes) != 2 {
		t.Fatalf("readings = %d, want 2", len(s.Takes))
	}
	// Replaced in place: capture order is the order he worked in.
	if s.Takes[0].Note != 60 || s.Takes[1].Note != 62 {
		t.Fatalf("readings out of capture order: %d, %d", s.Takes[0].Note, s.Takes[1].Note)
	}

	rows := s.Display()
	if len(rows) != 2 {
		t.Fatalf("display rows = %d, want 2", len(rows))
	}
	if rows[0].Note != 60 || rows[1].Note != 62 {
		t.Fatalf("rows out of note order: %d, %d", rows[0].Note, rows[1].Note)
	}
	if rows[0].Take.Reeds[0].DevCents != 7 {
		t.Fatalf("row 60 kept the older reading: %v", rows[0].Take.Reeds[0].DevCents)
	}
}

func TestUndoLastOnEmptyPassIsNoOp(t *testing.T) {
	s := newTestSession(t, 3)
	if _, ok := s.UndoLast(); ok {
		t.Fatal("undo on an empty pass reported a take")
	}
	if len(s.Takes) != 0 {
		t.Fatalf("takes = %d after undo on empty", len(s.Takes))
	}
}

func TestUndoLastAndClear(t *testing.T) {
	s := newTestSession(t, 3)
	now := time.Now()
	s.UpsertTake(take(60, 3, now))
	s.UpsertTake(take(62, 3, now.Add(time.Second)))

	got, ok := s.UndoLast()
	if !ok || got.Note != 62 {
		t.Fatalf("undo = (%d, %v), want (62, true)", got.Note, ok)
	}
	if len(s.Takes) != 1 || s.Takes[0].Note != 60 {
		t.Fatalf("takes after undo: %+v", s.Takes)
	}

	s.Clear()
	if len(s.Takes) != 0 {
		t.Fatalf("takes after clear = %d", len(s.Takes))
	}
	if _, ok := s.UndoLast(); ok {
		t.Fatal("undo after clear reported a take")
	}
}

// Nothing here may assume the musette three.
func TestReedCountsOneAndFive(t *testing.T) {
	for _, n := range []int{1, 5} {
		s := newTestSession(t, n)
		s.UpsertTake(take(69, n, time.Now()))
		rows := s.Display()
		if len(rows) != 1 {
			t.Fatalf("reeds=%d: rows = %d", n, len(rows))
		}
		if got := len(rows[0].Take.Reeds); got != n {
			t.Fatalf("reeds=%d: take carries %d reeds", n, got)
		}
	}
}

// The session's A4 is the reference every reading was taken against; core keeps it.
func TestReadingsKeepTheSessionReference(t *testing.T) {
	s := newTestSession(t, 3)
	if s.A4 != 442 {
		t.Fatalf("session a4 = %v, want 442", s.A4)
	}
	s.UpsertTake(take(69, 3, time.Now()))
	if len(s.Takes) != 1 {
		t.Fatalf("readings = %d, want 1", len(s.Takes))
	}
}

// A take outlives the live measurement it came from; the engine's DTO is not ours to keep pointers into.
func TestAppendTakeCopiesMeasurement(t *testing.T) {
	s := newTestSession(t, 2)
	m := dsp.Measurement{
		Note:  69,
		Reeds: reeds(2),
		Beats: []dsp.BeatMeasure{{Pair: "1-2", Hz: 3.1}},
	}
	s.UpsertTake(TakeFrom(m, time.Now()))

	m.Reeds[0].DevCents = 999
	m.Beats[0].Hz = 999

	if got := s.Takes[0].Reeds[0].DevCents; got != 0 {
		t.Fatalf("take aliased the measurement's reeds: %v", got)
	}
	if got := s.Takes[0].Beats[0].Hz; got != 3.1 {
		t.Fatalf("take aliased the measurement's beats: %v", got)
	}
}

// Only the engine knows whether the peaks were reeds or one merged pair, so the take has to carry it.
func TestTakeFromCarriesMergedReeds(t *testing.T) {
	now := time.Now()
	m := dsp.Measurement{Note: 69, Reeds: reeds(3), ReedsSeparated: true}

	if TakeFrom(m, now).ReedsMerged {
		t.Fatal("separated reeds came back as a merged pair")
	}
	m.ReedsSeparated = false
	if !TakeFrom(m, now).ReedsMerged {
		t.Fatal("a merged pair came back as separated reeds, which is the reading " +
			"a technician would act on")
	}
}

// Readings is what the fitter is handed. One note, one row: the last take of it, which the table shows.
func TestPassReadings(t *testing.T) {
	s := newTestSession(t, 3)
	now := time.Now()

	s.UpsertTake(take(69, 3, now))
	second := take(69, 3, now.Add(time.Second))
	second.Reeds[0].DevCents = 7.5
	s.UpsertTake(second)
	merged := take(60, 3, now.Add(2*time.Second))
	merged.ReedsMerged = true
	s.UpsertTake(merged)

	got := s.Readings()
	if len(got) != 2 {
		t.Fatalf("three takes of two notes gave %d readings", len(got))
	}
	if got[0].Note != 60 || !got[0].ReedsMerged {
		t.Fatalf("the merged take lost its flag on the way to the fit: %+v", got[0])
	}
	if got[1].Note != 69 || got[1].Reeds[0].DevCents != 7.5 {
		t.Fatalf("the second take of a note is the one that counts: %+v", got[1])
	}
}
