package tuner

import (
	"reflect"
	"time"

	"smegg.me/smeggtuner/backend/common/i18n"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
)

const (
	// toneDuration is the reference tone's audible length; the config file's tuner.tone_duration_ms overrides it.
	toneDuration = 43 * time.Second

	// emitQueue is the DSP-to-Wails handoff depth: two analyses a second plus twenty heartbeats.
	emitQueue = 16

	defaultA4    = 440.0
	minA4, maxA4 = 430.0, 450.0
	minReeds     = 1
	// maxReeds is the most a register can sound; the engine never resolves more.
	maxReeds     = 6
	maxTranspose = 11 // the instrument's C sounds as one of the twelve keys C..B
	autoNote     = 0  // 0 disables the manual note pin
)

// fileSettings is what the engine takes from the config file alone; comparable, so adopt can tell
// when the settings page changed one under a running engine.
type fileSettings struct {
	scale         tuning.ScaleNaming
	detection     dsp.Detection
	stopAfterLock bool
	continuous    bool
	hertz         bool
	relative      bool
}

func readFileSettings(fc *appconfig.Config) fileSettings {
	return fileSettings{
		scale:         tuning.ParseNaming(fc.Tuner.ScaleNaming),
		detection:     dsp.Detection(fc.Tuner.Detection),
		stopAfterLock: fc.Tuner.StopAfterLock,
		continuous:    fc.Tuner.ContinuousUpdateManual,
		hertz:         fc.Tuner.BeatUnit == "hz",
		relative:      fc.Tuner.ErrorReference == string(target.RefGoal),
	}
}

func (f fileSettings) apply(c *dsp.EngineConfig) {
	c.Scale = f.scale
	c.Detection = f.detection
	c.StopAfterLock = f.stopAfterLock
	c.Continuous = f.continuous
	c.BeatsInHertz = f.hertz
	c.TargetRelative = f.relative
}

// fileSettingsOf is the config file's engine settings, or the ones c already carries when no config is loaded.
func fileSettingsOf(c dsp.EngineConfig) fileSettings {
	if appconfig.GetConfigPath() == "" {
		return fileSettings{scale: c.Scale, detection: c.Detection, stopAfterLock: c.StopAfterLock,
			continuous: c.Continuous, hertz: c.BeatsInHertz, relative: c.TargetRelative}
	}
	return readFileSettings(appconfig.Get())
}

func (s *Service) toneDur() time.Duration {
	if appconfig.GetConfigPath() == "" {
		return toneDuration
	}
	if ms := appconfig.Get().Tuner.ToneDurationMs; ms > 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return toneDuration
}

// set changes the tuner's own settings under persistMu, so memory and the config file agree between
// setters. mut sees them with the file folded in and may refuse; a running engine gets the result,
// keys writes the changed values back, and the UI hears EventSettings.
func (s *Service) set(setting string, mut func(c *dsp.EngineConfig, pn *panel, open bool) error,
	keys func(fc *appconfig.Config, c dsp.EngineConfig, pn panel)) error {
	p := s.session()
	s.persistMu.Lock()
	defer s.persistMu.Unlock()

	s.mu.Lock()
	c, pn := s.merged(s.cfg), s.mergedPanel(s.panel)
	before := impose(c, pn, p)
	if err := mut(&c, &pn, p.a4 > 0); err != nil {
		s.mu.Unlock()
		logger.Warn(i18n.KeyLogTunerSettingRejected, logger.Str("setting", setting), logger.Err(err))
		return err
	}
	s.cfg, s.panel = c, pn
	if after := impose(c, pn, p); !reflect.DeepEqual(before, after) {
		if r := s.liveLocked(); r != nil {
			r.reconfigure(after)
		}
	}
	s.mu.Unlock()

	if keys != nil {
		s.write(func(fc *appconfig.Config) { keys(fc, c, pn) })
	}
	s.announce()
	return nil
}

// merged folds the config file into c; while persistFailed is set, the keys this service writes stay from memory to avoid a silent revert.
func (s *Service) merged(c dsp.EngineConfig) dsp.EngineConfig {
	c.Calibration = s.correction()
	if appconfig.GetConfigPath() == "" {
		return c
	}
	fc := appconfig.Get()
	if !s.persistFailed.Load() {
		c.A4 = fc.Tuner.A4
		c.Transpose = fc.Tuner.Transpose
		c.Hum50 = fc.Audio.HumFilter50
		c.Hum60 = fc.Audio.HumFilter60
	}
	// ReedCount is not a config setting: the session's register or the panel decides it.
	readFileSettings(fc).apply(&c)
	return c
}

// mergedPanel is the config file's panel, or pn when none is loaded or a failed write left the file behind.
func (s *Service) mergedPanel(pn panel) panel {
	if appconfig.GetConfigPath() == "" || s.persistFailed.Load() {
		return pn
	}
	return readPanel(appconfig.Get().Tuner)
}

// view is the config the engine would run with now, the panel and the session. Read-only: PlayTone must not move service state.
func (s *Service) view() (dsp.EngineConfig, panel, imposed) {
	p := s.session()
	s.mu.Lock()
	defer s.mu.Unlock()
	pn := s.mergedPanel(s.panel)
	return impose(s.merged(s.cfg), pn, p), pn, p
}

func (s *Service) snapshot() dsp.EngineConfig {
	c, _, _ := s.view()
	return c
}

// refreshLocked folds the config file into the stored config and returns what the engine runs on; the session's own values never enter the stored config. Callers hold persistMu and mu.
func (s *Service) refreshLocked(p imposed) dsp.EngineConfig {
	s.cfg, s.panel = s.merged(s.cfg), s.mergedPanel(s.panel)
	s.imposed, s.file = p, fileSettingsOf(s.cfg)
	return impose(s.cfg, s.panel, p)
}

func (s *Service) refresh() dsp.EngineConfig {
	p := s.session()
	s.persistMu.Lock()
	defer s.persistMu.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.refreshLocked(p)
}

// write puts settings back into the config file; callers hold persistMu. A failed write is logged, not
// returned, and until one succeeds every key this service owns is re-asserted from memory.
func (s *Service) write(mut func(*appconfig.Config)) {
	if appconfig.GetConfigPath() == "" {
		return // config was never initialized (unit tests, headless tooling)
	}
	err := appconfig.Update(func(c *appconfig.Config) {
		if s.persistFailed.Load() {
			s.mu.Lock()
			cur, pn := s.cfg, s.panel
			s.mu.Unlock()
			c.Tuner.A4, c.Tuner.Transpose = cur.A4, cur.Transpose
			c.Audio.HumFilter50, c.Audio.HumFilter60 = cur.Hum50, cur.Hum60
			pn.write(&c.Tuner)
		}
		mut(c)
	})
	if err != nil {
		s.persistFailed.Store(true)
		logger.Warn(i18n.KeyLogConfigSaveFailed, logger.Err(err))
		return
	}
	s.persistFailed.Store(false)
}

// announce tells the UI what the engine measures with now.
func (s *Service) announce() {
	emitEvent(EventSettings, s.Settings())
}
