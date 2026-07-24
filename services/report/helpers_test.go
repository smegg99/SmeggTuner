package report

import (
	"path/filepath"
	"testing"

	"smegg.me/smeggtuner/core/datastore/datastoretest"
	"smegg.me/smeggtuner/core/dsp"
	coresession "smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/tuning"
	sessionsvc "smegg.me/smeggtuner/services/session"
)

const a4 = 440.0

// services builds a report service on a throwaway db and swaps the desktop seams (save dialog, browser).
func services(t *testing.T) (*Service, *sessionsvc.Service, string) {
	t.Helper()

	datastoretest.Init(t)
	sessions := sessionsvc.New()
	t.Cleanup(func() { _ = sessions.ServiceShutdown() })

	dir := t.TempDir()
	saveDialog = func(name, _ string) (string, error) { return filepath.Join(dir, name), nil }
	openInBrowser = func(string) (bool, error) { return true, nil }
	t.Cleanup(func() {
		saveDialog = realSaveDialog
		openInBrowser = realOpenInBrowser
	})

	return New(sessions), sessions, dir
}

// pass opens a session with one pass: a clean note, a merged one, and one recovered from its beat.
func pass(t *testing.T, sessions *sessionsvc.Service) {
	t.Helper()

	if _, err := sessions.Create(sessionsvc.NewSessionDTO{
		Name:       "Morino",
		Instrument: coresession.Instrument{ReedCount: 3, A4: a4},
	}); err != nil {
		t.Fatal(err)
	}
	for _, take := range []coresession.Take{
		reeds(60, -8.2, 0.4, 8.1),
		merged(64),
		reeds(67, -7.9, 0.2, 12.4),
	} {
		if err := sessions.UpsertTake(take); err != nil {
			t.Fatal(err)
		}
	}
}

// passWithReeds is pass for a given reed count; past four reeds the sheet turns landscape.
func passWithReeds(t *testing.T, sessions *sessionsvc.Service, count int) {
	t.Helper()

	if _, err := sessions.Create(sessionsvc.NewSessionDTO{
		Name:       "Morino",
		Instrument: coresession.Instrument{ReedCount: count, A4: a4},
	}); err != nil {
		t.Fatal(err)
	}
	for _, note := range []tuning.Note{60, 67} {
		cents := make([]float64, count)
		for i := range cents {
			cents[i] = float64(i)*4 - 8
		}
		if err := sessions.UpsertTake(reeds(note, cents...)); err != nil {
			t.Fatal(err)
		}
	}
}

func reeds(note tuning.Note, cents ...float64) coresession.Take {
	ref := note.Freq(a4)
	take := coresession.Take{Note: note}
	for _, c := range cents {
		take.Reeds = append(take.Reeds, dsp.ReedMeasure{Freq: tuning.FreqAtCents(ref, c), DevCents: c})
	}
	return take
}

// merged is a take whose two reeds could not be separated; its frequencies are lobes of one peak.
func merged(note tuning.Note) coresession.Take {
	take := reeds(note, -33.7, 39.1, 0)
	take.ReedsMerged = true
	take.Beats = []dsp.BeatMeasure{{Pair: "1-2", Hz: 1.8, FromEnvelope: true, Depth: 0.8}}
	return take
}

func export(t *testing.T, s *Service, opts OptionsDTO) *ResultDTO {
	t.Helper()
	out, err := s.Export(opts)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	return out
}
