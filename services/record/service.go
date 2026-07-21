// Package record is record mode: while armed, engine locks are written to the open session as takes.
package record

import (
	"math"
	"sync"

	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
	sessionsvc "smegg.me/smeggtuner/services/session"
)

// Service owns record mode. OnMeasurement is called by services/tuner, not the frontend.
type Service struct {
	sessions *sessionsvc.Service

	mu sync.Mutex
	// locked/note are the lock edge; only the rise of a lock is a take.
	locked bool
	note   tuning.Note
	// armed says whether locks are saved; lastSession scopes it to the session it was made in (see SessionChanged).
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

// Undo drops the take captured last; on an empty pass it does nothing.
func (s *Service) Undo() (*TableDTO, error) {
	if _, err := s.sessions.UndoTake(); err != nil {
		return nil, err
	}
	logger.Debug(logger.MsgRecordUndone)
	return s.publishTable()
}

// Clear empties the session of readings, keeping the session and its curve.
func (s *Service) Clear() (*TableDTO, error) {
	if err := s.sessions.ClearTakes(); err != nil {
		return nil, err
	}
	logger.Info(logger.MsgRecordCleared)
	return s.publishTable()
}

// EditReed writes a reed of a take by hand and marks the take manual; value is in
// unit "cent" or "hz". A merged take with no per-reed answer cannot be edited per reed.
func (s *Service) EditReed(take, reed int, value float64, unit string) (*TableDTO, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return nil, sessionsvc.ErrInvalidValue
	}
	d, err := s.sessions.Data()
	if err != nil {
		return nil, err
	}
	if take < 0 || take >= len(d.Takes) {
		return nil, ErrTakeNotFound
	}
	t := d.Takes[take]
	if t.ReedsMerged && !t.ReedsFromBeat {
		logger.Warn(logger.MsgRecordRejected, logger.Int("take", take), logger.Int("reed", reed))
		return nil, ErrReedsMerged
	}
	if reed < 0 || reed >= len(t.Reeds) {
		return nil, sessionsvc.ErrInvalidReed
	}

	cents := value
	switch unit {
	case string(target.UnitCents):
	case string(target.UnitHz):
		cents = target.CentsFromHz(t.Note, value, d.A4)
	default:
		return nil, sessionsvc.ErrInvalidUnit
	}
	if math.IsNaN(cents) || math.IsInf(cents, 0) {
		return nil, sessionsvc.ErrInvalidValue
	}

	if err := s.sessions.SetTakeReed(take, reed, cents); err != nil {
		return nil, err
	}
	logger.Info(logger.MsgRecordEdited,
		logger.Int("note", int(t.Note)), logger.Int("reed", reed), logger.Any("cents", cents))
	return s.publishTable()
}
