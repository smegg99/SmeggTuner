package report

import (
	"time"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
)

// Build lays the session's readings against its goal curve into a sheet; a merged take prints its beat instead of per-reed cells.
func Build(s *session.Session, opts Options) (*Report, error) {
	if s == nil {
		return nil, ErrNoSession
	}
	if len(s.Takes) == 0 {
		return nil, ErrNoReadings
	}

	tol, beatTol := opts.Tolerance, opts.BeatTolerance
	if tol <= 0 {
		tol = target.DefaultTolerance
	}
	if beatTol <= 0 {
		beatTol = target.DefaultBeatTolerance
	}
	// Instrument tolerances override the defaults; precedence lives in core/session so card, needle and table agree.
	tol, beatTol = s.Instrument.Tolerances(tol, beatTol)

	reedCount := s.Instrument.ReedCount
	if reedCount < session.MinReeds {
		reedCount = session.MinReeds
	}
	if reedCount > session.MaxReeds {
		reedCount = session.MaxReeds
	}

	// Named columns only when the bank list matches the reed count, so header and body cannot disagree.
	var banks []session.Bank
	if len(s.Instrument.Banks) == reedCount {
		banks = s.Instrument.Banks
	}

	rows, pairs := build(s, s.Curve, s.A4, s.Instrument, banks, reedCount, opts.Naming, tol, beatTol)

	rep := &Report{
		Identity: Identity{
			Session:   s.Name,
			Make:      s.Instrument.Make,
			Model:     s.Instrument.Model,
			Serial:    s.Instrument.Serial,
			ReedCount: reedCount,
			Notes:     s.Notes,
			Registers: s.Instrument.Registers,
			HasCurve:  s.Curve != nil,
		},
		Session: SessionInfo{
			At:       s.Updated,
			A4:       s.A4,
			Readings: len(s.Takes),
		},
		Letterhead: opts.Letterhead,
		Date:       opts.Date,
		Generated:  opts.Now,
		Pairs:      pairs,
		Rows:       rows,
		Banks:      banks,
	}
	rep.MultiRegister = manyRegisters(rows)
	if s.Curve != nil {
		rep.Identity.CurveName = s.Curve.Name
		rep.Identity.CurveReeds = s.Curve.ReedCount
	}
	if rep.Date.IsZero() {
		rep.Date = s.Updated
	}
	if rep.Generated.IsZero() {
		rep.Generated = time.Now()
	}
	for i := 1; i <= reedCount; i++ {
		rep.Reeds = append(rep.Reeds, i)
	}
	rep.Summary = summarize(rows, tol, beatTol)
	rep.Layout = layoutFor(rep.Columns())
	rep.Graph = graph(rep, s.Curve)
	return rep, nil
}

func build(
	s *session.Session, c *target.Curve, a4 float64, i session.Instrument, banks []session.Bank,
	reedCount int, naming tuning.ScaleNaming, tol, beatTol float64,
) ([]Row, []Pair) {
	display := s.Display()
	rows := make([]Row, 0, len(display))
	beats := make([]map[string]target.BeatError, 0, len(display))

	for _, d := range display {
		m := measurementOf(d.Take)
		row := Row{
			Note:     d.Note,
			Name:     d.Note.Name(naming),
			Register: d.Register,
			At:       d.Take.At,
			Manual:   d.Take.Manual,
			Merged:   d.Take.ReedsMerged && !d.Take.ReedsFromBeat,
			Derived:  d.Take.ReedsMerged && d.Take.ReedsFromBeat,
		}

		// A merged take gets no per-reed cells: its figures are lobes of one peak.
		if !row.Merged {
			errs := target.Errors(m, c, a4, tol)
			row.Reeds = reedCells(errs, reedCount, banks, columnsFor(i, d.Take, len(errs)))
			for _, cell := range row.Reeds {
				if cell.Present && !cell.InTol {
					row.OutOfTol++
				}
			}
		}

		byPair := make(map[string]target.BeatError)
		for _, b := range target.BeatErrors(m, c, a4, beatTol) {
			byPair[b.Pair] = b
		}
		beats = append(beats, byPair)
		rows = append(rows, row)
	}

	pairs := pairsOf(reedCount, beats)
	for i := range rows {
		rows[i].Beats = beatCells(beats[i], pairs)
		for _, cell := range rows[i].Beats {
			if cell.Present && !cell.InTol {
				rows[i].OutOfTol++
			}
		}
	}
	return rows, pairs
}

// measurementOf reconstructs a take's measurement carrying no scale pitch, so core/target measures against the pass's frozen A4.
func measurementOf(t session.Take) dsp.Measurement {
	return dsp.Measurement{
		Note:           t.Note,
		Reeds:          t.Reeds,
		Beats:          t.Beats,
		ReedsSeparated: !t.ReedsMerged,
		ReedsFromBeat:  t.ReedsFromBeat,
	}
}

func manyRegisters(rows []Row) bool {
	first := ""
	for i, r := range rows {
		if i == 0 {
			first = r.Register
			continue
		}
		if r.Register != first {
			return true
		}
	}
	return false
}
