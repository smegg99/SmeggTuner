// Package record is record mode: while armed, every take the engine makes is written to the open session.
package record

import (
	"math"
	"sync"

	"smegg.me/smeggtuner/backend/common/i18n"

	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
	sessionsvc "smegg.me/smeggtuner/backend/services/session"
)

// Service owns record mode. OnMeasurement is called by services/tuner, not the frontend.
type Service struct {
	sessions *sessionsvc.Service

	mu sync.Mutex
	// armed says whether takes are saved; lastSession scopes it to the session it was made in (see SessionChanged).
	armed       bool
	lastSession string
}

func New(sessions *sessionsvc.Service) *Service {
	return &Service{sessions: sessions}
}

// State reports which session the readings are landing in.
func (s *Service) State() StateDTO { return s.state() }

// Table is the session's readings against the goal curve.
func (s *Service) Table() (*TableDTO, error) {
	d, err := s.sessions.Data()
	if err != nil {
		return nil, err
	}
	return s.tableOf(d), nil
}

// Undo is Backspace: it drops the take captured last, and on an empty pass it leaves record mode.
func (s *Service) Undo() (*TableDTO, error) {
	removed, err := s.sessions.UndoTake()
	if err != nil {
		return nil, err
	}
	if !removed {
		s.SetArmed(false)
		return s.Table()
	}
	logger.Debug(i18n.KeyLogRecordUndone)
	return s.publishTable()
}

// Clear empties the session of readings, keeping the session and its curve.
func (s *Service) Clear() (*TableDTO, error) {
	if err := s.sessions.ClearTakes(); err != nil {
		return nil, err
	}
	logger.Info(i18n.KeyLogRecordCleared)
	return s.publishTable()
}

// EditReed writes a sounding slot of a take by hand and marks the take manual; value is in unit
// "reference", "cent" or "hz" and is stored in reference units at the slot's own sounding pitch.
func (s *Service) EditReed(take, reed int, value float64, unit string) (*TableDTO, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return nil, sessionsvc.ErrInvalidValue
	}
	u := target.Unit(unit)
	if !u.Valid() {
		return nil, sessionsvc.ErrInvalidUnit
	}
	d, err := s.sessions.Data()
	if err != nil {
		return nil, err
	}
	if take < 0 || take >= len(d.Takes) {
		return nil, ErrTakeNotFound
	}
	t := d.Takes[take]
	if reed < 0 || reed >= len(t.Reeds) || !t.Reeds[reed].Present {
		logger.Warn(i18n.KeyLogRecordRejected, logger.Int("take", take), logger.Int("reed", reed))
		return nil, sessionsvc.ErrInvalidReed
	}

	nominal := (t.Note + tuning.Note(d.Transpose+t.Reeds[reed].Octave)).Freq(d.A4)
	dev := u.ToReference(value, nominal)
	if math.IsNaN(dev) || math.IsInf(dev, 0) || meter.FrequencyHz(nominal, dev) <= 0 {
		return nil, sessionsvc.ErrInvalidValue
	}

	if err := s.sessions.SetTakeReed(take, reed, dev); err != nil {
		return nil, err
	}
	logger.Info(i18n.KeyLogRecordEdited,
		logger.Int("note", int(t.Note)), logger.Int("reed", reed), logger.Any("dev", dev))
	return s.publishTable()
}
