package record

import (
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	coresession "smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
	audiosvc "smegg.me/smeggtuner/services/audio"
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
	// ErrReedsMerged reports a per-reed edit of a take with no per-reed answer.
	ErrReedsMerged = &ServiceError{Key: "record.error.reedsMerged"}
	// ErrTakeNotFound reports a take index the pass does not hold.
	ErrTakeNotFound = &ServiceError{Key: "record.error.takeNotFound"}
)

// StateDTO is record mode as the UI sees it.
type StateDTO struct {
	SessionID string `json:"sessionId"` // "" when no session is open
	// Readings is how many voices the session has heard.
	Readings int `json:"readings"`
	// Armed says whether locks are being saved; not persisted.
	Armed bool `json:"armed"`
}

// RowDTO is one note of a pass as the tuning table prints it: one row per note.
type RowDTO struct {
	Note     tuning.Note `json:"note"`
	NoteName string      `json:"noteName"`

	// Register is the switch this note was played on. Empty when no registers are described.
	Register string `json:"register,omitempty"`

	// Banks is which column each reed belongs in: Reeds[i] is bank Banks[i]. Empty
	// when unmapped, and the table then numbers the reeds.
	Banks []coresession.Bank `json:"banks,omitempty"`
	// Take is the index into the session's readings, which is what an edit or removal aims at.
	Take int       `json:"take"`
	At   time.Time `json:"at"`
	// Manual marks a row whose value was typed rather than heard.
	Manual bool `json:"manual"`
	// ReedsMerged says the spectrum did not tell this take's reeds apart;
	// ReedsFromBeat says they were recovered from the beat anyway. Merged with no
	// recovery means Reeds are lobes of one peak, so show the beat instead.
	ReedsMerged   bool               `json:"reedsMerged"`
	ReedsFromBeat bool               `json:"reedsFromBeat"`
	Reeds         []target.ReedError `json:"reeds"`
	Beats         []target.BeatError `json:"beats"`
}

// TableDTO is a whole pass, read against the goal.
type TableDTO struct {
	SessionID string `json:"sessionId"`
	// A4 is the reference every reading in the session was measured against.
	A4 float64 `json:"a4"`
	// ReedCount is what the instrument sounds; a row can hold fewer reeds than that.
	ReedCount int `json:"reedCount"`
	// Banks is the instrument's, in card order: the columns this table prints; empty when undescribed.
	Banks         []coresession.Bank `json:"banks,omitempty"`
	Tolerance     float64            `json:"tolerance"`
	BeatTolerance float64            `json:"beatTolerance"`
	Rows          []RowDTO           `json:"rows"`
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
