package record

import (
	"time"

	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/dsp"
	coresession "smegg.me/smeggtuner/core/session"
)

// OnMeasurement is the engine's stream, handed over by services/tuner. It must
// not block: it runs on the goroutine that feeds the UI.
func (s *Service) OnMeasurement(m dsp.Measurement) {
	if !s.edge(m) {
		return
	}
	if !s.Armed() {
		return
	}
	// No session, or a finished one, is an ordinary state, not a fault.
	if err := s.sessions.UpsertTake(coresession.TakeFrom(m, time.Now())); err != nil {
		return
	}
	logger.Debug(logger.MsgRecordTake,
		logger.Int("note", int(m.Note)), logger.Int("reeds", len(m.Reeds)),
		logger.Bool("merged", !m.ReedsSeparated))
	_, _ = s.publishTable()
}

// edge reports whether this measurement is a new lock (its rise) worth recording.
// A heartbeat carries no note (ScalePitch 0) and must not be read as an unlock.
func (s *Service) edge(m dsp.Measurement) bool {
	if m.ScalePitch <= 0 {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rise := m.Locked && (!s.locked || m.Note != s.note)
	s.locked, s.note = m.Locked, m.Note
	return rise && len(m.Reeds) > 0
}
