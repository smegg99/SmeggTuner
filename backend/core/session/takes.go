package session

import (
	"sort"
	"time"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// TakeFrom records a measured frame as a take, at the given time.
func TakeFrom(m dsp.Measurement, at time.Time) Take {
	return Take{Note: m.Note, At: at, Reeds: m.Reeds, Beats: m.Beats}
}

// UpsertTake records a reading, replacing that voice's previous take or appending a new one. The reed and beat slices are copied because a take outlives the live measurement.
func (s *Session) UpsertTake(t Take) {
	t.Reeds = append([]dsp.ReedMeasure(nil), t.Reeds...)
	t.Beats = append([]dsp.BeatMeasure(nil), t.Beats...)

	v := voiceOf(t)
	for i := range s.Takes {
		if voiceOf(s.Takes[i]) == v {
			s.Takes[i] = t
			return
		}
	}
	s.Takes = append(s.Takes, t)
}

// SetReed hand-writes a sounding slot's deviation in reference units at a4 and transpose, re-reads
// the beats the way the engine does, and marks the take manual. False for a slot out of range or one
// that did not sound.
func (t *Take) SetReed(slot int, dev, a4 float64, transpose int) bool {
	if slot < 0 || slot >= len(t.Reeds) || !t.Reeds[slot].Present {
		return false
	}
	nominal := (t.Note + tuning.Note(transpose+t.Reeds[slot].Octave)).Freq(a4)
	t.Reeds[slot].Dev = dev
	t.Reeds[slot].Freq = meter.FrequencyHz(nominal, dev)
	t.Beats, t.Manual = dsp.AdjacentBeats(t.Reeds), true
	return true
}

// Reading is the take as core/target scores it, measured with the session's transpose.
func (t Take) Reading(transpose int) target.Reading {
	return target.Reading{Note: t.Note, Transpose: transpose, Reeds: t.Reeds, Beats: t.Beats}
}

// Latest is the index of the take captured last, or -1 on an empty pass. A replayed voice keeps
// its place in Takes, so this is the newest At, the last index on a tie.
func (s *Session) Latest() int {
	latest := -1
	for i, t := range s.Takes {
		if latest < 0 || !t.At.Before(s.Takes[latest].At) {
			latest = i
		}
	}
	return latest
}

// UndoLast drops the take captured last (see Latest). An empty session is not an error.
func (s *Session) UndoLast() (Take, bool) {
	i := s.Latest()
	if i < 0 {
		return Take{}, false
	}
	last := s.Takes[i]
	s.Takes = append(s.Takes[:i], s.Takes[i+1:]...)
	return last, true
}

// DeleteTake removes one reading by index into Takes.
func (s *Session) DeleteTake(i int) bool {
	if i < 0 || i >= len(s.Takes) {
		return false
	}
	s.Takes = append(s.Takes[:i], s.Takes[i+1:]...)
	return true
}

func (s *Session) Clear() { s.Takes = nil }

// DisplayRow is one voice as the tuning table shows it. Index is where the reading sits in Takes.
type DisplayRow struct {
	Note     tuning.Note `json:"note"`
	Register string      `json:"register,omitempty"`
	Take     Take        `json:"take"`
	Index    int         `json:"index"`
}

// voice keys a row by note and register: the same note under different registers reads different reeds.
type voice struct {
	note     tuning.Note
	register string
	bass     bool
}

func voiceOf(t Take) voice {
	return voice{note: t.Note, register: t.Register, bass: t.Bass}
}

// Display is the session's readings in reading order. Takes stays in capture order; only the table is sorted.
func (s *Session) Display() []DisplayRow {
	rows := make([]DisplayRow, 0, len(s.Takes))
	for i, t := range s.Takes {
		rows = append(rows, DisplayRow{
			Note:     t.Note,
			Register: t.Register,
			Take:     t,
			Index:    i,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.Note != b.Note {
			return a.Note < b.Note
		}
		return a.Register < b.Register
	})
	return rows
}

// Readings hands the treble rows to target.Fit, each slot moved to its card column of the curve's
// width columns (Instrument.Columns) and beats left out. Built next to Display so the fit and the
// tuning table look at the same rows.
func (s *Session) Readings(width int) []target.Reading {
	rows := s.Display()
	out := make([]target.Reading, 0, len(rows))
	for _, r := range rows {
		if r.Take.Bass {
			continue
		}
		reeds := make([]dsp.ReedMeasure, width)
		for c := range reeds {
			reeds[c].Dev = meter.Missing
		}
		cols := s.Instrument.Columns(r.Take)
		for slot, m := range r.Take.Reeds {
			if c := ColumnOf(cols, slot); c < width {
				reeds[c] = m
			}
		}
		out = append(out, target.Reading{Note: r.Note, Transpose: s.Transpose, Reeds: reeds})
	}
	return out
}
