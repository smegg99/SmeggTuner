package record

import (
	"testing"

	"smegg.me/smeggtuner/core/datastore/datastoretest"
	"smegg.me/smeggtuner/core/dsp"
	coresession "smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/tuning"
	sessionsvc "smegg.me/smeggtuner/services/session"
)

const a4 = 440.0

// services builds a record service over a session service on a throwaway database; nothing is opened.
func services(t *testing.T) (*Service, *sessionsvc.Service) {
	t.Helper()
	datastoretest.Init(t)
	sessions := sessionsvc.New()
	t.Cleanup(func() { _ = sessions.ServiceShutdown() })
	return New(sessions), sessions
}

func open(t *testing.T, sessions *sessionsvc.Service, reeds int) {
	t.Helper()
	_, err := sessions.Create(sessionsvc.NewSessionDTO{
		Name:       "Morino",
		Instrument: coresession.Instrument{ReedCount: reeds, A4: a4},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func openInstrument(t *testing.T, sessions *sessionsvc.Service, i coresession.Instrument) {
	t.Helper()
	i.A4 = a4
	if _, err := sessions.Create(sessionsvc.NewSessionDTO{Name: "Bench", Instrument: i}); err != nil {
		t.Fatal(err)
	}
}

// morino is a three-voice accordion whose ranks and switches are described.
func morino() coresession.Instrument {
	return coresession.Instrument{
		Make:      "Hohner",
		Model:     "Morino",
		ReedCount: 3,
		Banks:     banks(coresession.BankM1, coresession.BankM2, coresession.BankM3),
		Registers: []coresession.Register{
			{Name: "MMM", Banks: banks(coresession.BankM1, coresession.BankM2, coresession.BankM3)},
			{Name: "M1M3", Banks: banks(coresession.BankM1, coresession.BankM3)},
			{Name: "M2", Banks: banks(coresession.BankM2)},
		},
	}
}

func banks(b ...coresession.Bank) []coresession.Bank { return b }

// reeds is the engine's output for a note at the given per-reed deviations, low reed first.
func reeds(note tuning.Note, cents ...float64) []dsp.ReedMeasure {
	ref := note.Freq(a4)
	out := make([]dsp.ReedMeasure, 0, len(cents))
	for _, c := range cents {
		out = append(out, dsp.ReedMeasure{Freq: tuning.FreqAtCents(ref, c), DevCents: c})
	}
	return out
}

// fine is a fine result: a note, its reeds, and whether the engine has locked onto it.
func fine(note tuning.Note, locked bool, cents ...float64) dsp.Measurement {
	ref := note.Freq(a4)
	m := dsp.Measurement{
		Note:           note,
		NoteName:       note.Name(tuning.NamingCDEFGAB),
		Locked:         locked,
		ScalePitch:     ref,
		ReedsSeparated: true,
		State:          dsp.StateRunning,
	}
	for _, c := range cents {
		m.Reeds = append(m.Reeds, dsp.ReedMeasure{Freq: tuning.FreqAtCents(ref, c), DevCents: c})
	}
	return m
}

// merged is a fine result the spectrum could not pull apart; the only reading is the beat off the amplitude.
func merged(note tuning.Note, locked bool) dsp.Measurement {
	m := fine(note, locked, -4, 4)
	m.ReedsSeparated = false
	m.Beats = []dsp.BeatMeasure{{Pair: "1-2", Hz: 1.2, FromEnvelope: true, Depth: 0.8}}
	return m
}

// recovered is the same pair, with the reeds reconstructed from that beat.
func recovered(note tuning.Note, locked bool) dsp.Measurement {
	m := merged(note, locked)
	m.ReedsFromBeat = true
	return m
}

// heartbeat is the ~12/s tick carrying state and level, with no note, so it says nothing about the lock.
func heartbeat() dsp.Measurement {
	return dsp.Measurement{State: dsp.StateRunning, InputLevel: 0.2}
}

// lock plays a note: fine results with the lock true, interleaved with heartbeats, then release.
func lock(s *Service, note tuning.Note, cents ...float64) {
	s.OnMeasurement(fine(note, false, cents...))
	s.OnMeasurement(heartbeat())
	s.OnMeasurement(fine(note, true, cents...))
	s.OnMeasurement(heartbeat())
	s.OnMeasurement(fine(note, true, cents...))
	s.OnMeasurement(heartbeat())
	// The note ends: the lock streak breaks.
	s.OnMeasurement(fine(note, false))
}

func table(t *testing.T, r *Service) *TableDTO {
	t.Helper()
	tab, err := r.Table()
	if err != nil {
		t.Fatal(err)
	}
	return tab
}

// capture swaps the emit seam and records what goes through it.
func capture(t *testing.T) *[]struct {
	name string
	data any
} {
	t.Helper()
	var seen []struct {
		name string
		data any
	}
	prev := emitEvent
	emitEvent = func(name string, data any) {
		seen = append(seen, struct {
			name string
			data any
		}{name, data})
	}
	t.Cleanup(func() { emitEvent = prev })
	return &seen
}

func near(a, b float64) bool { return a-b < 0.01 && b-a < 0.01 }
