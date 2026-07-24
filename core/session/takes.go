package session

import (
	"sort"
	"time"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
)

// TakeFrom records a measurement as a take, at the given time.
func TakeFrom(m dsp.Measurement, at time.Time) Take {
	return Take{
		Note:          m.Note,
		At:            at,
		Reeds:         m.Reeds,
		Beats:         m.Beats,
		ReedsMerged:   !m.ReedsSeparated,
		ReedsFromBeat: m.ReedsFromBeat,
	}
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

// UndoLast drops the reading captured last. An empty session is not an error.
func (s *Session) UndoLast() (Take, bool) {
	if len(s.Takes) == 0 {
		return Take{}, false
	}
	last := s.Takes[len(s.Takes)-1]
	s.Takes = s.Takes[:len(s.Takes)-1]
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

// Readings hands the session's rows to target.Fit. Built next to Display so the fit and the tuning table look at the same rows.
func (s *Session) Readings() []target.Reading {
	rows := s.Display()
	out := make([]target.Reading, 0, len(rows))
	for _, r := range rows {
		out = append(out, target.Reading{
			Note:          r.Note,
			Reeds:         r.Take.Reeds,
			ReedsMerged:   r.Take.ReedsMerged,
			ReedsFromBeat: r.Take.ReedsFromBeat,
		})
	}
	return out
}
