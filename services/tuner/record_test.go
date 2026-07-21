package tuner

import (
	"sync"
	"testing"
	"time"

	appconfig "smegg.me/smeggtuner/common/config"
	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/target"
)

// recorder counts what services/record sees: the engine's stream, not the screen's filtered one.
type recorder struct {
	mu     sync.Mutex
	locked int
}

func (r *recorder) OnMeasurement(m dsp.Measurement) {
	if m.Locked && len(m.Reeds) > 0 {
		r.mu.Lock()
		r.locked++
		r.mu.Unlock()
	}
}

func (r *recorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.locked
}

// With the hold on, the frontend never sees two locked readings running; the recorder still sees every one.
func TestStopAfterLockHoldsTheScreenAndNotTheRecorder(t *testing.T) {
	initConfig(t)
	cfg := *appconfig.Get()
	cfg.Tuner.StopAfterLock = true
	if err := appconfig.SetConfig(cfg); err != nil {
		t.Fatal(err)
	}

	s := fixtureService(t)
	rec := &recorder{}
	s.record = rec
	log := captureEvents(t, s)

	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	if !waitFor(8*time.Second, func() bool { return rec.count() >= 2 }) {
		t.Fatalf("the recorder saw %d locked readings, want at least 2", rec.count())
	}
	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}

	// Every locked reading the engine produced reached the recorder; the frontend got a subset: no two locked in a row.
	var prevLocked bool
	var shown int
	for _, e := range log.all() {
		if e.name != EventMeasurement || len(e.meas.Reeds) == 0 {
			continue
		}
		shown++
		if e.meas.Locked && prevLocked {
			t.Fatal("two locked readings in a row reached the screen: the hold did not hold")
		}
		prevLocked = e.meas.Locked
	}
	if shown == 0 {
		t.Fatal("the screen was never given a reading at all")
	}
	if rec.count() <= 0 {
		t.Fatal("the recorder must see the engine's stream, hold or no hold")
	}
}

// The tolerance windows come from the config, not core/target's hardcoded defaults.
func TestTolerancesComeFromTheConfig(t *testing.T) {
	initConfig(t)
	cfg := *appconfig.Get()
	cfg.Tuner.Tolerance = 5.0
	cfg.Tuner.BeatTolerance = 0.5
	if err := appconfig.SetConfig(cfg); err != nil {
		t.Fatal(err)
	}

	s := New(nil, nil, nil)
	g, _ := s.adopt()
	if g.tol != 5.0 || g.beatTol != 0.5 {
		t.Fatalf("goal windows = %v / %v, want the config's 5 / 0.5", g.tol, g.beatTol)
	}
	if st := s.Settings(); st.Tolerance != 5.0 || st.BeatTolerance != 0.5 {
		t.Fatalf("settings = %+v, want the config's windows", st)
	}

	// A reed four cents out is in tune at a five cent window and nowhere near it at the one cent default.
	m := dsp.Measurement{Note: 69, ScalePitch: 440, Reeds: []dsp.ReedMeasure{{Freq: 441.02, DevCents: 4.0}}}
	dto := decorate(m, g)
	if len(dto.ReedErrors) != 1 || !dto.ReedErrors[0].InTol {
		t.Fatalf("reed rows = %+v, want one, in tolerance at a five cent window", dto.ReedErrors)
	}
}

// Unset windows fall back to core/target's defaults, not the zeros an empty struct holds.
func TestUnsetTolerancesFallBackToTheDefaults(t *testing.T) {
	var zero appconfig.Tuner
	tol, beatTol := target.Tolerances(zero.Tolerance, zero.BeatTolerance)
	if tol != target.DefaultTolerance || beatTol != target.DefaultBeatTolerance {
		t.Fatalf("unset windows = %v / %v, want core/target's defaults", tol, beatTol)
	}
}

// The manual flag the rules are read with is the engine's pinned note, not a second copy of it.
func TestRulesFollowTheManualNote(t *testing.T) {
	initConfig(t)
	s := New(nil, nil, nil)

	if _, r := s.adopt(); r.manual {
		t.Fatal("nothing is pinned: the detector is tracking")
	}
	if err := s.SetManualNote(69); err != nil {
		t.Fatal(err)
	}
	if _, r := s.adopt(); !r.manual {
		t.Fatal("a pinned note is manual mode")
	}
	if err := s.SetManualNote(autoNote); err != nil {
		t.Fatal(err)
	}
	if _, r := s.adopt(); r.manual {
		t.Fatal("auto hands tracking back to the detector")
	}
}

// A reference tone is our own sine in the mic; it must never reach a take, which reaches a printed report.
func TestAToneNeverReachesTheRecorder(t *testing.T) {
	initConfig(t)

	s := fixtureService(t)
	rec := &recorder{}
	s.record = rec

	s.toneUntil.Store(time.Now().Add(toneDuration).UnixNano())

	for i := 0; i < 20; i++ {
		s.observe(reading(true))
	}
	if got := rec.count(); got != 0 {
		t.Fatalf("the recorder took %d readings while the app was playing a tone, want 0", got)
	}

	if err := s.StopTone(); err != nil {
		t.Fatal(err)
	}

	s.observe(reading(true))
	if got := rec.count(); got != 1 {
		t.Fatalf("the recorder took %d readings after the tone stopped, want 1", got)
	}
}

// StopTone is called on pointer-up, pointer-cancel and window blur, so it must be safe with no tone playing.
func TestStopToneIsIdempotent(t *testing.T) {
	initConfig(t)
	s := fixtureService(t)

	for i := 0; i < 3; i++ {
		if err := s.StopTone(); err != nil {
			t.Fatalf("StopTone with nothing playing: %v", err)
		}
	}
	if s.tonePlaying() {
		t.Fatal("tonePlaying is true after StopTone")
	}
}
