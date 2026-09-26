package meter

// analyzeCompound is the Chords/Bass handler event: one shared fine read, one band per working identity.
func (a *Analyzer) analyzeCompound(ring *SampleRing) Frame {
	a.History.BeginAnalysis()
	coarse, end, ok := a.spectrum(&a.coarsePlan, ring)
	a.last = coarseResult{end: end, ok: ok}
	if ok {
		// Retention ages run on the audio clock, so a replay of the same samples is deterministic.
		tick := func() uint32 { return uint32(ring.Appended() * 1000 / int64(a.cfg.SampleRate)) }
		a.last.compound = a.coarse.RefreshCompound(coarse, CompoundNomination{Chords: a.cfg.Chords, Bass: a.cfg.Bass,
			UpperOctaves: a.cfg.UpperOctaves, Octave: a.cfg.Octave, PitchClass: a.cfg.PitchClass, Recording: a.Recording,
			MeasureContinuously: a.cfg.Continuous}, &a.retention, a.cfg.ChordLabel, tick)
	}
	return a.measureCompound(ring, a.last.compound)
}

// measureCompound is the event after the coarse stage: the fine read for each working identity.
func (a *Analyzer) measureCompound(ring *SampleRing, c CompoundCoarse) Frame {
	f := Frame{CoarseEnd: a.last.end}
	if !a.last.ok {
		f.Status = FrameNoCoarse
		return f
	}
	if c.Published {
		a.Selection = c.Working
	}
	notes := c.Working
	f.Notes, f.Note, f.LearningStep = notes[:], notes[0], c.LearningStep
	defer func() { a.Previous = notes[0] + notes[1] + notes[2] }()
	if notes[0] == 0 {
		a.noNote()
		f.Status = FrameNoNote
		if notes == [3]int{} && a.Previous > 0 {
			a.ResetPending, f.ResetRequested = true, true
			f.Accepted = a.History.Evaluate([3]int{}, [3]float64{}, [3]int{})
		}
		return f
	}
	if a.History.Acquired && a.cfg.StopAfterLock {
		f.Status = FrameHeld
		return f
	}
	if a.cfg.Chords {
		a.publishLabel(c.Label, -1, false)
	}
	if a.cfg.Bass {
		a.publishLabel(a.cfg.Name(notes[0]), notes[0]/12, true)
	}
	fine, end, ok := a.spectrum(&a.finePlan, ring)
	f.FineEnd = end
	if !ok {
		f.Status = FrameNoFine
		return f
	}
	var records [3]CompoundRecord
	for slot, note := range notes {
		if note > 0 {
			band := a.fineBands[note-FirstActiveNote]
			a.records[note].Peak = ExtractPeaks(fine, band.LowerBin, band.UpperBin)[0]
		}
		records[slot] = a.records[note]
	}
	out := ConsumeCompound(&a.Display, &a.History, fine, notes, records, CompoundOptions{Chords: a.cfg.Chords,
		Bass: a.cfg.Bass, Hertz: a.cfg.Hertz, RequestedCount: len(a.cfg.Reeds), Table: &a.cfg.Table, Annotate: a.cfg.Name})
	f.Accepted = out.Call.Accepted
	f.SlotNotes = notes[:]
	f.Refined = Refined{Bins: append([]int(nil), a.Display.Readouts.Peaks...), Offsets: make([]float64, 3), Deviations: make([]float64, 3)}
	measured := []float64{Missing, Missing, Missing}
	for slot, s := range out.Slots {
		f.Refined.Offsets[slot], f.Refined.Deviations[slot] = s.Offset, s.Deviation
		if s.Computed {
			measured[slot] = s.Deviation
		}
	}
	// Chords never record; Bass takes are the SmeggTuner extension of the tone-mode rule.
	if a.cfg.Bass {
		f.Taken = a.take(measured)
	}
	f.Status = FrameMeasured
	return f
}
