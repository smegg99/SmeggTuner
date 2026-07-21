// Opening a session must publish EventState before a note is played; it used to go
// out only from publishTable, so an opened-but-unplayed session read as no session.
package record

import (
	"testing"
)

// A freshly opened session names itself before any reading.
func TestAnOpenSessionPublishesItsStateBeforeAnyReading(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)

	seen := capture(t)
	svc.PublishState()

	var states []StateDTO
	for _, e := range *seen {
		if e.name != EventState {
			continue
		}
		state, ok := e.data.(StateDTO)
		if !ok {
			t.Fatalf("EventState carried %T, but the tray asserts record.StateDTO", e.data)
		}
		states = append(states, state)
	}

	if len(states) != 1 {
		t.Fatalf("published %d EventState, want exactly one", len(states))
	}
	if states[0].SessionID == "" {
		t.Error("EventState carried no SessionID, so the light says to open a session while one is open")
	}
	if states[0].Readings != 0 {
		t.Errorf("Readings = %d on a session nobody has played into, want 0", states[0].Readings)
	}
}

// With nothing open it says nothing is open, which is not an error.
func TestNoSessionPublishesAnEmptyState(t *testing.T) {
	svc, _ := services(t)

	seen := capture(t)
	svc.PublishState()

	for _, e := range *seen {
		if e.name != EventState {
			continue
		}
		if state := e.data.(StateDTO); state.SessionID != "" {
			t.Errorf("EventState named session %q with none open", state.SessionID)
		}
	}
}

// PublishState ships the state, not the table, to avoid reshipping per curve-drag frame.
func TestPublishStateDoesNotShipTheWholeTable(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)
	svc.SetArmed(true)
	lock(svc, 60, -8, 0)

	seen := capture(t)
	svc.PublishState()

	for _, e := range *seen {
		if e.name == EventTable {
			t.Fatal("PublishState shipped a TableDTO; it is called on every session change")
		}
	}
}
