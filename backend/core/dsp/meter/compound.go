package meter

// CompoundRecord is what the compound consumer reads from a fine note record.
// A zero identity reads record zero, which is real state and still publishes its peak.
type CompoundRecord struct {
	Peak         int
	ReferenceBin float64
	BinCount     int
}

// RecordZero is the never-configured record zero as the record constructor leaves it.
var RecordZero = CompoundRecord{Peak: 0, ReferenceBin: -1, BinCount: -1}

// CompoundGraph is the fine-graph metadata; Scratch means it points at the shared fine spectrum.
type CompoundGraph struct {
	Length        int
	Scratch       bool
	Amplitudes    [3]float32
	Peaks         [3]int
	Offsets       [3]float64
	ReferenceBins [3]float64
}

type CompoundOptions struct {
	Chords, Bass, Hertz bool
	RequestedCount      int
	Table               *NoteTable
	Annotate            func(note int) string
}

// CompoundSlot holds invocation-local values; an uncomputed slot stays all zero, not Missing.
type CompoundSlot struct {
	Computed  bool
	Offset    float64
	Deviation float64
	Amplitude float32
	Reference float64
	Frequency float64
}

type HistoryCall struct {
	Table      bool
	Notes      [3]int
	Deviations [3]float64
	BinCounts  [3]int
	Accepted   bool
}

type CompoundOutcome struct {
	Slots          [3]CompoundSlot
	Resets         []int
	GraphPublished bool
	Call           HistoryCall
}

// ConsumeCompound is one compound numeric call after fine peak extraction; records[i] belongs to notes[i]
// (RecordZero or its current state for a zero identity) and spectrum is the half-length magnitude slice.
func ConsumeCompound(s *Display, h *History, spectrum []float32, notes [3]int, records [3]CompoundRecord, o CompoundOptions) CompoundOutcome {
	var out CompoundOutcome
	if notes[0] == 0 {
		out.Call = callHistory(s, h, false, [3]int{}, [3]float64{}, [3]int{})
		return out
	}
	r := &s.Readouts
	for slot := range notes {
		r.Peaks[slot] = records[slot].Peak
	}
	for slot, note := range notes {
		peak := r.Peaks[slot] // the published value: an earlier reset may have cleared it
		switch {
		case peak > 0:
			c := computeSlot(spectrum, peak, records[slot].ReferenceBin, o.Table[note])
			out.Slots[slot] = c
			writeReed(r, slot, c.Deviation, note, o.Annotate(note))
			if slot > 0 && o.Bass && r.Peaks[slot-1] != 0 {
				prev := out.Slots[slot-1]
				writePair(r, slot-1, HarmonicBeat(prev.Frequency, c.Frequency, prev.Reference, o.Hertz), o.Hertz)
			}
			if c.Deviation >= -50 && c.Deviation <= 50 {
				r.Accept(slot, c.Deviation, c.Frequency)
			}
		case peak == -1:
			r.Reset()
			s.Graph.ReferenceBins = [3]float64{}
			out.Resets = append(out.Resets, slot)
		}
	}
	if r.Peaks[0] <= 0 {
		out.Call = callHistory(s, h, false, [3]int{}, [3]float64{}, [3]int{})
		return out
	}
	n2, n3 := notes[1], notes[2]
	complete := (o.Chords && n2 != 0 && n3 != 0) || (o.Bass && n2 != 0 && (o.RequestedCount == 2 || n3 != 0))
	if complete {
		var d [3]float64
		var counts [3]int
		for slot := range notes {
			d[slot], counts[slot] = out.Slots[slot].Deviation, records[slot].BinCount
		}
		out.Call = callHistory(s, h, true, notes, d, counts)
	} else {
		out.Call = callHistory(s, h, false, [3]int{}, [3]float64{}, [3]int{})
	}
	for _, c := range out.Slots {
		if c.Amplitude > s.Level {
			s.Level = c.Amplitude
		}
	}
	second, third := o.RequestedCount >= 2 || o.Chords, o.RequestedCount >= 3 || o.Chords
	g := CompoundGraph{Length: len(spectrum), Scratch: true, Peaks: [3]int(r.Peaks)}
	g.Amplitudes[0], g.ReferenceBins[0] = s.Level, records[0].ReferenceBin
	if second {
		g.Amplitudes[1], g.ReferenceBins[1] = s.Level, records[1].ReferenceBin
	}
	if third {
		g.Amplitudes[2], g.ReferenceBins[2] = s.Level, records[2].ReferenceBin
	}
	for slot, c := range out.Slots {
		g.Offsets[slot] = c.Offset
	}
	s.Graph, out.GraphPublished = g, true
	return out
}

func computeSlot(spectrum []float32, peak int, referenceBin, reference float64) CompoundSlot {
	offset := Interpolate(spectrum[peak-1], spectrum[peak], spectrum[peak+1])
	d := DeviationUnits(float64(peak)+offset, referenceBin)
	// No ordinary sentinel or +/-58 gate here: a positive peak always yields a frequency.
	return CompoundSlot{true, offset, d, spectrum[peak], reference, FrequencyHz(reference, d)}
}

func writeReed(r *Readouts, slot int, value float64, note int, annotation string) {
	p := &r.Panels[slot*2]
	p.Current, p.Signed, p.Decimals, p.Dimmed, p.Annotation, p.Octave = value, true, 1, false, annotation, note/12
}

func writePair(r *Readouts, pair int, value float64, hertz bool) {
	p := &r.Panels[pair*2+1]
	p.Current, p.Signed, p.Decimals, p.Dimmed = value, false, 1, false
	if hertz {
		p.Decimals = 2
	}
}

func callHistory(s *Display, h *History, table bool, notes [3]int, d [3]float64, counts [3]int) HistoryCall {
	accepted := h.Evaluate(notes, d, counts)
	s.VisibleLock = h.Locked
	return HistoryCall{table, notes, d, counts, accepted}
}
