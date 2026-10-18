package record

import (
	appconfig "smegg.me/smeggtuner/backend/common/config"
	coresession "smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
	sessionsvc "smegg.me/smeggtuner/backend/services/session"
)

// tableOf joins a pass's takes to the goal curve: one row per voice, every number computed by
// core/target in reference units (and Hz) against the pass's own reference and transposition.
func (s *Service) tableOf(d *sessionsvc.ReadingData) *TableDTO {
	cfg := appconfig.Get().Tuner
	naming := tuning.ParseNaming(cfg.ScaleNaming)
	tol, beatTol := target.Tolerances(cfg.Tolerance, cfg.BeatTolerance)
	tol, beatTol = d.Instrument.Tolerances(tol, beatTol)
	sess := coresession.Session{Takes: d.Takes}
	rows := make([]RowDTO, 0, len(d.Takes))

	bassFeet := coresession.BassFeet(d.Instrument.BassReeds)
	for _, r := range sess.Display() {
		slots := len(r.Take.Reeds)
		// The columns first: the goal is read at the slot's own column, so a register short of the
		// instrument's full one is not judged against another rank's goal.
		var banks []coresession.Bank
		if !r.Take.Bass {
			banks = coresession.SlotBanks(d.Instrument.TakeBanks(r.Take), slots)
		}
		cols := d.Instrument.Columns(r.Take)
		reading := r.Take.Reading(d.Transpose)
		rows = append(rows, RowDTO{
			Note:     r.Note,
			NoteName: r.Note.Name(naming),
			Register: r.Register,
			Bass:     r.Take.Bass,
			Take:     r.Index,
			At:       r.Take.At,
			Manual:   r.Take.Manual,
			Banks:    banks,
			Cols:     cols,
			Slots:    slots,
			Reeds:    target.Errors(reading, d.Curve, d.A4, tol, cols),
			Beats:    target.BeatErrors(reading, d.Curve, d.A4, beatTol, cols),
		})
	}
	return &TableDTO{
		SessionID:     d.SessionID,
		A4:            d.A4,
		Transpose:     d.Transpose,
		Latest:        sess.Latest(),
		ReedCount:     d.ReedCount,
		Banks:         d.Instrument.Banks,
		Tolerance:     tol,
		BeatTolerance: beatTol,
		Rows:          rows,
		BassFeet:      bassFeet,
		BassReedCount: len(bassFeet),
	}
}
