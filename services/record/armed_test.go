package record

import (
	"testing"
)

func TestASessionOpensInWarmUp(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)

	if svc.Armed() {
		t.Fatal("a session opened armed; the notes played to warm up would be filed as readings")
	}
}

func TestADisarmedEngineRecordsNothing(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)

	lock(svc, 60, -8, 0)

	d, err := sessions.Data()
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Takes) != 0 {
		t.Fatalf("recorded %d takes while disarmed, want 0", len(d.Takes))
	}
}

func TestADisarmedEngineIsSilent(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)

	seen := capture(t)
	lock(svc, 60, -8, 0)

	if len(*seen) != 0 {
		t.Fatalf("published %d events while disarmed, want none", len(*seen))
	}
}

func TestArmingRecordsFromTheNextNote(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)

	lock(svc, 60, -8, 0) // warming up, must not be kept
	svc.SetArmed(true)
	lock(svc, 62, -3, 1) // the first real reading

	d, err := sessions.Data()
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Takes) != 1 {
		t.Fatalf("kept %d takes, want only the one played after arming", len(d.Takes))
	}
	if d.Takes[0].Note != 62 {
		t.Fatalf("kept note %d, want 62: the warm-up note was filed", d.Takes[0].Note)
	}
}

func TestDisarmingStopsRecordingAndKeepsWhatLanded(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)

	svc.SetArmed(true)
	lock(svc, 60, -8, 0)
	svc.SetArmed(false)
	lock(svc, 62, -3, 1)

	d, err := sessions.Data()
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Takes) != 1 {
		t.Fatalf("hold %d takes, want the one recorded while armed", len(d.Takes))
	}
	if d.Takes[0].Note != 60 {
		t.Fatalf("kept note %d, want 60", d.Takes[0].Note)
	}
}

func TestStateCarriesTheArmedFlag(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)

	if svc.State().Armed {
		t.Error("State says armed on a session in warm-up")
	}
	svc.SetArmed(true)
	if !svc.State().Armed {
		t.Error("State says disarmed after arming")
	}
}

func TestSetArmedPublishesTheState(t *testing.T) {
	svc, sessions := services(t)
	open(t, sessions, 2)

	seen := capture(t)
	svc.SetArmed(true)

	var states []StateDTO
	for _, e := range *seen {
		if e.name == EventState {
			states = append(states, e.data.(StateDTO))
		}
	}
	if len(states) != 1 {
		t.Fatalf("published %d EventState, want exactly one", len(states))
	}
	if !states[0].Armed {
		t.Error("published a state that does not say it is armed")
	}
	for _, e := range *seen {
		if e.name == EventTable {
			t.Fatal("SetArmed shipped a TableDTO; arming changes no reading")
		}
	}
}
