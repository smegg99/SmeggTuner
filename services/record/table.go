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

	bassFeet := coresession.BassFeet(d.Instrument.BassReeds)
	for _, r := range sess.Display() {
		m := measurementOf(r.Take)
		// The columns first: the goal is read at the reed's own column, so a register short of the
		// instrument's full one is not judged against another rank's goal.
		var banks []coresession.Bank
		var cols []int
		if r.Take.Bass {
			cols = coresession.FootColumns(bassFeet, d.Instrument.SoundingFeet(r.Take.Register), r.Take.Reeds)
		} else {
			sounding := d.Instrument.TakeBanks(r.Take)
			banks = coresession.AssignBanks(sounding, r.Take.Reeds)
			cols = coresession.BankColumns(d.Instrument.Banks, sounding, r.Take.Reeds)
		}
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
			Banks:         banks,
			Cols:          cols,
			Bands:         r.Take.Bands,
			Reeds:         target.Errors(m, d.Curve, d.A4, tol, cols),
			Beats:         target.BeatErrors(m, d.Curve, d.A4, beatTol, cols),
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

// measurementOf reads a take back as the measurement it came from. It carries no
// scale pitch, so core/target measures it against the pass's own A4; the separation
// flags decide whether the beats are derived from the reeds or reported as measured.
func measurementOf(t coresession.Take) dsp.Measurement {
	return dsp.Measurement{
		Note:           t.Note,
		Reeds:          t.Reeds,
		Beats:          t.Beats,
		Bands:          t.Bands,
		ReedsSeparated: !t.ReedsMerged,
		ReedsFromBeat:  t.ReedsFromBeat,
	}
}
