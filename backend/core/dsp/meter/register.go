package meter

// Band is one octave band of a register: its offset in semitones above the lowest band and its reed count.
type Band struct {
	Offset, Reeds int
}

// DirkBass reports whether a band layout is the reference Bass mode: one reed per band, a base and the
// enabled upper boxes base+12 and base+24 in increasing order. Such layouts run the Bass path, never the extension.
func DirkBass(bands []Band) (upper [2]bool, ok bool) {
	if len(bands) < 2 || len(bands) > 3 || bands[0] != (Band{0, 1}) {
		return [2]bool{}, false
	}
	for i, b := range bands[1:] {
		if b.Reeds != 1 || b.Offset != 12 && b.Offset != 24 || b.Offset <= bands[i].Offset {
			return [2]bool{}, false
		}
		upper[b.Offset/12-1] = true
	}
	return upper, true
}

// RegisterCandidates extends the Bass candidate rule to any band offsets: the base from the unlocked
// ordinary nomination, each higher band validated and above a tenth of the base score.
func (t *ScoreTable) RegisterCandidates(n CompoundNomination, bands []Band) []int {
	base := t.Ordinary(Nomination{Octave: n.Octave, PitchClass: n.PitchClass, Recording: n.Recording,
		MeasureContinuously: n.MeasureContinuously})
	out := make([]int, len(bands))
	out[0] = base
	for i, b := range bands[1:] {
		note := base + b.Offset
		if base == 0 || note > 108 {
			continue
		}
		validated := t.Ordinary(Nomination{Octave: note / 12, PitchClass: note % 12, Recording: n.Recording,
			MeasureContinuously: n.PitchClass >= 0 && n.MeasureContinuously})
		if validated != 0 && t.Scores[note] > t.Scores[base]/10 {
			out[i+1] = note
		}
	}
	return out
}

// RefreshRegister is RefreshCompound for the register extension: selection unless learning, then a
// membership refresh of the band notes.
func (s *CoarseState) RefreshRegister(spectrum []float32, n CompoundNomination, bands []Band) (notes []int, step int) {
	step = s.Learning.Prepare(&s.Notes, spectrum, s.Background)
	s.before(spectrum)
	notes = make([]int, len(bands))
	if s.Learning.SelectionAllowed() {
		table := ScoreTable{Scores: s.Scores}
		for i := range s.Notes {
			table.Foreground[i+FirstActiveNote] = s.Notes[i].Foreground
		}
		notes = table.RegisterCandidates(n, bands)
	}
	s.after(spectrum, notes...)
	return notes, step
}
