package tuner

import (
	"sync"
	"testing"
	"time"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	coreaudio "smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// recorder counts what services/record would see: visibly locked analyses with fresh slot values.
type recorder struct {
	mu     sync.Mutex
	locked int
}

func (r *recorder) OnMeasurement(m dsp.Measurement) {
	if m.Frame == dsp.FrameMeasured && m.Lock.Visible && len(m.Reeds) > 0 {
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

// The settings page's engine switches reach the engine config: none of them is a screen rule any more.
func TestEngineSettingsComeFromTheConfig(t *testing.T) {
	initConfig(t)
	cfg := *appconfig.Get()
	cfg.Tuner.StopAfterLock = true
	cfg.Tuner.ContinuousUpdateManual = true
	cfg.Tuner.BeatUnit = "reference"
	cfg.Tuner.ErrorReference = "goal"
	cfg.Tuner.ScaleNaming = "polish"
	cfg.Tuner.Detection = "steady"
	if err := appconfig.SetConfig(cfg); err != nil {
		t.Fatal(err)
	}

	c := New(nil, nil, nil).snapshot()
	if c.Detection != dsp.DetectionSteady {
		t.Fatalf("engine detection %q, want the config's steady", c.Detection)
	}
	if !c.StopAfterLock || !c.Continuous || c.BeatsInHertz || !c.TargetRelative {
		t.Fatalf("engine = stop %v / continuous %v / hertz %v / relative %v, want the config's true / true / false / true",
			c.StopAfterLock, c.Continuous, c.BeatsInHertz, c.TargetRelative)
	}
	if c.Scale != tuning.ParseNaming("polish") {
		t.Fatalf("engine scale %v, want the config's polish", c.Scale)
	}

	cfg.Tuner.StopAfterLock, cfg.Tuner.BeatUnit, cfg.Tuner.ErrorReference = false, "hz", "scale"
	if err := appconfig.SetConfig(cfg); err != nil {
		t.Fatal(err)
	}
	if c := New(nil, nil, nil).snapshot(); c.StopAfterLock || !c.BeatsInHertz || c.TargetRelative {
		t.Fatalf("engine = stop %v / hertz %v / relative %v, want the defaults false / true / false",
			c.StopAfterLock, c.BeatsInHertz, c.TargetRelative)
	}
}

// A switch flipped on the settings page under a running engine reaches it without a restart.
func TestARunningEngineAdoptsTheSettingsPage(t *testing.T) {
	initConfig(t)
	s := fixtureService(t)
	captureEvents(t, s)
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	cfg := *appconfig.Get()
	cfg.Tuner.StopAfterLock = true
	if err := appconfig.SetConfig(cfg); err != nil {
		t.Fatal(err)
	}
	ok := waitFor(6*time.Second, func() bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		return s.file.stopAfterLock && s.cfg.StopAfterLock
	})
	if !ok {
		t.Fatal("the running engine never adopted stop-after-lock from the config")
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
	g := s.adopt()
	if g.tol != 5.0 || g.beatTol != 0.5 {
		t.Fatalf("goal windows = %v / %v, want the config's 5 / 0.5", g.tol, g.beatTol)
	}
	if st := s.Settings(); st.Tolerance != 5.0 || st.BeatTolerance != 0.5 {
		t.Fatalf("settings = %+v, want the config's windows", st)
	}

	// A reed four units out is in tune at a five unit window and nowhere near it at the one unit default.
	dto := decorate(measured(69, 4), g)
	if len(dto.ReedErrors) != 1 || !dto.ReedErrors[0].InTol {
		t.Fatalf("reed rows = %+v, want one, in tolerance at a five unit window", dto.ReedErrors)
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

// A reference tone is our own sine in the mic; it must never reach a take, which reaches a printed report.
func TestAToneNeverReachesTheRecorder(t *testing.T) {
	initConfig(t)

	s := fixtureService(t)
	rec := &recorder{}
	s.record = rec

	s.toneUntil.Store(time.Now().Add(toneDuration).UnixNano())

	for range 20 {
		s.observe(measured(69, 0))
	}
	if got := rec.count(); got != 0 {
		t.Fatalf("the recorder took %d readings while the app was playing a tone, want 0", got)
	}

	if err := s.StopTone(); err != nil {
		t.Fatal(err)
	}
	// The stop fade still sounds for half a second.
	s.observe(measured(69, 0))
	if got := rec.count(); got != 0 {
		t.Fatalf("the recorder took %d readings during the stop fade, want 0", got)
	}

	time.Sleep(coreaudio.ToneFade + 50*time.Millisecond)
	s.observe(measured(69, 0))
	if got := rec.count(); got != 1 {
		t.Fatalf("the recorder took %d readings after the tone stopped, want 1", got)
	}
}

// StopTone is called on pointer-up, pointer-cancel and window blur, so it must be safe with no tone playing.
func TestStopToneIsIdempotent(t *testing.T) {
	initConfig(t)
	s := fixtureService(t)

	for range 3 {
		if err := s.StopTone(); err != nil {
			t.Fatalf("StopTone with nothing playing: %v", err)
		}
	}
	if s.tonePlaying() {
		t.Fatal("tonePlaying is true after StopTone")
	}
}
