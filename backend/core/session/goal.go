// backend/core/session/goal.go
package session

import "smegg.me/smeggtuner/backend/core/target"

// Goal is what the active session imposes on the engine, and what services/tuner reads. The zero
// Goal (no curve, no reference, no reeds) is the first-class "no session open" state. Curve is a
// snapshot the caller may read without a lock (see cloneCurve).
type Goal struct {
	Curve *target.Curve `json:"curve"`
	A4    float64       `json:"a4"`
	// Transpose is the session's, 0..11: what the engine's band table moves by while it is open.
	Transpose int `json:"transpose"`
	// Reeds is what the instrument sounds (1..8), not what the engine resolves.
	Reeds int `json:"reeds"`
	// Banks is the pulled register's ranks in card order, or nil when none is pulled (or the
	// instrument predates banks). The tuner maps these onto the engine's octave layout.
	Banks []Bank `json:"banks"`
	// CardBanks is the instrument's own ranks, which are the goal curve's columns. A register that
	// sounds only some of them still has to be judged against its own columns, so a reader needs
	// both lists: Banks says what is sounding, this says where each of those belongs.
	CardBanks []Bank `json:"cardBanks"`

	// BassFeet is what sounds when the bench faces the bass side, largest foot first; nil while it
	// faces the treble. The tuner maps these onto the engine's octave layout the way Banks map.
	BassFeet []int `json:"bassFeet"`
	// CardFeet is the machine's whole ladder - the bass side's columns, as CardBanks is the treble's.
	CardFeet []int `json:"cardFeet"`

	// Tolerance and BeatTolerance are this accordion's judging windows in reference units, or zero for the app default.
	Tolerance     float64 `json:"tolerance"`
	BeatTolerance float64 `json:"beatTolerance"`
}
