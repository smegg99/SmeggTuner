// Package session holds a tuning session: one accordion, its goal curve, and every pass made on it.
package session

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
)

// Reeds sounding per note; the range is deliberate: nothing may assume the musette three.
const (
	MinReeds = 1
	MaxReeds = 8
)

var (
	ErrBadID     = errors.New("session: invalid id")
	ErrReedCount = fmt.Errorf("session: reed count outside %d..%d", MinReeds, MaxReeds)
	ErrA4        = errors.New("session: a4 must be a positive frequency")
	ErrNote      = errors.New("session: note outside the tuning range")
)

// Register is one of the instrument's switches: its name, and which reed banks it sounds when pulled.
type Register struct {
	Name  string `json:"name"`
	Banks []Bank `json:"banks"`
}

// Instrument is the accordion on the bench.
type Instrument struct {
	// Name travels with the session so it survives on a machine that has never seen this instrument.
	Name   string `json:"name,omitempty"`
	Serial string `json:"serial"`

	// Banks is every reed rank this instrument has, in card order: the columns of the printed card.
	Banks []Bank `json:"banks"`

	Registers []Register `json:"registers,omitempty"`

	// Lo and Hi are the keyboard's lowest and highest key; zero means unspecified.
	Lo tuning.Note `json:"lo,omitempty"`
	Hi tuning.Note `json:"hi,omitempty"`

	// BassReeds is how many octave-stacked ranks the bass machine sounds (see bass.go); zero means
	// no bass section described. BassRegisters are its switches, when it has any - a fixed machine
	// (every older instrument) has none and always sounds them all.
	BassReeds     int            `json:"bassReeds,omitempty"`
	BassRegisters []BassRegister `json:"bassRegisters,omitempty"`

	// A4 is this accordion's reference pitch; zero means unspecified and falls back to the app default.
	A4 float64 `json:"a4,omitempty"`

	// Tolerance and BeatTolerance are how tight this accordion is judged, in cents; zero means unspecified.
	Tolerance     float64 `json:"tolerance,omitempty"`
	BeatTolerance float64 `json:"beatTolerance,omitempty"`

	// ReedCount is how many reeds the register currently on the bench sounds: what the engine resolves.
	ReedCount int `json:"reedCount"`
}

// Take is one note, measured once. Reeds and Beats are kept raw so a later curve or tolerance re-reads the same recording.
type Take struct {
	Note tuning.Note `json:"note"`
	At   time.Time   `json:"at"`

	// Register is the switch pulled when this note was played. Empty on a session recorded before the instrument had registers.
	Register string `json:"register,omitempty"`
	// Bass says the take came from the bass side; Register then names a bass register (or nothing,
	// on a fixed machine). Kept apart from the treble registers so their names cannot collide.
	Bass bool `json:"bass,omitempty"`

	Reeds []dsp.ReedMeasure `json:"reeds"`
	Beats []dsp.BeatMeasure `json:"beats,omitempty"`
	// ReedsMerged inverts dsp.Measurement.ReedsSeparated so the zero value is the ordinary case; when set (and ReedsFromBeat is not) the per-reed numbers are lobes of one merged peak.
	ReedsMerged bool `json:"reedsMerged,omitempty"`
	// ReedsFromBeat mirrors dsp.Measurement.ReedsFromBeat: reeds recovered from the beat; set only together with ReedsMerged.
	ReedsFromBeat bool `json:"reedsFromBeat,omitempty"`
	// Manual marks a take whose measured values were hand-edited.
	Manual bool `json:"manual,omitempty"`
}

// Session rows live in the datastore; the aggregate children below travel as JSON columns.
type Session struct {
	ID   string `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`

	// Instrument is a copy, not a reference, so a session stays interpretable on a machine that has never seen this accordion.
	Instrument Instrument `json:"instrument" gorm:"serializer:json"`

	// InstrumentID is a soft link to the shelf entry: nothing here reads it; the app uses it to rejoin the two to show the photograph.
	InstrumentID string `json:"instrumentId,omitempty" gorm:"index"`

	A4    float64       `json:"a4" gorm:"column:a4"`
	Curve *target.Curve `json:"curve,omitempty" gorm:"serializer:json"`

	// Takes is one reading per voice: playing a note again replaces it. A4 may not change once there are any, since it is the reference they were all measured against.
	Takes []Take `json:"takes,omitempty" gorm:"serializer:json"`

	Notes   string    `json:"notes,omitempty"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

func NewID() string { return uuid.NewString() }

// New builds a session with no curve and no passes. A nil Curve is the first-class "no goal yet" state.
func New(name string, inst Instrument, a4 float64) *Session {
	now := time.Now()
	return &Session{
		ID:         NewID(),
		Name:       name,
		Instrument: inst,
		A4:         a4,
		Created:    now,
		Updated:    now,
	}
}

func (s *Session) Validate() error {
	if !ValidID(s.ID) {
		return fmt.Errorf("%w: %q", ErrBadID, s.ID)
	}
	// Only a usable number matters here; the practical 432..442 range belongs to the config schema.
	if math.IsNaN(s.A4) || math.IsInf(s.A4, 0) || s.A4 <= 0 {
		return ErrA4
	}
	if err := validReeds(s.Instrument.ReedCount); err != nil {
		return err
	}
	if err := s.Instrument.validate(); err != nil {
		return err
	}
	// Nil-safe on purpose: no goal yet is a state, not a fault.
	if err := s.Curve.Validate(); err != nil {
		return fmt.Errorf("curve: %w", err)
	}
	for _, t := range s.Takes {
		if !t.Note.Valid() {
			return fmt.Errorf("%w: %d", ErrNote, t.Note)
		}
		if err := s.Instrument.validTake(t); err != nil {
			return err
		}
	}
	return nil
}
