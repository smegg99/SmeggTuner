package record

import (
	"fmt"
	"testing"

	"smegg.me/smeggtuner/backend/core/datastore/datastoretest"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	coresession "smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/tuning"
	sessionsvc "smegg.me/smeggtuner/backend/services/session"
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

// reeds is the engine's slots for a note at the given deviations in reference units, low slot first.
func reeds(note tuning.Note, devs ...float64) []dsp.ReedMeasure {
	ref := note.Freq(a4)
	out := make([]dsp.ReedMeasure, 0, len(devs))
	for _, d := range devs {
		out = append(out, dsp.ReedMeasure{Present: true, Freq: meter.FrequencyHz(ref, d), Dev: d})
	}
	return out
}

// silent is a slot that did not sound, as the engine reports one.
func silent() dsp.ReedMeasure { return dsp.ReedMeasure{Dev: meter.Missing} }

// beats are the adjacent pairs the engine derives from its slots.
func beats(note tuning.Note, slots []dsp.ReedMeasure) []dsp.BeatMeasure {
	var out []dsp.BeatMeasure
	for i := 1; i < len(slots); i++ {
		b := dsp.BeatMeasure{Pair: fmt.Sprintf("%d-%d", i, i+1)}
		if slots[i-1].Present && slots[i].Present {
			b.Present, b.Hz = true, meter.HarmonicBeat(slots[i-1].Freq, slots[i].Freq, note.Freq(a4), true)
		}
		out = append(out, b)
	}
	return out
}

// measured is an analysis with fresh slot values; took says the engine took it as a reading.
func measured(note tuning.Note, took bool, slots ...dsp.ReedMeasure) dsp.Measurement {
	return dsp.Measurement{
		State:      dsp.StateRunning,
		Frame:      dsp.FrameMeasured,
		Note:       note,
		NoteName:   note.Name(tuning.NamingCDEFGAB),
		ScalePitch: note.Freq(a4),
		Taken:      took,
		Lock:       dsp.Lock{Visible: took, Internal: took, Acquired: took, Recorded: took},
		Reeds:      slots,
		Beats:      beats(note, slots),
	}
}

// fine is a measured analysis with every slot sounding.
func fine(note tuning.Note, took bool, devs ...float64) dsp.Measurement {
	return measured(note, took, reeds(note, devs...)...)
}

// heartbeat is the ~20/s tick between analyses; its lock is the screen's, so it says nothing new.
func heartbeat() dsp.Measurement {
	return dsp.Measurement{State: dsp.StateRunning, InputLevel: 0.2, Reeds: []dsp.ReedMeasure{}}
}

// released is the analysis after the note stops: no note, the visible lock off.
func released() dsp.Measurement {
	return dsp.Measurement{State: dsp.StateRunning, Frame: dsp.FrameNoNote, Reeds: []dsp.ReedMeasure{}}
}

// lock plays a note: settling analyses, the one the engine takes, the held note after it (the
// engine inhibits a second take of it), interleaved with heartbeats, then release.
func lock(s *Service, note tuning.Note, devs ...float64) {
	s.OnMeasurement(fine(note, false, devs...))
	s.OnMeasurement(heartbeat())
	s.OnMeasurement(fine(note, true, devs...))
	s.OnMeasurement(heartbeat())
	s.OnMeasurement(fine(note, false, devs...))
	s.OnMeasurement(heartbeat())
	s.OnMeasurement(released())
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
