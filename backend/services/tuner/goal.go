package tuner

import (
	"strconv"
	"strings"

	"smegg.me/smeggtuner/backend/common/i18n"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/core/dsp"
	coresession "smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/target"
)

// imposed is what the active session imposes on the engine; the zero value means no session. reeds
// is the instrument's count (not what the engine resolves) and the bank and foot lists are joined
// strings, so the struct stays comparable across ticks.
type imposed struct {
	a4        float64
	transpose int
	reeds     int
	banks     string
	// bassFeet is the sounding bass ladder ("32,16,8") while the bench faces the bass side.
	bassFeet string
	// cardBanks and cardFeet are the instrument's own ranks: the goal curve's columns.
	cardBanks string
	cardFeet  string
	// curve is swapped on every edit and never changed in place, so the pointer says when the targets move.
	curve *target.Curve
	// tol and beatTol are the instrument's own judging windows in reference units, or zero for the app default.
	tol     float64
	beatTol float64
}

func joinBanks(banks []coresession.Bank) string {
	names := make([]string, len(banks))
	for i, b := range banks {
		names[i] = string(b)
	}
	return strings.Join(names, ",")
}

func splitBanks(s string) []coresession.Bank {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	banks := make([]coresession.Bank, len(parts))
	for i, p := range parts {
		banks[i] = coresession.Bank(p)
	}
	return banks
}

func joinFeet(feet []int) string {
	parts := make([]string, len(feet))
	for i, f := range feet {
		parts[i] = strconv.Itoa(f)
	}
	return strings.Join(parts, ",")
}

func splitFeet(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	feet := make([]int, 0, len(parts))
	for _, p := range parts {
		if f, err := strconv.Atoi(p); err == nil {
			feet = append(feet, f)
		}
	}
	return feet
}

// columns is the card column each of a register's slots answers to; nil when the mapping is
// uncertain, and core/target falls back to the slots' own positions.
func (p imposed) columns(slots int) []int {
	if feet := splitFeet(p.bassFeet); len(feet) > 0 {
		return coresession.FootColumns(splitFeet(p.cardFeet), feet, slots)
	}
	return coresession.BankColumns(splitBanks(p.cardBanks), splitBanks(p.banks), slots)
}

// goal is what a measurement is read against: the session's curve and columns, the reference, the
// transpose the keys sound at and the windows.
type goal struct {
	session   imposed
	a4        float64
	transpose int
	tol       float64
	beatTol   float64
}

// session returns what the active session imposes; no session answers the zero value.
func (s *Service) session() imposed {
	if s.sessions == nil {
		return imposed{}
	}
	g := s.sessions.Goal()
	if g.A4 <= 0 {
		return imposed{} // no session: the empty curve, and the app's own A4
	}
	return imposed{
		a4: g.A4, transpose: g.Transpose, reeds: g.Reeds, banks: joinBanks(g.Banks), bassFeet: joinFeet(g.BassFeet),
		cardBanks: joinBanks(g.CardBanks), cardFeet: joinFeet(g.CardFeet),
		curve: g.Curve, tol: g.Tolerance, beatTol: g.BeatTolerance,
	}
}

// adopt makes a running engine agree with the active session and the config file and returns the
// goal. It acts only on a change, so the tuner's own controls win in between; the config is re-read
// here rather than cached at Start, so the settings page takes effect without a restart.
func (s *Service) adopt() goal {
	p := s.session()
	tuner := appconfig.Get().Tuner

	// The instrument's own windows where it has them, the app default otherwise.
	rawTol, rawBeat := tuner.Tolerance, tuner.BeatTolerance
	if p.tol > 0 {
		rawTol = p.tol
	}
	if p.beatTol > 0 {
		rawBeat = p.beatTol
	}
	tol, beatTol := target.Tolerances(rawTol, rawBeat)

	s.mu.Lock()
	file := fileSettingsOf(s.cfg)
	sessionChanged := s.imposed != p
	changed := sessionChanged || s.file != file
	s.imposed, s.file = p, file
	file.apply(&s.cfg)
	cfg := impose(s.cfg, s.panel, p)
	pn := s.panel
	if r := s.liveLocked(); r != nil && changed {
		r.reconfigure(cfg)
	}
	s.mu.Unlock()

	if sessionChanged {
		if p.reeds > maxReeds {
			// The engine is clamped to what it can resolve; the instrument keeps its reeds and its curve its width.
			logger.Info(i18n.KeyLogSessionReedsClamped,
				logger.Int("instrument", p.reeds), logger.Int("engine", cfg.ReedCount))
		}
		emitEvent(EventSettings, settingsDTO(cfg, pn, p))
	}
	return goal{session: p, a4: cfg.A4, transpose: cfg.Transpose, tol: tol, beatTol: beatTol}
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

// decorate joins a measurement to the goal in reference units; a frame without slot values travels as is.
func decorate(m dsp.Measurement, g goal) MeasurementDTO {
	dto := MeasurementDTO{Measurement: m}

	// The pictures are packed to bytes here, at the wire; the Measurement keeps its floats for the golden tests.
	dto.Spectrum = encode(pack(m.Spectrum, 0, 1))
	dto.Equalizer = encode(pack(m.Equalizer, 0, dsp.EqualizerCeilingDB))
	dto.Waveform = encode(packSigned(m.Waveform))

	if len(m.Reeds) == 0 {
		return dto
	}
	reading := target.Reading{Note: m.Note, Transpose: g.transpose, Reeds: m.Reeds, Beats: m.Beats}
	cols := g.session.columns(len(m.Reeds))
	dto.ReedErrors = target.Errors(reading, g.session.curve, g.a4, g.tol, cols)
	dto.BeatErrors = target.BeatErrors(reading, g.session.curve, g.a4, g.beatTol, cols)
	return dto
}
