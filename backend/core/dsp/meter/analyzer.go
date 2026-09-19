package meter

// AnalyzerConfig is one fixed measurement configuration; changing any field means a new Analyzer.
type AnalyzerConfig struct {
	SampleRate    int
	RateFactor    float64 // calibration factor C; 1 when uncalibrated
	Table         NoteTable
	Targets       TargetTables
	Reeds         []int // measured slots, increasing; Dirk modes use 1..3, more is the N-slot extension
	Hum50, Hum60  bool
	StopAfterLock bool
	ScaleRelative bool // readouts are raw deviations, not target residuals
	Hertz         bool // beats in Hz rather than deviation units
	Octave        int  // manual octave, -1 when unset
	PitchClass    int  // manual pitch class, -1 when unset
	Continuous    bool // MeasureContinuously: a manual note skips validation
	Name          func(note int) string
	Chords, Bass  bool    // compound modes; at most one is set
	UpperOctaves  [2]bool // Bass: also measure base+12 and base+24
	Bands         []Band  // register layout; a Dirk Bass layout becomes Bass, any other two or more bands the extension
	ChordLabel    func(ChordIdentity) string
	Settle        int // History.Settle; the reference locks on the readings alone

	// The responsiveness extensions; all off is the reference.
	Release   bool // History.Release
	QuietSeed bool // Learning.QuietSeed
	Follow    bool // CoarseState.Follow
	Fresh     bool // CoarseState.Fresh
	Watch     int  // quiet short-window passes before Gone lets a note go; 0 never watches; wants Fresh
}

// Label is the published note label and its octave field.
type Label struct {
	Text   string
	Octave int
}

// Display is what the handler leaves for the screen before painting.
type Display struct {
	Readouts    Readouts
	Graph       CompoundGraph
	Level       float32
	Label       Label
	VisibleLock bool
	Recorded    bool // a take was written; the note window says so until the label changes
}

// Analyzer is the serial analysis handler: one Analyze call per posted capture message.
type Analyzer struct {
	cfg          AnalyzerConfig
	coarse       CoarseState
	History      History
	Display      Display
	Previous     int  // previous selection identity; a sum of identities in compound mode
	ResetPending bool // deferred capture reset, consumed by the next capture pass
	Recording    bool // the reference record index is set: takes are written and the manual bypass is off
	TakeRequest  bool // take the next eligible analysis without a lock; cleared only by a take
	coarsePlan   analyzerPlan
	finePlan     analyzerPlan
	fineBands    [ActiveNotes]NoteBand
	retention    ChordRetention
	records      [109]CompoundRecord // fine records by note; 0 is record zero
	Selection    [3]int              // last published coarse selection tuple
	slots        int                 // readout slots, at least the reference's three
	last         coarseResult        // the coarse stage of the last Analyze, for Remeasure
	watch        watch
}

// coarseResult is the coarse stage's output, reused by Remeasure until the next Analyze.
type coarseResult struct {
	end      int64
	ok       bool
	frame    CoarseFrame    // ordinary
	compound CompoundCoarse // Chords and Bass
	notes    []int          // register
	step     int            // register
}

// Remeasure repeats the fine stage of the last Analyze against the notes it selected, without the
// coarse stage or the lock countdown: readings and the lock history move between two native events
// while detection, the background and the countdown keep Dirk's pace.
func (a *Analyzer) Remeasure(ring *SampleRing) Frame {
	switch {
	case a.cfg.Chords || a.cfg.Bass:
		c := a.last.compound
		c.LearningStep, c.Published = 0, false // a reused stage neither learns nor republishes
		return a.measureCompound(ring, c)
	case len(a.cfg.Bands) >= 2:
		return a.measureRegister(ring, a.last.notes, 0)
	}
	c := a.last.frame
	c.LearningStep = 0
	return a.measureOrdinary(ring, c)
}

// NewAnalyzer starts from a known capture reset with idle learning and empty history.
func NewAnalyzer(cfg AnalyzerConfig) *Analyzer { return newAnalyzer(cfg, nil) }

// Rebuild is NewAnalyzer for cfg that takes over a's FFT plans where the window sizes match, so a
// settings change costs no new tables; a must not analyse again.
func (a *Analyzer) Rebuild(cfg AnalyzerConfig) *Analyzer { return newAnalyzer(cfg, a) }

