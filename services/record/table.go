package record

import (
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

	for _, r := range sess.Display() {
		m := measurementOf(r.Take)
		reeds := target.Errors(m, d.Curve, d.A4, tol)
		rows = append(rows, RowDTO{
			Note:          r.Note,
			NoteName:      r.Note.Name(naming),
			Register:      r.Register,
			Banks:         banksOf(d.Instrument, r.Take, len(reeds)),
			Take:          r.Index,
			At:            r.Take.At,
			Manual:        r.Take.Manual,
			ReedsMerged:   r.Take.ReedsMerged,
			ReedsFromBeat: r.Take.ReedsFromBeat,
			Reeds:         reeds,
			Beats:         target.BeatErrors(m, d.Curve, d.A4, beatTol),
		})
	}
	return &TableDTO{
		SessionID:     d.SessionID,
		A4:            d.A4,
		ReedCount:     d.ReedCount,
		Banks:         d.Instrument.Banks,
		Tolerance:     tol,
		BeatTolerance: beatTol,
		Rows:          rows,
	}
}

// banksOf maps each reed of a row to its column: reeds low to high, the register's
// banks in the same order. Returns nil (the table then numbers reeds) when no
// register is named, the register is gone, or the reed count does not match it.
func banksOf(i coresession.Instrument, t coresession.Take, reeds int) []coresession.Bank {
	if t.Register == "" || reeds == 0 {
		return nil
	}
	r, ok := i.Register(t.Register)
	if !ok || r.ReedCount() != reeds {
		return nil
	}
	return r.Banks
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
