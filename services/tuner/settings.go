package tuner

import (
	appconfig "smegg.me/smeggtuner/common/config"
	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
)

// SetA4 sets the reference pitch in hertz (430..450). With a session open it routes
// into the session (which persists it and refuses it while a pass is recorded).
func (s *Service) SetA4(hz float64) error {
	if hz < minA4 || hz > maxA4 {
		logger.Warn(logger.MsgTunerSettingRejected,
			logger.Str("setting", "a4"), logger.Any("value", hz))
		return ErrInvalidA4
	}
	if p, _ := s.session(); p.a4 > 0 {
		return s.sessions.SetA4(hz)
	}
	s.apply(func(c *dsp.EngineConfig) { c.A4 = hz })
	s.persist(func(c *appconfig.Config) { c.Tuner.A4 = hz })
	return nil
}

// Settings is what the engine is measuring with right now. The UI reads it once on mount and follows EventSettings.
func (s *Service) Settings() SettingsDTO {
	p, _ := s.session()
	return settingsDTO(s.snapshot(), p)
}

// SetManualNote pins the tracked note; 0 hands tracking back to the detector. Not persisted; the config schema has no field for it.
func (s *Service) SetManualNote(note int) error {
	if note != autoNote && !tuning.Note(note).Valid() {
		logger.Warn(logger.MsgTunerSettingRejected,
			logger.Str("setting", "manual_note"), logger.Int("value", note))
		return ErrInvalidNote
	}
	s.apply(func(c *dsp.EngineConfig) { c.ManualNote = tuning.Note(note) })
	return nil
}

// SetTranspose shifts the target pitch by semitones (-24..24) without moving the tracked note. Session setting.
func (s *Service) SetTranspose(semitones int) error {
	if semitones < -maxTranspose || semitones > maxTranspose {
		logger.Warn(logger.MsgTunerSettingRejected,
			logger.Str("setting", "transpose"), logger.Int("value", semitones))
		return ErrInvalidTranspose
	}
	s.apply(func(c *dsp.EngineConfig) { c.Transpose = semitones })
	return nil
}

// SetFilters toggles the 50 Hz and 60 Hz mains notches.
func (s *Service) SetFilters(hum50, hum60 bool) error {
	s.apply(func(c *dsp.EngineConfig) {
		c.Hum50 = hum50
		c.Hum60 = hum60
	})
	s.persist(func(c *appconfig.Config) {
		c.Audio.HumFilter50 = hum50
		c.Audio.HumFilter60 = hum60
	})
	return nil
}

// Freeze latches the last fine result on screen while the engine keeps running; re-asserted after every Start.
func (s *Service) Freeze(on bool) error {
	s.mu.Lock()
	s.frozen = on
	s.mu.Unlock()
	if r := s.current(); r != nil {
		r.engine.Freeze(on)
	}
	return nil
}

// Recalibrate restarts the engine's noise-floor warm-up on the current run; a no-op when nothing is running. Called when recording arms.
func (s *Service) Recalibrate() {
	if r := s.current(); r != nil {
		r.engine.Recalibrate()
	}
}

func clampReeds(n int) int {
	if n < minReeds {
		return minReeds
	}
	if n > maxReeds {
		return maxReeds
	}
	return n
}

// settingsDTO fills the DTO with the engine's config and the tolerance windows (from config, or core/target's defaults).
func settingsDTO(cfg dsp.EngineConfig, p imposed) SettingsDTO {
	tuner := appconfig.Get().Tuner
	tol, beatTol := target.Tolerances(tuner.Tolerance, tuner.BeatTolerance)
	return SettingsDTO{
		A4:            cfg.A4,
		ReedCount:     cfg.ReedCount,
		SessionReeds:  p.reeds,
		Tolerance:     tol,
		BeatTolerance: beatTol,
	}
}
