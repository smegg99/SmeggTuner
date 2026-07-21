package record

import (
	"testing"
)

func TestADisarmSurvivesEverythingDoneInsideOneSession(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)

	svc.SetArmed(true)
	lock(svc, 60, -8, 0)

	svc.SessionChanged() // what the running app does after a take lands

	if !svc.Armed() {
		t.Fatal("recording a note disarmed the tuner; the technician would file one reading and stop")
	}
}

func TestOpeningAnotherSessionStartsItsOwnWarmUp(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)

	svc.SetArmed(true)
	svc.SessionChanged() // same session, stays armed
	if !svc.Armed() {
		t.Fatal("the same session disarmed itself")
	}

	open(t, sessions, 3) // a second accordion; its own warm-up
	svc.SessionChanged()

	if svc.Armed() {
		t.Fatal("a newly opened session inherited the last one's arming")
	}
}

func TestClosingASessionDisarms(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)
	svc.SetArmed(true)

	if err := sessions.Close(); err != nil {
		t.Fatal(err)
	}
	svc.SessionChanged()

	if svc.Armed() {
		t.Fatal("still armed with no session on the bench")
	}
}

// Seed the identity first, then a routine mutation must ship the state and no table.
func TestSessionChangedPublishesTheState(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)
	svc.SessionChanged() // the session arrives, and is now the known one

	seen := capture(t)
	svc.SessionChanged() // a mutation within the session

	var states int
	for _, e := range *seen {
		if e.name == EventState {
			states++
		}
		if e.name == EventTable {
			t.Fatal("SessionChanged shipped a TableDTO; it fires on every curve drag")
		}
	}
	if states != 1 {
		t.Fatalf("published %d EventState, want exactly one", states)
	}
}

// lastSession is empty on a fresh service, so the first SessionChanged must not read it as a change and drop an arm made before it.
func TestArmingBelongsToTheSessionItWasMadeIn(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)

	svc.SetArmed(true)   // nothing has called SessionChanged yet
	svc.SessionChanged() // the first one this service has ever seen

	if !svc.Armed() {
		t.Fatal("the first session:active threw away an arm made before it")
	}
}

// A session arriving ships a table; every other mutation that lands here must not.
func TestOpeningASessionShipsItsReadings(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)
	svc.SetArmed(true)
	lock(svc, 60, -8, 0)

	open(t, sessions, 3) // a different accordion, carrying readings of its own

	seen := capture(t)
	svc.SessionChanged()

	var tables, states int
	for _, e := range *seen {
		switch e.name {
		case EventTable:
			tables++
		case EventState:
			states++
		}
	}
	if tables != 1 {
		t.Fatalf("a session arriving published %d tables, want exactly one: the screen has no other way to hear its readings", tables)
	}
	if states != 1 {
		t.Fatalf("published %d states, want exactly one", states)
	}
}

func TestAMutationWithinOneSessionShipsNoTable(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)
	svc.SessionChanged() // the session arrives, and is now the known one

	seen := capture(t)
	svc.SessionChanged() // a mutation within the session

	for _, e := range *seen {
		if e.name == EventTable {
			t.Fatal("a mutation within one session shipped a whole table; this fires per drag frame")
		}
	}
}
