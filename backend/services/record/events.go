package record

import (
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	coresession "smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
	audiosvc "smegg.me/smeggtuner/backend/services/audio"
)

const (
	// EventState carries a StateDTO when record mode arms/disarms or its session changes.
	EventState = "record:state"
	// EventTable carries the whole TableDTO when the open pass changes.
	EventTable = "record:table"
)

// ServiceError is the i18n-keyed error shape every service hands the frontend.
type ServiceError = audiosvc.ServiceError

var (
	// ErrTakeNotFound reports a take index the pass does not hold.
	ErrTakeNotFound = &ServiceError{Key: "record.error.takeNotFound"}
)

// StateDTO is record mode as the UI sees it.
type StateDTO struct {
	SessionID string `json:"sessionId"` // "" when no session is open
	// Readings is how many voices the session has heard.
	Readings int `json:"readings"`
	// Armed says whether takes are being saved; not persisted.
	Armed bool `json:"armed"`
}

// RowDTO is one note of a pass as the tuning table prints it: one row per note.
type RowDTO struct {
	Note     tuning.Note `json:"note"`
	NoteName string      `json:"noteName"`

	// Register is the switch this note was played on. Empty when no registers are described.
	Register string `json:"register,omitempty"`

	// Banks is the rank each slot measured: slot i is bank Banks[i]. Empty when unmapped, and the
	// table then numbers the slots.
	Banks []coresession.Bank `json:"banks,omitempty"`
	// Bass says the row came from the bass side; it belongs to the table's bass section.
	Bass bool `json:"bass,omitempty"`
	// Cols is the table column each slot belongs in: slot i sits in column Cols[i]. Empty when the
	// mapping is uncertain, and the row falls back to slot positions.
	Cols []int `json:"cols,omitempty"`
	// Take is the index into the session's readings, which is what an edit or removal aims at.
	Take int       `json:"take"`
	At   time.Time `json:"at"`
	// Manual marks a row whose value was typed rather than heard.
	Manual bool `json:"manual"`
	// Slots is how many slots the take measured; Reeds and Beats hold rows only for the ones that
	// sounded, each naming its slot, in reference units with the Hz beside them; HasGoal false
	// means the goal prints blank.
	Slots int                `json:"slots"`
	Reeds []target.ReedError `json:"reeds"`
	Beats []target.BeatError `json:"beats"`
}

// TableDTO is a whole pass, read against the goal.
type TableDTO struct {
	SessionID string `json:"sessionId"`
	// A4 is the reference every reading in the session was measured against.
	A4 float64 `json:"a4"`
	// Transpose is the transposition they were measured with: the instrument's C sounded as C+Transpose.
	Transpose int `json:"transpose"`
	// Latest is the take captured last, which Undo removes next; -1 on an empty pass. Takes are one
	// per voice, so the next take is a new row or replaces its voice's: there is no next box.
	Latest int `json:"latest"`
	// ReedCount is what the instrument sounds; a row can hold fewer reeds than that.
	ReedCount int `json:"reedCount"`
	// Banks is the instrument's, in card order: the columns this table prints; empty when undescribed.
	Banks []coresession.Bank `json:"banks,omitempty"`
	// Tolerance and BeatTolerance are the judging windows in reference units.
	Tolerance     float64  `json:"tolerance"`
	BeatTolerance float64  `json:"beatTolerance"`
	Rows          []RowDTO `json:"rows"`
	// BassFeet is the bass section's columns - the machine's ranks by foot, largest first - and
	// BassReedCount their number. Zero-valued when the instrument declares no bass machine.
	BassFeet      []int `json:"bassFeet,omitempty"`
	BassReedCount int   `json:"bassReedCount,omitempty"`
}

func init() {
	application.RegisterEvent[StateDTO](EventState)
	application.RegisterEvent[TableDTO](EventTable)
}

// emitEvent is the test seam for frontend events; with no application running it is a no-op.
var emitEvent = func(name string, data any) {
	if app := application.Get(); app != nil {
		app.Event.Emit(name, data)
	}
}
