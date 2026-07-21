// The tray asserts record.StateDTO off the event bus; the assertion fails silently,
// so the shape EventState carries is pinned here.
package record

import (
	"testing"
)

func TestEventStateCarriesStateDTOByValue(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)
	svc.SetArmed(true)
	seen := capture(t)

	lock(svc, 60, -8, 0)

	var states []StateDTO
	for _, e := range *seen {
		if e.name != EventState {
			continue
		}
		// The assertion main.go makes, made here.
		state, ok := e.data.(StateDTO)
		if !ok {
			t.Fatalf("EventState carried %T, but the tray asserts record.StateDTO", e.data)
		}
		states = append(states, state)
	}

	if len(states) == 0 {
		t.Fatal("recording published no EventState")
	}
	if states[len(states)-1].SessionID == "" {
		t.Error("EventState carried no SessionID; the tray badge would stay dark while recording")
	}
}
