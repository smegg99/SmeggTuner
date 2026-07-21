package tuner

import (
	"time"

	appconfig "smegg.me/smeggtuner/common/config"
	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

const (
	// Mic calibrates on its first seconds; a file starts mid-note and gets none.
	micCalibSecs  = 5.0
	fileCalibSecs = 0.0

	// Low-cut always on: removes footstep and bellows thump, below any reed.
	lowCutAlways = true

	// Reference tone length; the config file's tuner.tone_duration_ms overrides it.
	toneDuration = 10 * time.Second

	// emitQueue is the DSP-to-Wails handoff depth: four fine results plus heartbeats.
	emitQueue = 16

	defaultA4    = 440.0
	minA4, maxA4 = 430.0, 450.0
	minReeds     = 1
	// maxReeds is the most a register can sound; the engine never resolves more.
	maxReeds         = 6
	maxTranspose     = 24 // two octaves either way
	autoNote     int = 0  // 0 disables the manual note pin
)

func (s *Service) toneDur() time.Duration {
	if appconfig.GetConfigPath() == "" {
		return toneDuration
	}
	if ms := appconfig.Get().Tuner.ToneDurationMs; ms > 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return toneDuration
}

// apply mutates the stored config and forwards it to a running engine; a stopped engine adopts it at the next Start.
func (s *Service) apply(mut func(*dsp.EngineConfig)) {
	s.mu.Lock()
	mut(&s.cfg)
	s.mu.Unlock()
	if r := s.current(); r != nil {
		r.engine.Update(mut)
	}
}

// merged folds the config file into c; while persistFailed is set, A4/filters/clock stay from the stored config to avoid a silent revert at the next Start.
func (s *Service) merged(c dsp.EngineConfig) dsp.EngineConfig {
	if appconfig.GetConfigPath() == "" {
		return c
	}
	fc := appconfig.Get()
	if !s.persistFailed.Load() {
		c.A4 = fc.Tuner.A4
		c.Hum50 = fc.Audio.HumFilter50
		c.Hum60 = fc.Audio.HumFilter60
	}
	// ReedCount is not a config setting: it is the open session's bank count.
	c.Scale = tuning.ParseNaming(fc.Tuner.ScaleNaming)
	c.ClockPPM = fc.Audio.ClockPPM
	c.Highpass = lowCutAlways

	// App-to-engine timing boundary: core/dsp knows nothing of common/config.
	c.FineWindow = time.Duration(fc.Engine.FineWindowMs) * time.Millisecond
	c.LockHold = time.Duration(fc.Engine.LockHoldMs) * time.Millisecond
	c.LockEpsilonHz = fc.Engine.LockEpsilonHz
	return c
}

// snapshot returns the config the engine would run with now, session included. Read-only: PlayTone must not move service state.
func (s *Service) snapshot() dsp.EngineConfig {
	p, _ := s.session()
	s.mu.Lock()
	defer s.mu.Unlock()
	return impose(s.merged(s.cfg), p)
}

// refresh folds the config file into the stored config and returns what the engine runs on; the session's own values never enter the stored config that persist writes back.
func (s *Service) refresh() dsp.EngineConfig {
	p, _ := s.session()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = s.merged(s.cfg)
	s.imposed = p
	return impose(s.cfg, p)
}

// persist writes a setting back to the config file; a failed write is logged, not returned, and writes are serialized because Wails runs SetA4 and SetFilters concurrently.
func (s *Service) persist(mut func(*appconfig.Config)) {
	if appconfig.GetConfigPath() == "" {
		return // config was never initialized (unit tests, headless tooling)
	}
	s.persistMu.Lock()
	defer s.persistMu.Unlock()

	c := *appconfig.Get()
	if s.persistFailed.Load() {
		// An earlier write failed, so the file is behind the engine; re-assert the fields this service owns.
		s.mu.Lock()
		cur := s.cfg
		s.mu.Unlock()
		c.Tuner.A4 = cur.A4
		c.Audio.HumFilter50 = cur.Hum50
		c.Audio.HumFilter60 = cur.Hum60
	}
	mut(&c)

	if err := appconfig.SetConfig(c); err != nil {
		s.persistFailed.Store(true)
		logger.Warn(logger.MsgConfigSaveFailed, logger.Err(err))
		return
	}
	s.persistFailed.Store(false)
}
