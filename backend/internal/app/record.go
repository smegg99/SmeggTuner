package app

import (
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"

	recordservice "smegg.me/smeggtuner/backend/services/record"
	sessionservice "smegg.me/smeggtuner/backend/services/session"
)

// engineLink is what record's state drives on the tuner.
type engineLink interface {
	Recalibrate()
	SetRecording(on bool)
}

// recordLink turns record's state into engine controls; wired here so record stays ignorant of the tuner.
type recordLink struct {
	mu     sync.Mutex
	state  func() recordservice.StateDTO
	engine engineLink
	armed  bool
}

// follow relearns the room on the rising edge of Armed and records while armed in a session. It reads
// the state now, not the event's copy: Wails runs every listener in its own goroutine, so events can
// arrive out of order.
func (l *recordLink) follow() {
	l.mu.Lock()
	defer l.mu.Unlock()
	st := l.state()
	if st.Armed && !l.armed {
		l.engine.Recalibrate()
	}
	l.armed = st.Armed
	l.engine.SetRecording(st.Armed && st.SessionID != "")
}

func setupRecordEngine(app *application.App) {
	link := &recordLink{state: recordSvc.State, engine: tunerSvc}
	app.Event.On(recordservice.EventState, func(*application.CustomEvent) { link.follow() })
}

// setupRecordSessionState calls SessionChanged (not PublishState) on session change, since a new session starts in warm-up.
func setupRecordSessionState(app *application.App) {
	app.Event.On(sessionservice.EventActive, func(*application.CustomEvent) {
		recordSvc.SessionChanged()
	})
}
