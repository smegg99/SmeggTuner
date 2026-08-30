package report

import (
	"time"

	"smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// Build lays the session's readings against its goal curve into a sheet, every deviation in opts.Unit.
func Build(s *session.Session, opts Options) (*Report, error) {
	if s == nil {
		return nil, ErrNoSession
	}
	if len(s.Takes) == 0 {
		return nil, ErrNoReadings
	}
	unit := opts.Unit
	switch unit {
	case "":
		unit = target.UnitReference
	case target.UnitReference, target.UnitCents:
	default:
		return nil, ErrUnit
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

	display, bassDisplay := splitBySide(s.Display())
	judge := judging{curve: s.Curve, a4: s.A4, transpose: s.Transpose, tol: tol, beatTol: beatTol, unit: unit}
	rows, pairs := buildRows(display, judge, banks, reedCount, opts.Naming, columnsFor(s.Instrument))

	rep := &Report{
		Identity: Identity{
			Session:   s.Name,
			Accordion: s.Instrument.Name,
			Serial:    s.Instrument.Serial,
			ReedCount: reedCount,
			Notes:     s.Notes,
			Registers: s.Instrument.Registers,
			HasCurve:  s.Curve != nil,
		},
		Session: SessionInfo{
			At:        s.Updated,
			A4:        s.A4,
			Transpose: s.Transpose,
			Readings:  len(s.Takes),
		},
		Letterhead: opts.Letterhead,
		Date:       opts.Date,
		Generated:  opts.Now,
		Unit:       unit,
		Pairs:      pairs,
		Rows:       rows,
		Banks:      banks,
	}
	rep.MultiRegister = manyRegisters(rows)
	if s.Transpose != 0 {
		c := tuning.Note(60)
		rep.Session.Key, rep.Session.Sounds = c.PitchName(opts.Naming), (c + tuning.Note(s.Transpose)).PitchName(opts.Naming)
	}
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
	// The bass side gets its own table: different columns (the machine's ranks by foot), and the
	// buttons' real pitches for rows. It shares the page, the summary and the layout decision.
	if len(bassDisplay) > 0 {
		count := s.Instrument.BassReeds
		for _, d := range bassDisplay {
			if n := len(d.Take.Reeds); n > count {
				count = n
			}
		}
		feet := session.BassFeet(count)
		bassRows, bassPairs := buildRows(bassDisplay, judge, nil, count, opts.Naming,
			bassColumnsFor(s.Instrument, feet))
		part := &BassPart{Feet: feet, Pairs: bassPairs, Rows: bassRows}
		for i := 1; i <= count; i++ {
			part.Reeds = append(part.Reeds, i)
		}
		part.MultiRegister = manyRegisters(bassRows)
		rep.Bass = part
	}

	rep.Summary = summarize(allRows(rep), unit.FromReference(tol, 0), unit.FromReference(beatTol, 0))
	rep.Layout = layoutFor(rep.Columns())
	rep.Graph = graph(rep, s.Curve)
	return rep, nil
}

// splitBySide deals a pass's readings to the keyboard they came from.
func splitBySide(display []session.DisplayRow) (treble, bass []session.DisplayRow) {
	for _, d := range display {
		if d.Take.Bass {
			bass = append(bass, d)
		} else {
			treble = append(treble, d)
		}
	}
	return treble, bass
}

// allRows is every row of the sheet, both keyboards, for the summary.
func allRows(r *Report) []Row {
	if r.Bass == nil {
		return r.Rows
	}
	return append(append([]Row(nil), r.Rows...), r.Bass.Rows...)
}

// columnsFor maps a treble take's slots onto the card's columns by bank; nil on any uncertain
// mapping, and the slots fall back to positions.
func columnsFor(i session.Instrument) func(session.Take) []int {
	return func(t session.Take) []int {
		return session.BankColumns(i.Banks, i.TakeBanks(t), len(t.Reeds))
	}
}

// bassColumnsFor maps a bass take's slots onto the machine's rank columns by foot; nil on any
// uncertain mapping, and the slots fall back to positions.
func bassColumnsFor(i session.Instrument, feet []int) func(session.Take) []int {
	return func(t session.Take) []int {
		if !t.Bass {
			return nil
		}
		return session.FootColumns(feet, i.SoundingFeet(t.Register), len(t.Reeds))
	}
}
