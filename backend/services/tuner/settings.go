package tuner

import (
	"math"

	"smegg.me/smeggtuner/backend/common/i18n"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/target"
)

// SetA4 sets the reference pitch in hertz, 430..450 on a 0.1 Hz grid. Refused while tuner.lock_a4 is
// on; with a session open it routes into the session, which refuses it once readings exist.
func (s *Service) SetA4(hz float64) error {
	if appconfig.GetConfigPath() != "" && appconfig.Get().Tuner.LockA4 {
		return rejected("a4", hz, ErrA4Locked)
	}
	hz = math.Round(hz*10) / 10
	if !(hz >= minA4 && hz <= maxA4) {
		return rejected("a4", hz, ErrInvalidA4)
	}
	if p := s.session(); p.a4 > 0 {
		return s.sessions.SetA4(hz)
	}
	return s.set("a4", func(c *dsp.EngineConfig, _ *panel, _ bool) error {
		c.A4 = hz
		return nil
	}, func(fc *appconfig.Config, c dsp.EngineConfig, _ panel) { fc.Tuner.A4 = c.A4 })
}

// SetTranspose sets how many semitones (0..11) the instrument's keys sound above their names. With a
// session open it routes into the session, which refuses it once readings exist; otherwise tones mode only.
func (s *Service) SetTranspose(semitones int) error {
	if semitones < 0 || semitones > maxTranspose {
		return rejected("transpose", semitones, ErrInvalidTranspose)
	}
	if p := s.session(); p.a4 > 0 {
		return s.sessions.SetTranspose(semitones)
	}
	return s.set("transpose", func(c *dsp.EngineConfig, pn *panel, open bool) error {
		if !open && pn.mode != ModeTones {
			return ErrTransposeMode
		}
		c.Transpose = semitones
		return nil
	}, func(fc *appconfig.Config, c dsp.EngineConfig, _ panel) { fc.Tuner.Transpose = c.Transpose })
}

// Settings is what the engine is measuring with right now. The UI reads it once on mount and follows EventSettings.
func (s *Service) Settings() SettingsDTO {
	return settingsDTO(s.view())
}

// SetFilters toggles the 50 Hz and 60 Hz mains notches.
func (s *Service) SetFilters(hum50, hum60 bool) error {
	return s.set("filters", func(c *dsp.EngineConfig, _ *panel, _ bool) error {
		c.Hum50, c.Hum60 = hum50, hum60
		return nil
	}, func(fc *appconfig.Config, c dsp.EngineConfig, _ panel) {
		fc.Audio.HumFilter50, fc.Audio.HumFilter60 = c.Hum50, c.Hum60
	})
}

// Recalibrate starts the engine's background learning on the current run; a no-op when nothing is running. Called when recording arms.
func (s *Service) Recalibrate() {
	if r := s.current(); r != nil {
		r.engine.Recalibrate()
	}
}

func rejected(setting string, value any, err *ServiceError) error {
	logger.Warn(i18n.KeyLogTunerSettingRejected, logger.Str("setting", setting), logger.Any("value", value))
	return err
}

func clampReeds(n int) int {
	return min(maxReeds, max(minReeds, n))
}

// settingsDTO fills the DTO with the engine's config, the panel and the tolerance windows (from config, or core/target's defaults).
func settingsDTO(cfg dsp.EngineConfig, pn panel, p imposed) SettingsDTO {
	tuner := appconfig.Get().Tuner
	tol, beatTol := target.Tolerances(tuner.Tolerance, tuner.BeatTolerance)
	return SettingsDTO{
		A4:            cfg.A4,
		ReedCount:     slotCount(cfg),
		SessionReeds:  p.reeds,
		Tolerance:     tol,
		BeatTolerance: beatTol,
		Mode:          string(pn.mode),
		ReedBoxes:     pn.reeds.ids(),
		OctaveOffsets: pn.offsets(),
		Transpose:     cfg.Transpose,
		ManualNote:    int(cfg.ManualNote),
	}
}
