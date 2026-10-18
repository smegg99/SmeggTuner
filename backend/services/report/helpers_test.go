package report

import (
	"path/filepath"
	"testing"

	"smegg.me/smeggtuner/backend/core/datastore/datastoretest"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	coresession "smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/tuning"
	sessionsvc "smegg.me/smeggtuner/backend/services/session"
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

// pass opens a session with one pass: two clean notes and one where a reed did not sound.
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
		silentMiddle(64),
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
		devs := make([]float64, count)
		for i := range devs {
			devs[i] = float64(i)*4 - 8
		}
		if err := sessions.UpsertTake(reeds(note, devs...)); err != nil {
			t.Fatal(err)
		}
	}
}

// reeds is a take of note with every slot sounding at the given reference-unit deviations.
func reeds(note tuning.Note, devs ...float64) coresession.Take {
	ref := note.Freq(a4)
	take := coresession.Take{Note: note}
	for _, d := range devs {
		take.Reeds = append(take.Reeds, dsp.ReedMeasure{Present: true, Freq: meter.FrequencyHz(ref, d), Dev: d})
	}
	return take
}

// silentMiddle is a three-slot take whose middle reed did not sound; the engine marks it with the Missing sentinel.
func silentMiddle(note tuning.Note) coresession.Take {
	take := reeds(note, -7.6, 0, 8.3)
	take.Reeds[1] = dsp.ReedMeasure{Dev: meter.Missing}
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
