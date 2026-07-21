package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// Version is the schema version stamped into every session file; an unknown version is refused.
const Version = 1

const (
	// LegacyFileExt is the pre-datastore session format (bare JSON); import reads it, nothing writes it.
	LegacyFileExt = ".session.json"

	filePerm = 0o644
	dirPerm  = 0o755
)

var ErrVersion = errors.New("session: unsupported file version")

// sessionFile is the serialized shape: a session behind a version field.
type sessionFile struct {
	V int `json:"v"`
	Session
}

// Summary is a session as the list screen draws it.
type Summary struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Instrument Instrument `json:"instrument"`
	// InstrumentID links to the shelf entry so the list can find the photo.
	InstrumentID string  `json:"instrumentId,omitempty"`
	A4           float64 `json:"a4"`
	// Readings is how many takes the session holds.
	Readings int       `json:"readings"`
	HasCurve bool      `json:"hasCurve"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

// Summarize returns the session as the list screen draws it.
func (s *Session) Summarize() Summary {
	return Summary{
		ID:           s.ID,
		Name:         s.Name,
		Instrument:   s.Instrument,
		InstrumentID: s.InstrumentID,
		A4:           s.A4,
		Readings:     len(s.Takes),
		HasCurve:     s.Curve != nil,
		Created:      s.Created,
		Updated:      s.Updated,
	}
}

// Read loads a bare legacy session file (a pre-datastore import).
func Read(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s, err := decodeSession(data)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return s, nil
}

func decodeSession(data []byte) (*Session, error) {
	var f sessionFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	if f.V != Version {
		return nil, fmt.Errorf("%w: %d", ErrVersion, f.V)
	}
	s := f.Session
	if s.Curve != nil {
		s.Curve.Sort() // a hand-edited file may have its anchors in any order
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return &s, nil
}

// ValidID accepts what NewID produces and nothing that could escape a file path or row id.
func ValidID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case c == '-' || c == '_':
		default:
			return false
		}
	}
	return true
}
