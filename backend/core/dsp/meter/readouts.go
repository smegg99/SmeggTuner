package meter

// Panel is one numeric readout as the measurement writer leaves it, before painting.
type Panel struct {
	Current    float64
	Target     float64
	Signed     bool
	Decimals   int
	Enabled    bool
	Dimmed     bool
	Annotation string
	Octave     int
}

// FrequencySlot is a retained per-reed frequency cache; numbers survive invalidation.
type FrequencySlot struct {
	Deviation float64
	Frequency float64
	Position  float64
	Motion    float64
	Valid     bool
}

// Readouts panels run reed1, pair12, reed2, ..., reedN. Peaks are the published peak bins. Dirk modes use N = 3.
type Readouts struct {
	Panels          []Panel
	Frequencies     []FrequencySlot
	FrequencyDimmed bool
	Peaks           []int
}

// NewReadouts is an N-slot state with dimmed missing panels.
func NewReadouts(slots int) Readouts {
	r := Readouts{Panels: make([]Panel, 2*slots-1), Frequencies: make([]FrequencySlot, slots), Peaks: make([]int, slots), FrequencyDimmed: true}
	for i := range r.Panels {
		r.Panels[i] = Panel{Current: Missing, Target: Missing, Signed: true, Decimals: 1, Enabled: true, Dimmed: true}
	}
	return r
}

// Dim is the no-note writer: peaks clear, everything dims, numbers and validity stay.
func (r *Readouts) Dim() {
	clear(r.Peaks)
	r.FrequencyDimmed = true
	for i := range r.Panels {
		r.Panels[i].Dimmed = true
	}
}

// Reset is the broad reset: validity and panels clear; frequency numbers and the shared dim stay.
func (r *Readouts) Reset() {
	clear(r.Peaks)
	for i := range r.Frequencies {
		r.Frequencies[i].Valid = false
	}
	for i := range r.Panels {
		p := &r.Panels[i]
		p.Current, p.Target, p.Annotation, p.Octave, p.Dimmed = Missing, Missing, "", 0, true
	}
}

// Accept replaces an in-range cache; a newly valid slot restarts its needle animation.
func (r *Readouts) Accept(slot int, deviation, frequency float64) {
	f := &r.Frequencies[slot]
	if !f.Valid {
		f.Position, f.Motion = 0, 0
	}
	f.Deviation, f.Frequency, f.Valid = deviation, frequency, true
	r.FrequencyDimmed = false
}

// Ordinary writes one ordinary-mode measurement; the caller already applied the +/-58 gate.
// deviations and targets have one entry per slot.
func (r *Readouts) Ordinary(reference float64, deviations, targets []float64, coverage, scaleRelative, hertz bool) {
	references := make([]float64, len(deviations))
	for i := range references {
		references[i] = reference
	}
	r.Slots(references, deviations, targets, coverage, scaleRelative, hertz)
}

// Slots is Ordinary with a reference per slot, for registers spanning octaves; a pair is stated at its
// upper slot's pitch, where an octave pair's beat is the two deviations' difference.
func (r *Readouts) Slots(references, deviations, targets []float64, coverage, scaleRelative, hertz bool) {
	measured := make([]float64, len(deviations))
	present := make([]bool, len(deviations))
	for slot, d := range deviations {
		if d > Missing {
			measured[slot], present[slot] = FrequencyHz(references[slot], d), true
		}
		value := d
		if !scaleRelative {
			value = Residual(d, targets[slot])
		}
		target := Missing
		if coverage && targets[slot] > Missing {
			target = targets[slot]
		}
		p := &r.Panels[slot*2]
		p.Current, p.Target, p.Signed, p.Decimals, p.Dimmed, p.Annotation = value, target, true, 1, false, ""
	}
	for pair := range len(deviations) - 1 {
		reference := references[pair+1]
		desired := TargetPair(targets[pair], targets[pair+1], reference, coverage, hertz)
		both := present[pair] && present[pair+1]
		beat := Missing
		switch {
		case !both:
		case hertz || reference == references[pair]:
			beat = HarmonicBeat(measured[pair], measured[pair+1], reference, hertz)
		default:
			beat = (HarmonicBeat(measured[pair], measured[pair+1], reference, true) * 100) / float64(reference*unitFraction)
		}
		decimals := 1
		if hertz && both {
			decimals = 2
		}
		p := &r.Panels[pair*2+1]
		p.Current = SubtractPairTarget(beat, desired, scaleRelative, both)
		p.Target, p.Signed, p.Decimals, p.Dimmed, p.Annotation = desired, !scaleRelative, decimals, false, ""
	}
	for slot, d := range deviations {
		switch {
		case d <= Missing:
			r.Frequencies[slot].Valid = false
		case d >= -50 && d <= 50:
			r.Accept(slot, d, measured[slot])
		}
	}
}
