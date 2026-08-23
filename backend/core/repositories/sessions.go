package repositories

import (
	"fmt"
	"time"

	"smegg.me/smeggtuner/backend/core/session"
)

// SessionRepository owns the sessions table.
type SessionRepository struct{ *Repository[session.Session] }

func GetSessionRepository() *SessionRepository {
	return &SessionRepository{Repository: New[session.Session]()}
}

// List returns a summary of every session, most recently updated first.
func (r *SessionRepository) List() ([]session.Summary, error) {
	var all []session.Session
	if err := db().Order("updated DESC").Find(&all).Error; err != nil {
		return nil, err
	}
	out := make([]session.Summary, 0, len(all))
	for i := range all {
		out = append(out, all[i].Summarize())
	}
	return out, nil
}

// Get loads one session; a malformed id answers ErrNotFound.
func (r *SessionRepository) Get(id string) (*session.Session, error) {
	if !session.ValidID(id) {
		return nil, fmt.Errorf("%w: %q", ErrNotFound, id)
	}
	return r.GetByID(id)
}

// Save validates, stamps Updated and upserts the session whole.
func (r *SessionRepository) Save(s *session.Session) error {
	if !session.ValidID(s.ID) {
		return fmt.Errorf("%w: %q", session.ErrBadID, s.ID)
	}
	if err := s.Validate(); err != nil {
		return err
	}
	s.Updated = time.Now()
	return db().Save(s).Error
}

func (r *SessionRepository) Delete(id string) error {
	if !session.ValidID(id) {
		return fmt.Errorf("%w: %q", ErrNotFound, id)
	}
	return r.Repository.Delete(id)
}

// LinkInstrument updates only the shelf link, never upserting a stale session snapshot.
func (r *SessionRepository) LinkInstrument(id, instrumentID string) error {
	res := db().Model(&session.Session{}).Where("id = ?", id).Update("instrument_id", instrumentID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
