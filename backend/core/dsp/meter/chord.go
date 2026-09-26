package meter

// CompoundNomination is the Chords/Bass dispatch state; a set PitchClass needs 12*Octave+PitchClass of 0 or 4..108.
type CompoundNomination struct {
	Chords, Bass        bool
	UpperOctaves        [2]bool
	Octave              int
	PitchClass          int
	Recording           bool
	MeasureContinuously bool
}

// ChordKind is a recognized interval class; the zero value is not a kind.
type ChordKind uint8

const (
	ChordSuspendedSecond ChordKind = iota + 1
	ChordDiminished
	ChordMinor
	ChordMajorFlatFifth
	ChordMajor
	ChordAugmented
	ChordSuspendedFourth
)

// ChordIdentity is a symbolic root and kind; formatting the native label is the caller's job.
type ChordIdentity struct {
	Root int
	Kind ChordKind
}

// CompoundCandidates is the raw tuple before retention and trial gates; Bass replaces Chords; not called while learning.
func (t *ScoreTable) CompoundCandidates(n CompoundNomination) [3]int {
	switch {
	case n.Bass:
		return t.bassCandidates(n)
	case n.Chords:
		return t.chordCandidates(n)
	}
	return [3]int{}
}

// chordCandidates cuts notes an octave above the first only after all three scans, with no rescan.
func (t *ScoreTable) chordCandidates(n CompoundNomination) [3]int {
	var found [3]int
	start := 36
	for slot := range found {
		// A scan from 109 reads fine record 0, whose foreground byte is never set: no note.
		if start > 108 {
			break
		}
		note := t.Automatic(start, n.Octave, n.PitchClass)
		if note == 0 {
			break
		}
		found[slot], start = note, note+1
	}
	for slot := 1; slot < 3; slot++ {
		if found[slot] >= found[0]+12 {
			found[slot] = 0
		}
	}
	return found
}

// bassCandidates validates base+12 and base+24 independently; a missing base still yields upper notes.
func (t *ScoreTable) bassCandidates(n CompoundNomination) [3]int {
	base := t.Ordinary(Nomination{
		Octave:              n.Octave,
		PitchClass:          n.PitchClass,
		Recording:           n.Recording,
		MeasureContinuously: n.MeasureContinuously,
	})
	out := [3]int{base}
	for i, enabled := range n.UpperOctaves {
		note := base + 12*(i+1)
		if !enabled || note > 108 {
			continue
		}
		validated := t.Ordinary(Nomination{
			Octave:              note / 12,
			PitchClass:          note % 12,
			Recording:           n.Recording,
			MeasureContinuously: n.PitchClass >= 0 && n.MeasureContinuously,
		})
		if validated != 0 && t.Scores[note] > t.Scores[base]/10 {
			out[i+1] = note
		}
	}
	return out
}

// RecognizeChord labels an ascending in-octave tuple, false meaning the native unknown label; the root never reorders it.
func RecognizeChord(notes [3]int) (ChordIdentity, bool) {
	if notes[0] == 0 || notes[1] == 0 || notes[2] == 0 {
		return ChordIdentity{}, false
	}
	intervals := [3]int{notes[1] - notes[0], notes[2] - notes[1], 12 - notes[2] + notes[0]}
	root := 0
	if chordKindOf(intervals[0], intervals[1]) == 0 {
		if intervals[0] >= 5 {
			root = 1
		}
		if intervals[(root+1)%3] >= 5 {
			root = 2
		}
	}
	kind := chordKindOf(intervals[root], intervals[(root+1)%3])
	if kind == 0 {
		return ChordIdentity{}, false
	}
	return ChordIdentity{Root: notes[root], Kind: kind}, true
}

func chordKindOf(first, second int) ChordKind {
	switch [2]int{first, second} {
	case [2]int{2, 5}:
		return ChordSuspendedSecond
	case [2]int{3, 3}:
		return ChordDiminished
	case [2]int{3, 4}:
		return ChordMinor
	case [2]int{4, 2}:
		return ChordMajorFlatFifth
	case [2]int{4, 3}:
		return ChordMajor
	case [2]int{4, 4}:
		return ChordAugmented
	case [2]int{5, 2}:
		return ChordSuspendedFourth
	}
	return 0
}

// BindBands gives each nonzero slot its own band; a zero slot stays the zero NoteBand.
func BindBands(notes [3]int, frequencies *NoteTable, fftSize, sampleRate int, rateFactor float64) [3]NoteBand {
	var out [3]NoteBand
	for slot, note := range notes {
		if note != 0 {
			out[slot] = NewNoteBand(frequencies[note-1], frequencies[note], frequencies[note+1], fftSize, sampleRate, rateFactor)
		}
	}
	return out
}
