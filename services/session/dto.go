package session

import (
	"time"

	"smegg.me/smeggtuner/core/dsp"
	coresession "smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/target"
)

// SessionDTO is the active session as the UI draws it.
type SessionDTO struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Instrument coresession.Instrument `json:"instrument"`
	// InstrumentID links to a shelf instrument (a link, not a dependency); an imported session may
	// reference one not present here.
	InstrumentID string  `json:"instrumentId"`
	A4           float64 `json:"a4"`
	// Curve is the goal. Null means no goal yet, a legal state: the tuner is then a pure indicator.
	Curve *target.Curve `json:"curve"`
	// Readings is how many voices the session has heard; the readings themselves travel through services/record.
	Readings int    `json:"readings"`
	Notes    string `json:"notes"`
	// Bench is the current setup (pulled register). Not saved with the session, but every take is stamped with it.
	Bench   BenchDTO  `json:"bench"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

// ActiveDTO is the EventActive payload. Session is null when none is open.
type ActiveDTO struct {
	Session *SessionDTO `json:"session"`
}

// ErrorDTO is the EventSaveFailed payload: an i18n key.
type ErrorDTO struct {
	Key string `json:"key"`
}

// NewSessionDTO is what Create needs. The A4 is the instrument's reference, not the app's.
type NewSessionDTO struct {
	Name       string                 `json:"name"`
	Instrument coresession.Instrument `json:"instrument"`
	// InstrumentID is the one off the shelf this was started from, or empty for one described on the spot.
	InstrumentID string `json:"instrumentId"`
	Notes        string `json:"notes"`
}

// FitDTO is a curve recovered from a pass, plus the outliers the fit refused.
type FitDTO struct {
	Curve    *target.Curve    `json:"curve"`
	Outliers []target.Outlier `json:"outliers"`
	Used     int              `json:"used"`   // readings the curve was fitted to
	Merged   int              `json:"merged"` // readings dropped for un-separated reeds
}

// Goal is what the active session imposes on the engine, and what services/tuner reads. The zero
// Goal (no curve, no reference, no reeds) is the first-class "no session open" state. Curve is a
// snapshot the caller may read without a lock (see cloneCurve).
type Goal struct {
	Curve *target.Curve `json:"curve"`
	A4    float64       `json:"a4"`
	// Reeds is what the instrument sounds (1..8), not what the engine resolves.
	Reeds int `json:"reeds"`
	// Banks is the pulled register's ranks in card order, or nil when none is pulled (or the
	// instrument predates banks). The tuner maps these onto the engine's octave layout.
	Banks []coresession.Bank `json:"banks"`

	// BassFeet is what sounds when the bench faces the bass side, largest foot first; nil while it
	// faces the treble. The tuner maps these onto the engine's octave layout the way Banks map.
	BassFeet []int `json:"bassFeet"`

	// Profile is the rank voices this session's calibration takes taught (see Session.Profile),
	// BassProfiles the bass side's (keyed by foot; the octave depends on the pulled register), and
	// ProfileRev the fingerprint that says when to re-read them.
	Profile      []dsp.RankProfile         `json:"profile"`
	BassProfiles []coresession.BassProfile `json:"bassProfiles"`
	ProfileRev   int64                     `json:"profileRev"`

	// Tolerance and BeatTolerance are this accordion's judging windows, in cents, or zero for the app default.
	Tolerance     float64 `json:"tolerance"`
	BeatTolerance float64 `json:"beatTolerance"`
}

// ReadingData is a session's readings plus what services/record builds its table from, taken under
// one lock so takes, reference and curve cannot disagree.
type ReadingData struct {
	SessionID string
	A4        float64 // the reference every reading in it was measured against
	ReedCount int     // the instrument's, so the table has a column per reed it sounds
	// Instrument maps a take's register to its banks to a column; nothing else can.
	Instrument coresession.Instrument
	Curve      *target.Curve
	Takes      []coresession.Take
}
