package tuner

import (
	"sort"
	"strconv"
	"strings"

	appconfig "smegg.me/smeggtuner/common/config"
	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/dsp"
	coresession "smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
)

// imposed is what the active session imposes on the engine: reference pitch, reed count and the
// pulled register's banks; the zero value means no session, and reeds is the instrument's count (not
// what the engine resolves) so it compares equal across ticks. banks is the card-order join ("L,M1")
// rather than a slice so the struct stays comparable.
type imposed struct {
	a4    float64
	reeds int
	banks string
	// bassFeet is the sounding bass ladder ("32,16,8") while the bench faces the bass side; empty
	// on the treble side. Like banks, a join keeps the struct comparable.
	bassFeet string
	// profileRev fingerprints the calibrated rank voices, so a fresh take re-imposes them; the
	// profile itself travels beside this struct (a slice cannot live in a comparable one).
	profileRev int64
	// tol and beatTol are the instrument's own judging windows in cents, or zero; zero falls back to the app default in adopt.
	tol     float64
	beatTol float64
}

// joinBanks and splitBanks carry the register's banks through the comparable imposed struct.
func joinBanks(banks []coresession.Bank) string {
	if len(banks) == 0 {
		return ""
	}
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
	if len(feet) == 0 {
		return ""
	}
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

// bassProfilesFor resolves foot-keyed bass calibrations onto the pulled ladder's octave slots: the
// largest sounding foot is the base, so a rank's offset moves with the register - which is why the
// foot, not the offset, is what the session stores.
func bassProfilesFor(feet []int, profs []coresession.BassProfile) []dsp.RankProfile {
	reqs := coresession.OctavesOfFeet(feet)
	if len(reqs) == 0 {
		return nil
	}
	sorted := append([]int(nil), feet...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	var out []dsp.RankProfile
	for _, p := range profs {
		for i, f := range sorted {
			if f == p.Foot {
				out = append(out, dsp.RankProfile{Offset: reqs[i].Offset, Note: p.Note, R2: p.R2, R4: p.R4})
				break
			}
		}
	}
	return out
}

// goal is the target a measurement is read against: curve, reference, and reed/beat windows. A nil curve is the empty curve.
type goal struct {
	curve   *target.Curve
	a4      float64
	tol     float64
	beatTol float64
}

// session returns what the active session imposes, its calibrated rank profile and the goal curve
// in force; no session answers the zero set, and the curve is swapped rather than edited, so
// reading it needs no lock.
func (s *Service) session() (imposed, []dsp.RankProfile, *target.Curve) {
	if s.sessions == nil {
		return imposed{}, nil, nil
	}
	g := s.sessions.Goal()
	if g.A4 <= 0 {
		return imposed{}, nil, nil // no session: the empty curve, and the app's own A4
	}
	prof := g.Profile
	if len(g.BassFeet) > 0 {
		prof = bassProfilesFor(g.BassFeet, g.BassProfiles)
	}
	return imposed{
		a4: g.A4, reeds: g.Reeds, banks: joinBanks(g.Banks), bassFeet: joinFeet(g.BassFeet),
		profileRev: g.ProfileRev, tol: g.Tolerance, beatTol: g.BeatTolerance,
	}, prof, g.Curve
}

// adopt makes a running engine agree with the active session and returns the goal and rules; it acts only when the session changed, so the tuner's own reed-count control wins in between, and the config is re-read here (not cached at Start) so a tightened tolerance takes effect without a restart.
func (s *Service) adopt() (goal, rules) {
	p, prof, curve := s.session()
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
	cfg := impose(s.cfg, p, prof)
	manual := s.cfg.ManualNote != tuning.Note(autoNote)
	s.mu.Unlock()

	if changed {
		if r := s.current(); r != nil {
			r.engine.Update(func(c *dsp.EngineConfig) {
				c.A4 = cfg.A4
				c.ReedCount = cfg.ReedCount
				c.Octaves = cfg.Octaves
				c.Profiles = cfg.Profiles
				c.ProfileHarmonics = cfg.ProfileHarmonics
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

// impose folds the active session over an engine config: its reference pitch becomes the engine's,
// its instrument sets the reed count (up to what the tones mode can split) and the pulled
// register's banks - or, facing the bass side, the sounding ladder's feet - become the compound
// octave layout. Nil when the register stays in one octave, which keeps the single band and its
// musette machinery. A solo-rank register on either side turns harmonic profiling on: the
// calibration sweep of such a register is what teaches the profile the engine reads back.
func impose(c dsp.EngineConfig, p imposed, prof []dsp.RankProfile) dsp.EngineConfig {
	if p.a4 > 0 {
		c.A4 = p.a4
	}
	if p.reeds > 0 {
		c.ReedCount = clampReeds(p.reeds)
	}
	if feet := splitFeet(p.bassFeet); len(feet) > 0 {
		c.Octaves = coresession.OctavesOfFeet(feet)
		c.ProfileHarmonics = len(feet) == 1
	} else {
		banks := splitBanks(p.banks)
		c.Octaves = coresession.OctavesOf(banks)
		c.ProfileHarmonics = len(banks) == 1
	}
	c.Profiles = prof
	return c
}
