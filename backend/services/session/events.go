package session

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	// EventActive carries an ActiveDTO whenever the active session changes; Session is null when none is open.
	EventActive = "session:active"
	// EventSaveFailed carries an ErrorDTO when a background write to disk failed.
	EventSaveFailed = "session:saveFailed"
)

func init() {
	application.RegisterEvent[ActiveDTO](EventActive)
	application.RegisterEvent[ErrorDTO](EventSaveFailed)
}

// emitEvent is the seam tests replace; a no-op when no app is running.
var emitEvent = func(name string, data any) {
	if app := application.Get(); app != nil {
		app.Event.Emit(name, data)
	}
}
