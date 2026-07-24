package record

import (
	"slices"

	appconfig "smegg.me/smeggtuner/common/config"
	"smegg.me/smeggtuner/core/dsp"
	coresession "smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
	sessionsvc "smegg.me/smeggtuner/services/session"
)

// tableOf joins a pass's takes to the goal curve: one row per note (the last take
// of it), every number computed by core/target against the pass's own reference.
func (s *Service) tableOf(d *sessionsvc.ReadingData) *TableDTO {
	cfg := appconfig.Get().Tuner
	naming := tuning.ParseNaming(cfg.ScaleNaming)
	tol, beatTol := target.Tolerances(cfg.Tolerance, cfg.BeatTolerance)
	// The instrument's own tolerances where it has them.
	tol, beatTol = d.Instrument.Tolerances(tol, beatTol)
	sess := coresession.Session{Takes: d.Takes}
	rows := make([]RowDTO, 0, len(d.Takes))

	bassFeet := coresession.BassFeet(d.Instrument.BassReeds)
	for _, r := range sess.Display() {
		m := measurementOf(r.Take)
		reeds := target.Errors(m, d.Curve, d.A4, tol)
		row := RowDTO{
			Note:          r.Note,
			NoteName:      r.Note.Name(naming),
			Register:      r.Register,
			Bass:          r.Take.Bass,
			Take:          r.Index,
			At:            r.Take.At,
			Manual:        r.Take.Manual,
			ReedsMerged:   r.Take.ReedsMerged,
			ReedsFromBeat: r.Take.ReedsFromBeat,
			Reeds:         reeds,
			Beats:         target.BeatErrors(m, d.Curve, d.A4, beatTol),
		}
		if r.Take.Bass {
			row.Cols = feetColsOf(d.Instrument, r.Take, bassFeet)
		} else {
			row.Banks = banksOf(d.Instrument, r.Take, len(reeds))
			row.Cols = bankColsOf(d.Instrument, row.Banks)
		}
		rows = append(rows, row)
	}
	return &TableDTO{
		SessionID:     d.SessionID,
		A4:            d.A4,
		ReedCount:     d.ReedCount,
		Banks:         d.Instrument.Banks,
		Tolerance:     tol,
		BeatTolerance: beatTol,
		Rows:          rows,
		BassFeet:      bassFeet,
		BassReedCount: len(bassFeet),
	}
}

// feetColsOf places a bass take's reeds in the machine's rank columns by foot; nil when uncertain,
// and the row falls back to positions.
func feetColsOf(i coresession.Instrument, t coresession.Take, feet []int) []int {
	tf := i.TakeFeet(t)
	if len(tf) == 0 || len(feet) == 0 {
		return nil
	}
	cols := make([]int, len(tf))
	for n, f := range tf {
		col := slices.Index(feet, f)
		if col < 0 {
			return nil
		}
		cols[n] = col
	}
	return cols
}

// bankColsOf places a treble take's reeds in the instrument's bank columns; nil when unmapped.
func bankColsOf(i coresession.Instrument, banks []coresession.Bank) []int {
	if len(banks) == 0 || len(i.Banks) == 0 {
		return nil
	}
	cols := make([]int, len(banks))
	for n, b := range banks {
		col := slices.Index(i.Banks, b)
		if col < 0 {
			return nil
		}
		cols[n] = col
	}
	return cols
}

// banksOf maps each reed of a row to its column: each reed claims a register bank in its own
// octave (see session.AssignBanks), so a take short a rank still names the ranks it has. Returns
// nil (the table then numbers reeds) when no register is named, the register is gone, or the
// claim fails.
func banksOf(i coresession.Instrument, t coresession.Take, reeds int) []coresession.Bank {
	if t.Register == "" || reeds == 0 {
		return nil
	}
	r, ok := i.Register(t.Register)
	if !ok {
		return nil
	}
	return coresession.AssignBanks(r.Banks, t.Reeds)
}

// measurementOf reads a take back as the measurement it came from. It carries no
// scale pitch, so core/target measures it against the pass's own A4; the separation
// flags decide whether the beats are derived from the reeds or reported as measured.
func measurementOf(t coresession.Take) dsp.Measurement {
	return dsp.Measurement{
		Note:           t.Note,
		Reeds:          t.Reeds,
		Beats:          t.Beats,
		ReedsSeparated: !t.ReedsMerged,
		ReedsFromBeat:  t.ReedsFromBeat,
	}
}
