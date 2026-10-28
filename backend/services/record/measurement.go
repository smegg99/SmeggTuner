package record

import (
	"time"

	"smegg.me/smeggtuner/backend/common/i18n"

	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/core/dsp"
	coresession "smegg.me/smeggtuner/backend/core/session"
)

// OnMeasurement is the engine's stream, handed over by services/tuner. A take is filed exactly when
// the engine says the analysis took one. It must not block: it runs on the goroutine that feeds the UI.
func (s *Service) OnMeasurement(m dsp.Measurement) {
	if !taken(m) || !s.Armed() {
		return
	}
	// No session, or a finished one, is an ordinary state, not a fault.
	if err := s.sessions.UpsertTake(coresession.TakeFrom(m, time.Now())); err != nil {
		return
	}
	logger.Debug(i18n.KeyLogRecordTake, logger.Int("note", int(m.Note)), logger.Int("reeds", len(m.Reeds)))
	_, _ = s.publishTable()
}

// taken reports whether m is the analysis that took a reading; a chord frame (it carries a chord name) is never a take.
func taken(m dsp.Measurement) bool {
	return m.Taken && m.Label == "" && len(m.Reeds) > 0
}
