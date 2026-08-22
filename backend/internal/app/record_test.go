package app

import (
	"slices"
	"testing"

	recordservice "smegg.me/smeggtuner/backend/services/record"
)

type fakeEngine struct {
	learned   int
	recording []bool
}

func (f *fakeEngine) Recalibrate()         { f.learned++ }
func (f *fakeEngine) SetRecording(on bool) { f.recording = append(f.recording, on) }

// Arming relearns the room once, and takes are written only while armed with a session open.
func TestRecordStateDrivesTheEngine(t *testing.T) {
	var st recordservice.StateDTO
	engine := &fakeEngine{}
	link := &recordLink{state: func() recordservice.StateDTO { return st }, engine: engine}

	st = recordservice.StateDTO{Armed: true}
	link.follow()
	st = recordservice.StateDTO{Armed: true, SessionID: "s1"}
	link.follow()
	link.follow()
	st = recordservice.StateDTO{SessionID: "s1"}
	link.follow()
	st = recordservice.StateDTO{Armed: true, SessionID: "s1"}
	link.follow()

	if engine.learned != 2 {
		t.Fatalf("relearned %d times, want once per arming (2)", engine.learned)
	}
	if want := []bool{false, true, true, false, true}; !slices.Equal(engine.recording, want) {
		t.Fatalf("recording = %v, want %v", engine.recording, want)
	}
}