func newAnalyzer(cfg AnalyzerConfig, prev *Analyzer) *Analyzer {
	slots := max(3, len(cfg.Reeds))
	if upper, ok := DirkBass(cfg.Bands); ok && !cfg.Chords {
		cfg.Bass, cfg.UpperOctaves, cfg.Reeds, cfg.Bands = true, upper, registerSlots(len(cfg.Bands)), nil
	} else if len(cfg.Bands) >= 2 {
		slots = 0
		for _, b := range cfg.Bands {
			slots += b.Reeds
		}
	}
	a := &Analyzer{cfg: cfg, slots: max(3, slots)}
	var coarse, fine *analyzerPlan
	var watched *watch
	if prev != nil {
		coarse, fine, watched = &prev.coarsePlan, &prev.finePlan, &prev.watch
	}
	a.coarsePlan = newAnalyzerPlan(cfg, NewWindowPlan(cfg.SampleRate, false), coarse)
	a.finePlan = newAnalyzerPlan(cfg, NewWindowPlan(cfg.SampleRate, true), fine)
	if cfg.Watch > 0 {
		a.watch = newWatch(cfg, watched)
	}
	for i := range a.fineBands {
		n := i + FirstActiveNote
		a.fineBands[i] = NewNoteBand(cfg.Table[n-1], cfg.Table[n], cfg.Table[n+1], a.finePlan.window.FFTSize, cfg.SampleRate, cfg.RateFactor)
	}
	a.Restart()
	return a
}

// Restart is a capture that begins now: the plans stay, everything heard or learned is forgotten.
func (a *Analyzer) Restart() {
	cfg := a.cfg
	var coarseBands [ActiveNotes]NoteBand
	for i := range coarseBands {
		n := i + FirstActiveNote
		coarseBands[i] = NewNoteBand(cfg.Table[n-1], cfg.Table[n], cfg.Table[n+1], a.coarsePlan.window.FFTSize, cfg.SampleRate, cfg.RateFactor)
	}
	a.coarse = NewCoarseState(&coarseBands)
	a.coarse.Learning.QuietSeed, a.coarse.Follow, a.coarse.Fresh = cfg.QuietSeed, cfg.Follow, cfg.Fresh
	a.records[0] = RecordZero
	for i, b := range a.fineBands {
		a.records[i+FirstActiveNote] = CompoundRecord{ReferenceBin: b.ReferenceBin, BinCount: b.BinCount()}
	}
	a.History, a.Previous, a.ResetPending, a.Selection, a.retention = History{Settle: cfg.Settle, Release: cfg.Release}, 0, false, [3]int{}, ChordRetention{}
	a.last = coarseResult{}
	a.watch.note = 0
	a.Display = Display{Readouts: NewReadouts(a.slots)}
}

// Learn requests the eight-update background learning pass.
func (a *Analyzer) Learn() { a.coarse.Learning.Request = true }

// Learning is true while a learning request or pass is active.
func (a *Analyzer) Learning() bool {
	return a.coarse.Learning.Request || a.coarse.Learning.Counter != -1
}

// spectrum reads one window and returns its masked, hum-attenuated magnitudes; ok is false when the ring is short.
func (a *Analyzer) spectrum(p *analyzerPlan, ring *SampleRing) (spectrum []float32, endpoint int64, ok bool) {
	endpoint, ok = p.window.Read(ring, p.input)
	if !ok {
		return nil, endpoint, false
	}
	spectrum = Magnitudes(p.spectrum, p.transform.Packed(p.spectrum, p.input), p.bounds)
	if a.cfg.Hum50 {
		AttenuateHum(spectrum, a.cfg.SampleRate, a.cfg.RateFactor, 50)
	}
	if a.cfg.Hum60 {
		AttenuateHum(spectrum, a.cfg.SampleRate, a.cfg.RateFactor, 60)
	}
	return spectrum, endpoint, true
}

// Spectrum is the last fine magnitude array; it is overwritten by the next analysis.
func (a *Analyzer) Spectrum() []float32 {
	return a.finePlan.spectrum[:a.finePlan.window.MagnitudeCount()]
}

// ReferenceBin is a note's fine reference bin in Spectrum.
func (a *Analyzer) ReferenceBin(note int) float64 {
	return a.fineBands[note-FirstActiveNote].ReferenceBin
}

// Scores is the coarse peak score of notes 4..108 and the background aggregate they are judged against.
func (a *Analyzer) Scores() (scores []float32, background float32) {
	return a.coarse.Scores[FirstActiveNote : FirstActiveNote+ActiveNotes], a.coarse.Background
}

// Carry keeps the learned coarse background and its learning state from prev when both analyzers see
// the same coarse spectrum (rate, calibration, note table, hum); reed, mode and display changes then
// cost no relearning. History and readouts start fresh because slots and mode may differ.
func (a *Analyzer) Carry(prev *Analyzer) bool {
	if prev == nil || prev.cfg.SampleRate != a.cfg.SampleRate || prev.cfg.RateFactor != a.cfg.RateFactor ||
		prev.cfg.Table != a.cfg.Table || prev.cfg.Hum50 != a.cfg.Hum50 || prev.cfg.Hum60 != a.cfg.Hum60 {
		return false
	}
	a.coarse = prev.coarse
	a.coarse.Learning.QuietSeed, a.coarse.Follow, a.coarse.Fresh = a.cfg.QuietSeed, a.cfg.Follow, a.cfg.Fresh
	return true
}
