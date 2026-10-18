package session

import (
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	coresession "smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// CurveViewDTO is what the curve editor draws: the goal line as the engine reads it and the pass's
// treble readings against it, in reference units on the curve's columns.
type CurveViewDTO struct {
	// Goals is one row per note the curve covers, low to high; empty without a curve.
	Goals []GoalRowDTO `json:"goals"`
	// Recorded is every treble reading of the pass, in note order.
	Recorded []RecordedDTO `json:"recorded"`
}

// GoalRowDTO is the curve at one note: per column the goal, or null where it sets none.
type GoalRowDTO struct {
	Note  tuning.Note `json:"note"`
	Reeds []*float64  `json:"reeds"`
}

// RecordedDTO is one reading laid on the curve's columns.
type RecordedDTO struct {
	Note     tuning.Note       `json:"note"`
	Register string            `json:"register,omitempty"`
	Manual   bool              `json:"manual,omitempty"`
	Take     int               `json:"take"` // index into the session's readings
	Reeds    []RecordedReedDTO `json:"reeds"`
}

// RecordedReedDTO is one sounding slot of a reading: its curve column, what it read, and the goal there.
type RecordedReedDTO struct {
	Column int      `json:"column"`
	Dev    float64  `json:"dev"`
	Goal   *float64 `json:"goal"` // null where the curve sets no goal
}

// CurveView is the open session's goal line and its readings, for the editor's recorded overlay.
func (s *Service) CurveView() (*CurveViewDTO, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.active == nil {
		return nil, ErrNoSession
	}
	a := s.active
	out := &CurveViewDTO{Goals: []GoalRowDTO{}, Recorded: []RecordedDTO{}}
	for n := tuning.MinNote; n <= tuning.MaxNote; n++ {
		if !a.Curve.Covers(n) {
			continue
		}
		goal := a.Curve.At(n)
		row := GoalRowDTO{Note: n, Reeds: make([]*float64, len(goal))}
		for col := range goal {
			row.Reeds[col] = goalOf(goal, col)
		}
		out.Goals = append(out.Goals, row)
	}
	for _, r := range a.Display() {
		if r.Take.Bass {
			continue
		}
		cols := a.Instrument.Columns(r.Take)
		goal := a.Curve.At(r.Note)
		rec := RecordedDTO{Note: r.Note, Register: r.Register, Manual: r.Take.Manual, Take: r.Index, Reeds: []RecordedReedDTO{}}
		for slot, m := range r.Take.Reeds {
			if !m.Present {
				continue
			}
			col := coresession.ColumnOf(cols, slot)
			rec.Reeds = append(rec.Reeds, RecordedReedDTO{Column: col, Dev: m.Dev, Goal: goalOf(goal, col)})
		}
		out.Recorded = append(out.Recorded, rec)
	}
	return out, nil
}

// goalOf is column col of a Curve.At result, nil where the curve sets no goal.
func goalOf(goal []float64, col int) *float64 {
	if col < 0 || col >= len(goal) || goal[col] <= meter.Missing {
		return nil
	}
	v := goal[col]
	return &v
}
