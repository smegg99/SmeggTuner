package tuner

import (
	"sync/atomic"
	"testing"
	"time"

	coreaudio "smegg.me/smeggtuner/core/audio"
	audiosvc "smegg.me/smeggtuner/services/audio"
)

// The event contract the frontend is built against: two event names, both cadences, a state event on either side.
func TestEventContract(t *testing.T) {
	if EventMeasurement != "tuner:measurement" || EventState != "tuner:state" {
		t.Fatalf("event names are frontend API: got %q / %q", EventMeasurement, EventState)
	}

	s := fixtureService(t)
	log := captureEvents(t, s)

	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	// The hook is installed after Start on purpose: a hook set later must still reach the running run's emitter.
	var ticks atomic.Int64
	s.setEmitHookForTest(func() { ticks.Add(1) })

	ok := waitFor(8*time.Second, func() bool {
		heartbeats, fine := log.cadences()
		return heartbeats > 0 && fine > 0 && ticks.Load() > 0
	})
	if !ok {
		heartbeats, fine := log.cadences()
		t.Fatalf("both cadences must reach the frontend: %d heartbeats, %d fine results, %d hook ticks",
			heartbeats, fine, ticks.Load())
	}

	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}

	events := log.all()
	if len(events) < 3 {
		t.Fatalf("got %d events, want a state event, a measurement stream and a state event", len(events))
	}

	first := events[0]
	if first.name != EventState || !first.state.Running {
		t.Fatalf("first event = %q %+v, want %q with Running:true", first.name, first.state, EventState)
	}
	if first.state.Source != fixtureName {
		t.Fatalf("first event Source = %q, want %q", first.state.Source, fixtureName)
	}

	last := events[len(events)-1]
	if last.name != EventState || last.state.Running {
		t.Fatalf("last event = %q %+v, want %q with Running:false", last.name, last.state, EventState)
	}
	if last.state.Error != "" {
		t.Fatalf("a clean Stop must carry no error key, got %q", last.state.Error)
	}
	if last.state.Source != fixtureName {
		t.Fatalf("last event Source = %q, want %q", last.state.Source, fixtureName)
	}

	// Everything between the two state events is the run's stream: measurements and - because the fixture is a file - the playhead.
	for i, e := range events[1 : len(events)-1] {
		if e.name != EventMeasurement && e.name != EventPlayback {
			t.Fatalf("event %d = %q, want %q or %q", i+1, e.name, EventMeasurement, EventPlayback)
		}
	}
}

// The playhead is its own event on its own clock; it fires only for a source that has one and must not outlive the engine.
func TestThePlayheadBelongsToTheRun(t *testing.T) {
	s := fixtureService(t) // a file
	log := captureEvents(t, s)

	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	seen := func() bool {
		for _, e := range log.all() {
			if e.name == EventPlayback {
				return true
			}
		}
		return false
	}
	if !waitFor(8*time.Second, seen) {
		t.Fatal("a file emitted no playhead; the file view would have no needle")
	}
	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}

	events := log.all()
	last := events[len(events)-1]
	if last.name != EventPlayback {
		return // the common case: the run's last word is its state event
	}
	t.Fatalf("a playhead event followed the terminal state event: the needle outlived the engine")
}

// A microphone has no playhead; this is the type check start() makes before launching followPlayhead.
func TestAMicrophoneHasNoPlayhead(t *testing.T) {
	mic, err := coreaudio.NewMicSource("")
	if err != nil {
		t.Fatal(err)
	}

	var src coreaudio.Source = mic
	if _, ok := src.(coreaudio.Transport); ok {
		t.Fatal("a microphone claims a transport: the file view would offer to seek the room")
	}
}

// Regression guard: a run that ends on its own must finish emitting before the next announces itself; finish once cleared s.run before emitting, latching "stopped" over a live engine forever.
func TestSpontaneousEndOrdersBeforeNextStart(t *testing.T) {
	as := audiosvc.New()
	// No loop: the run ends at EOF, on its own, at a moment no caller controls.
	if err := as.SelectFile(shortWav(t), false); err != nil {
		t.Fatal(err)
	}
	s := New(as, nil, nil)
	log := captureEvents(t, s)

	const runs = 4
	for i := 0; i < runs; i++ {
		if err := s.Start(); err != nil {
			t.Fatalf("Start %d: %v", i, err)
		}
		if !s.IsRunning() {
			t.Fatalf("Start %d: not running", i)
		}
		// The next Start fires the instant this run stops counting as running, which is the window the bug lived in.
		if !spinUntil(10*time.Second, func() bool { return !s.IsRunning() }) {
			t.Fatalf("run %d never reached EOF", i)
		}
	}
	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}
	if s.IsRunning() {
		t.Fatal("still running after the last run ended")
	}

	events := log.all()
	running := false
	announced, streamed := 0, 0
	for i, e := range events {
		switch e.name {
		case EventState:
			switch {
			case e.state.Running && running:
				t.Fatalf("event %d: Running:true while the frontend already believes it is running - "+
					"a previous run's terminal event was reordered behind it", i)
			case e.state.Running:
				running = true
				announced++
				streamed = 0
			case !e.state.Running && !running:
				t.Fatalf("event %d: a stale Running:false landed after its run was replaced", i)
			default:
				if streamed == 0 {
					t.Fatalf("run %d was announced but never streamed a measurement", announced)
				}
				running = false
			}
		case EventMeasurement:
			if !running {
				t.Fatalf("event %d: a measurement reached the frontend while it believes the engine is stopped", i)
			}
			streamed++
		case EventPlayback:
			// The needle is held to the same rule: nothing from a run may reach the frontend after it told the UI the engine is down.
			if !running {
				t.Fatalf("event %d: a playhead reached the frontend while it believes the engine is stopped", i)
			}
		default:
			t.Fatalf("event %d: unknown event %q", i, e.name)
		}
	}
	if running {
		t.Fatal("the last run never reported that it stopped")
	}
	if announced != runs {
		t.Fatalf("got %d Running:true events, want %d: a Start that lands on a retiring run must build a new engine", announced, runs)
	}
}
