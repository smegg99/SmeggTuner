package meter

// analyzeRegister is the register extension: Bass-style selection over arbitrary bands, ordinary slot
// assignment inside each band, one reading across all slots. It never runs for a Dirk Bass layout.
func (a *Analyzer) analyzeRegister(ring *SampleRing) Frame {
	a.History.BeginAnalysis()
	coarse, end, ok := a.spectrum(&a.coarsePlan, ring)
	a.last = coarseResult{end: end, ok: ok}
	if ok {
		a.last.notes, a.last.step = a.coarse.RefreshRegister(coarse, CompoundNomination{Octave: a.cfg.Octave,
			PitchClass: a.cfg.PitchClass, Recording: a.Recording, MeasureContinuously: a.cfg.Continuous}, a.cfg.Bands)
	}
	return a.measureRegister(ring, a.last.notes, a.last.step)
}

// measureRegister is the event after the coarse stage: the fine read across the selected bands.
func (a *Analyzer) measureRegister(ring *SampleRing, notes []int, step int) Frame {
	f := Frame{CoarseEnd: a.last.end}
	if !a.last.ok {
		f.Status = FrameNoCoarse
		return f
	}
	f.Note, f.LearningStep, f.Notes = notes[0], step, notes
	sum := 0
	for _, n := range notes {
		sum += n
	}
	defer func() { a.Previous = sum }()
	if notes[0] == 0 {
		a.noNote()
		f.Status = FrameNoNote
		if sum == 0 && a.Previous > 0 {
			a.ResetPending, f.ResetRequested = true, true
			f.Accepted = a.History.Evaluate([3]int{}, [3]float64{}, [3]int{})
		}
		return f
	}
	if a.History.Acquired && a.cfg.StopAfterLock {
		f.Status = FrameHeld
		return f
	}
	a.publishLabel(a.cfg.Name(notes[0]), notes[0]/12, true)
	fine, end, ok := a.spectrum(&a.finePlan, ring)
	f.FineEnd = end
	if !ok {
		f.Status = FrameNoFine
		return f
	}
	var references, deviations, targets []float64
	var bins []int
	var tuple [3]int
	var first [3]float64
	var counts [3]int
	for i, b := range a.cfg.Bands {
		note := notes[i]
		slotTargets := make([]float64, b.Reeds)
		bandBins := make([]int, b.Reeds)
		for k := range slotTargets {
			slotTargets[k] = Missing
		}
		if note > 0 {
			band := a.fineBands[note-FirstActiveNote]
			t, _ := a.cfg.Targets.Note(notes[0], len(bins)+b.Reeds)
			slotTargets = t[len(bins):]
			bandBins = a.assign(fine, band, registerSlots(b.Reeds), registerPad(slotTargets))[:b.Reeds]
			r := RefineSlots(fine, bandBins, band.ReferenceBin)
			deviations = append(deviations, r.Deviations...)
			if i < 3 {
				tuple[i], counts[i] = note, band.BinCount()
				if r.Deviations[0] > Missing {
					first[i] = r.Deviations[0]
				}
			}
		} else {
			for range b.Reeds {
				deviations = append(deviations, Missing)
			}
		}
		at := note
		if at == 0 {
			at = min(notes[0]+b.Offset, len(a.cfg.Table)-1)
		}
		for range b.Reeds {
			references = append(references, a.cfg.Table[at])
		}
		bins = append(bins, bandBins...)
		for range b.Reeds {
			f.SlotNotes = append(f.SlotNotes, note)
		}
		targets = append(targets, slotTargets...)
	}
	copy(a.Display.Readouts.Peaks, bins)
	if tuple[1] != 0 && (len(a.cfg.Bands) < 3 || tuple[2] != 0) {
		f.Accepted = a.History.Evaluate(tuple, first, counts)
	} else {
		f.Accepted = a.History.Evaluate([3]int{}, [3]float64{}, [3]int{})
	}
	f.Taken = a.take(deviations)
	_, covered := a.cfg.Targets.Note(notes[0], 0)
	a.Display.Readouts.Slots(references, deviations, targets, covered, a.cfg.ScaleRelative, a.cfg.Hertz)
	f.Refined = Refined{Bins: bins, Offsets: make([]float64, len(bins)), Deviations: deviations}
	a.Display.VisibleLock = a.History.Locked
	f.Status = FrameMeasured
	return f
}

// registerSlots is slots 1..n.
func registerSlots(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i + 1
	}
	return out
}

// registerPad widens a band's targets to the three-slot shape when it has fewer reeds.
func registerPad(t []float64) []float64 {
	for len(t) < 3 {
		t = append(t, Missing)
	}
	return t
}
