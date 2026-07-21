package tuner

import (
	appconfig "smegg.me/smeggtuner/common/config"
	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
)

// imposed is what the active session imposes on the engine: reference pitch and reed count; the zero value means no session, and reeds is the instrument's count (not what the engine resolves) so it compares equal across ticks.
type imposed struct {
	a4    float64
	reeds int
	// tol and beatTol are the instrument's own judging windows in cents, or zero; zero falls back to the app default in adopt.
	tol     float64
	beatTol float64
}

// goal is the target a measurement is read against: curve, reference, and reed/beat windows. A nil curve is the empty curve.
type goal struct {
	curve   *target.Curve
	a4      float64
	tol     float64
	beatTol float64
}

// session returns what the active session imposes and the goal curve in force; no session answers the zero pair and empty curve, and the curve is swapped rather than edited, so reading it needs no lock.
func (s *Service) session() (imposed, *target.Curve) {
	if s.sessions == nil {
		return imposed{}, nil
	}
	g := s.sessions.Goal()
	if g.A4 <= 0 {
		return imposed{}, nil // no session: the empty curve, and the app's own A4
	}
	return imposed{a4: g.A4, reeds: g.Reeds, tol: g.Tolerance, beatTol: g.BeatTolerance}, g.Curve
}

// adopt makes a running engine agree with the active session and returns the goal and rules; it acts only when the session changed, so the tuner's own reed-count control wins in between, and the config is re-read here (not cached at Start) so a tightened tolerance takes effect without a restart.
func (s *Service) adopt() (goal, rules) {
	p, curve := s.session()
	tuner := appconfig.Get().Tuner

	// The instrument's own windows where it has them, the app default otherwise; resolved then clamped once.
	rawTol, rawBeat := tuner.Tolerance, tuner.BeatTolerance
	if p.tol > 0 {
		rawTol = p.tol
	}
	if p.beatTol > 0 {
		rawBeat = p.beatTol
	}
	tol, beatTol := target.Tolerances(rawTol, rawBeat)

	s.mu.Lock()
	changed := s.imposed != p
	s.imposed = p
	cfg := impose(s.cfg, p)
	manual := s.cfg.ManualNote != tuning.Note(autoNote)
	s.mu.Unlock()

	if changed {
		if r := s.current(); r != nil {
			r.engine.Update(func(c *dsp.EngineConfig) {
				c.A4 = cfg.A4
				c.ReedCount = cfg.ReedCount
			})
		}
		if p.reeds > maxReeds {
			// The engine is clamped to what it can resolve; the instrument keeps its reeds and its curve its width.
			logger.Info(logger.MsgSessionReedsClamped,
				logger.Int("instrument", p.reeds), logger.Int("engine", cfg.ReedCount))
		}
		emitEvent(EventSettings, settingsDTO(cfg, p))
	}
	return goal{curve: curve, a4: cfg.A4, tol: tol, beatTol: beatTol}, rules{
		stopAfterLock:    tuner.StopAfterLock,
		continuousManual: tuner.ContinuousUpdateManual,
		manual:           manual,
	}
}

// observe hands the measurement to the recorder, unless a reference tone is playing: the mic would hear our own sine as a perfectly in-tune reed and taint a recorded take.
func (s *Service) observe(m dsp.Measurement) {
	if s.record == nil {
		return
	}
	if s.tonePlaying() {
		return
	}
	s.record.OnMeasurement(m)
}

// decorate joins a measurement to the goal; a heartbeat (no reeds) travels as is. With no curve every Goal is zero and every Error is the plain deviation.
func decorate(m dsp.Measurement, g goal) MeasurementDTO {
	dto := MeasurementDTO{Measurement: m}

	// The pictures are packed to bytes here, at the wire; the Measurement keeps its floats for the golden tests.
	dto.Spectrum = encode(pack(m.Spectrum, 0, 1))
	dto.Equalizer = encode(pack(m.Equalizer, 0, dsp.EqualizerCeilingDB))
	dto.Waveform = encode(packSigned(m.Waveform))

	if len(m.Reeds) == 0 {
		return dto
	}
	dto.ReedErrors = target.Errors(m, g.curve, g.a4, g.tol)
	dto.BeatErrors = target.BeatErrors(m, g.curve, g.a4, g.beatTol)
	return dto
}

// impose folds the active session over an engine config: its reference pitch becomes the engine's and its instrument sets the reed count, up to what the tones mode can split.
func impose(c dsp.EngineConfig, p imposed) dsp.EngineConfig {
	if p.a4 > 0 {
		c.A4 = p.a4
	}
	if p.reeds > 0 {
		c.ReedCount = clampReeds(p.reeds)
	}
	return c
}
