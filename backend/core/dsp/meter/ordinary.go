package meter

// FrameStatus says which handler branch an analysis event took.
type FrameStatus uint8

const (
	FrameNoCoarse   FrameStatus = iota // ring shorter than the coarse window; nothing else ran
	FrameNoNote                        // no selected note: readouts dimmed, numbers kept
	FrameHeld                          // result acquired with StopAfterLock: fine analysis skipped
	FrameNoFine                        // ring shorter than the fine window
	FrameIneligible                    // a present slot beyond +/-58: history ran, readouts kept
	FrameMeasured                      // readouts, graph and visible lock written
)

// Frame is one analysis event's outcome; Refined holds new values only for FrameIneligible and FrameMeasured.
type Frame struct {
	Status         FrameStatus
	Note           int
	Notes          []int // band identities: one for ordinary, the working tuple in compound, one per band in a register
	SlotNotes      []int // the note each readout slot was measured against, 0 for a missing band
	LearningStep   int
	Refined        Refined
	Accepted       bool
	Taken          bool // this analysis wrote a take
	ResetRequested bool
	CoarseEnd      int64
	FineEnd        int64
}

// Analyze runs one handler event against the ring as it is now.
func (a *Analyzer) Analyze(ring *SampleRing) Frame {
	switch {
	case a.cfg.Chords || a.cfg.Bass:
		return a.analyzeCompound(ring)
	case len(a.cfg.Bands) >= 2:
		return a.analyzeRegister(ring)
	}
	a.History.BeginAnalysis()
	coarse, end, ok := a.spectrum(&a.coarsePlan, ring)
	a.last = coarseResult{end: end, ok: ok}
	if ok {
		a.last.frame = a.coarse.Refresh(coarse, Nomination{Octave: a.cfg.Octave, PitchClass: a.cfg.PitchClass,
			Locked: a.History.Locked, AcceptedNote: a.History.AcceptedNotes[0], Recording: a.Recording,
			MeasureContinuously: a.cfg.Continuous})
	}
	return a.measureOrdinary(ring, a.last.frame)
}

// measureOrdinary is the event after the coarse stage: the fine read against the selected note.
func (a *Analyzer) measureOrdinary(ring *SampleRing, c CoarseFrame) Frame {
	f := Frame{CoarseEnd: a.last.end}
	if !a.last.ok {
		f.Status = FrameNoCoarse
		return f
	}
	f.Note, f.LearningStep, f.Notes = c.Selected, c.LearningStep, []int{c.Selected}
	defer func() { a.Previous = f.Note }()
	if f.Note == 0 {
		a.noNote()
		f.Status = FrameNoNote
		if a.Previous > 0 {
			a.ResetPending, f.ResetRequested = true, true
			f.Accepted = a.History.Evaluate([3]int{}, [3]float64{}, [3]int{})
		}
		return f
	}
	if a.History.Acquired && a.cfg.StopAfterLock {
		f.Status = FrameHeld
		return f
	}
	a.publishLabel(a.cfg.Name(f.Note), f.Note/12, true)
	fine, end, ok := a.spectrum(&a.finePlan, ring)
	f.FineEnd = end
	if !ok {
		f.Status = FrameNoFine
		return f
	}
	band := a.fineBands[f.Note-FirstActiveNote]
	targets, covered := a.cfg.Targets.Note(f.Note, len(a.Display.Readouts.Frequencies))
	assigned := a.assign(fine, band, a.cfg.Reeds, targets)
	f.Refined = RefineSlots(fine, assigned, band.ReferenceBin)
	f.SlotNotes = make([]int, len(assigned))
	for i := range f.SlotNotes {
		f.SlotNotes[i] = f.Note
	}
	copy(a.Display.Readouts.Peaks, assigned)
	f.Accepted = a.History.Evaluate([3]int{f.Note, 0, 0}, [3]float64(f.Refined.Deviations[:3]), [3]int{band.BinCount(), -1, -1})
	if !f.Refined.DisplayEligible() {
		f.Status = FrameIneligible
		return f
	}
	f.Taken = a.take(f.Refined.Deviations)
	a.Display.Readouts.Ordinary(a.cfg.Table[f.Note], f.Refined.Deviations, targets, covered, a.cfg.ScaleRelative, a.cfg.Hertz)
	a.publishOrdinaryGraph(fine, band, f.Refined)
	a.Display.VisibleLock = a.History.Locked
	f.Status = FrameMeasured
	return f
}

// noNote is the no-note writer plus the handler stores around it.
func (a *Analyzer) noNote() {
	a.History.Acquired, a.History.Inhibited = false, false
	a.Display.Readouts.Dim()
	a.Display.VisibleLock, a.Display.Recorded = false, false
	a.Display.Level = 0
}

// publishLabel clears the level trace, the recorded mark and take inhibition on a label change; note modes compare the octave.
// The change test stands in for the native label helper's report, whose accidental parsing is not recovered.
func (a *Analyzer) publishLabel(text string, octave int, compareOctave bool) bool {
	changed := text != a.Display.Label.Text || compareOctave && octave != a.Display.Label.Octave
	if changed {
		a.Display.Level, a.Display.Recorded, a.History.Inhibited = 0, false, false
		a.Display.Graph.ReferenceBins = [3]float64{}
	}
	a.Display.Label = Label{Text: text, Octave: octave}
	return changed
}

// publishOrdinaryGraph raises the level to the band maximum over [0.97b, 1.03b) and publishes slot one only.
func (a *Analyzer) publishOrdinaryGraph(spectrum []float32, band NoteBand, r Refined) {
	b := band.ReferenceBin
	lo, hi := bandRound(float64(b*0.97)), bandRound(float64(b*1.03))
	var peak float32
	for k := lo; k < hi; k++ {
		if spectrum[k] > peak {
			peak = spectrum[k]
		}
	}
	if peak > a.Display.Level {
		a.Display.Level = peak
	}
	a.Display.Graph = CompoundGraph{Length: len(spectrum), Scratch: true, Peaks: [3]int(a.Display.Readouts.Peaks[:3]),
		Amplitudes: [3]float32{a.Display.Level}, Offsets: [3]float64(r.Offsets[:3]), ReferenceBins: [3]float64{b}}
}

// assign is Dirk's three-slot assignment for up to three reeds and the N-slot extension beyond; it returns len(targets) bins.
func (a *Analyzer) assign(spectrum []float32, band NoteBand, reeds []int, targets []float64) []int {
	if len(targets) > 3 {
		return AssignPeaksN(ExtractPeaksN(spectrum, band.LowerBin, band.UpperBin, len(targets)), targets, band.ReferenceBin)
	}
	out := AssignPeaks(ExtractPeaks(spectrum, band.LowerBin, band.UpperBin), reeds, [3]float64(targets), band.ReferenceBin)
	return out[:len(targets)]
}
