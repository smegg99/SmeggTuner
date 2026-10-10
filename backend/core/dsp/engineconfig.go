// backend/core/dsp/engineconfig.go
package dsp

import (
	"errors"
	"slices"

	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// EngineConfig is everything the engine is configured by; it depends only on core/tuning and the meter stages.
type EngineConfig struct {
	A4         float64
	Transpose  int   // 0..11, semitones the keys sound higher: bands move, notes, labels and Targets stay the key
	ReedCount  int   // reeds in one band: 1..3 is the reference mode, 4..6 the N-slot extension
	Reeds      []int // enabled reed boxes of one band, a subset of 1..3; nil is 1..ReedCount, else it sets ReedCount
	ManualNote tuning.Note
	Scale      tuning.ScaleNaming
	Hum50      bool
	Hum60      bool

	// Calibration is the sample-clock correction in percent; the effective rate is nominal*(1+p/100).
	Calibration float64

	// LearnOnStart runs the eight-update background learning when a run begins (a microphone).
	LearnOnStart bool

	// Octaves is the register layout, Offset in semitones from the played key (-12 a 16', +12 a 4').
	// Empty or offset-0 only is one band of ReedCount reeds.
	Octaves []OctaveRequest

	// Bass marks the bass side: Octaves are feet, and the lowest sounding foot names the note.
	Bass bool

	// Chords measures three-note chords with the reference recognition and retention.
	Chords bool

	Detection     Detection // "" is DetectionBalanced
	StopAfterLock bool
	Continuous    bool // a manual note skips foreground validation (MeasureContinuously) while not recording
	BeatsInHertz  bool

	// Targets are tuning targets in deviation units, one curve per measured slot in Measurement.Reeds
	// order, indexed by the played key (note 4..108 = MIDI 16..120); nil means none.
	Targets        []meter.NoteCurve
	TargetCoverage meter.NoteCoverage
	TargetRelative bool // readouts show deviation minus target
}

// OctaveRequest is one octave band of a register: where it sounds relative to the key and how many reeds share it.
type OctaveRequest struct {
	Offset int
	Reeds  int
}

// DefaultEngineConfig returns the engine's own defaults.
func DefaultEngineConfig() EngineConfig {
	c := EngineConfig{}
	c.fill()
	return c
}

func (c *EngineConfig) fill() {
	if c.A4 == 0 {
		c.A4 = 440
	}
	if _, ok := detections[c.Detection]; !ok {
		c.Detection = DetectionBalanced
	}
	c.Transpose = min(11, max(0, c.Transpose))
	var reeds []int
	for id := 1; id <= 3; id++ {
		if slices.Contains(c.Reeds, id) {
			reeds = append(reeds, id)
		}
	}
	c.Reeds = reeds
	if len(reeds) > 0 {
		c.ReedCount = len(reeds)
	}
	if c.ReedCount == 0 {
		c.ReedCount = 1
	}
	for i := 1; i < len(c.Octaves); i++ {
		for j := i; j > 0 && c.Octaves[j].Offset < c.Octaves[j-1].Offset; j-- {
			c.Octaves[j], c.Octaves[j-1] = c.Octaves[j-1], c.Octaves[j]
		}
	}
	for i := range c.Octaves {
		c.Octaves[i].Reeds = max(1, c.Octaves[i].Reeds)
	}
}

var chordSuffix = map[meter.ChordKind]string{
	meter.ChordMajor: "", meter.ChordMinor: "m", meter.ChordDiminished: "dim", meter.ChordAugmented: "+",
	meter.ChordSuspendedSecond: "sus2", meter.ChordSuspendedFourth: "sus4", meter.ChordMajorFlatFifth: "-5",
}

// ErrSampleRate reports a capture too slow for the band table: the top band would lie above Nyquist.
var ErrSampleRate = errors.New("dsp: sample rate too low for the note table")

// table is the band table the analysis runs on.
func (c *EngineConfig) table() meter.NoteTable { return meter.EqualTemperament(c.A4, c.Transpose) }

// fits reports whether the top band's upper edge stays below Nyquist at sampleRate.
func (c *EngineConfig) fits(sampleRate int) bool {
	t := c.table()
	top := meter.FirstActiveNote + meter.ActiveNotes - 1
	return float64(sampleRate)*meter.CalibrationFactor(c.Calibration, true) > t[top]+t[top+1]
}

// analyzer builds the meter configuration for a capture at sampleRate.
func (c *EngineConfig) analyzer(sampleRate int) meter.AnalyzerConfig {
	bands, keyAbove := c.bands()
	d := c.detection()
	cfg := meter.AnalyzerConfig{
		SampleRate: sampleRate, RateFactor: meter.CalibrationFactor(c.Calibration, true),
		Table: c.table(), Reeds: c.reedIDs(), Bands: bands,
		Hum50: c.Hum50, Hum60: c.Hum60, StopAfterLock: c.StopAfterLock,
		ScaleRelative: !c.TargetRelative, Hertz: c.BeatsInHertz,
		Octave: -1, PitchClass: -1, Continuous: c.Continuous, Settle: d.settle,
		Release: d.extended, QuietSeed: d.extended, Follow: d.extended, Fresh: d.extended, Watch: d.watch,
		Name: func(index int) string { return tuning.Note(index + 12).Name(c.Scale) },
	}
	cfg.Targets = c.targets()
	if c.Chords {
		cfg.Chords, cfg.Reeds, cfg.Bands = true, []int{1, 2, 3}, nil
		cfg.ChordLabel = func(id meter.ChordIdentity) string {
			return tuning.Note(id.Root+12).PitchName(c.Scale) + chordSuffix[id.Kind]
		}
	}
	if c.ManualNote.Valid() {
		index := int(c.ManualNote) - 12 - keyAbove
		if index < meter.FirstActiveNote || index >= meter.FirstActiveNote+meter.ActiveNotes {
			// Nominate no note instead of scanning or indexing outside the band table.
			cfg.Octave, cfg.PitchClass = 0, 0
		} else {
			cfg.Octave, cfg.PitchClass = index/12, index%12
		}
	}
	return cfg
}

// targets moves the key-indexed curves onto the note the analysis runs on, the lowest band of a
// register, and onto the meter slot each measured slot reads.
func (c *EngineConfig) targets() meter.TargetTables {
	_, keyAbove := c.bands()
	ids := c.meterSlots()
	t := meter.TargetTables{Slots: make([]meter.NoteCurve, slices.Max(ids)+1)}
	for i := range meter.ActiveNotes {
		from := i + keyAbove
		covered := from >= 0 && from < meter.ActiveNotes
		for s := range t.Slots {
			t.Slots[s][i] = meter.Missing
		}
		for k, id := range ids {
			if k < len(c.Targets) && covered {
				t.Slots[id][i] = c.Targets[k][from]
			}
		}
		t.Coverage[i] = covered && c.TargetCoverage[from]
	}
	return t
}
