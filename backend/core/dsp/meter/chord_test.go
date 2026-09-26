package meter

import (
	"math"
	"testing"
)

func chordPick(t *ScoreTable, edit func(*CompoundNomination)) [3]int {
	n := CompoundNomination{Chords: true, UpperOctaves: [2]bool{true, true}, Octave: -1, PitchClass: -1}
	if edit != nil {
		edit(&n)
	}
	return t.CompoundCandidates(n)
}

func chordBassOnly(edit func(*CompoundNomination)) func(*CompoundNomination) {
	return func(n *CompoundNomination) {
		n.Chords, n.Bass = false, true
		if edit != nil {
			edit(n)
		}
	}
}

func chordManualBass(octave, pitch int) func(*CompoundNomination) {
	return chordBassOnly(func(n *CompoundNomination) { n.Octave, n.PitchClass, n.MeasureContinuously = octave, pitch, true })
}

func TestChordCandidates(t *testing.T) {
	cases := []struct {
		name       string
		values     map[int]float32
		foreground []int
		edit       func(*CompoundNomination)
		want       [3]int
	}{
		{"scans start at 36 and restart per slot", map[int]float32{8: 1000, 36: 1, 40: 30, 43: 30}, nil, nil, [3]int{36, 40, 43}},
		{"above 30x replaces the first", map[int]float32{36: 1, 40: 31, 43: 1}, nil, nil, [3]int{40, 43, 0}},
		{"octave cut runs after all scans", map[int]float32{36: 100, 40: 10, 60: 400}, nil, nil, [3]int{36, 0, 0}},
		{"octave cut without fallback", map[int]float32{36: 1, 47: 1, 48: 1}, []int{36, 48}, nil, [3]int{36, 0, 0}},
		{"eleven semitones stay", map[int]float32{36: 1, 47: 1}, nil, nil, [3]int{36, 47, 0}},
		{"twelve semitones drop", map[int]float32{36: 1, 40: 1, 48: 1}, nil, nil, [3]int{36, 40, 0}},
		{"octave constraint on every scan", map[int]float32{36: 1, 40: 1, 48: 1}, nil,
			func(n *CompoundNomination) { n.Octave = 3 }, [3]int{36, 40, 0}},
		{"note constraint on every scan", map[int]float32{36: 1, 48: 1}, nil,
			func(n *CompoundNomination) { n.Octave, n.PitchClass = 3, 0 }, [3]int{36, 0, 0}},
		{"pitch constraint on every scan", map[int]float32{36: 1, 40: 10, 48: 2}, nil,
			func(n *CompoundNomination) { n.PitchClass = 0 }, [3]int{36, 0, 0}},
		{"manual option does not nominate chords", map[int]float32{}, nil,
			func(n *CompoundNomination) { n.Octave, n.PitchClass, n.MeasureContinuously = 3, 0, true }, [3]int{}},
		{"empty", map[int]float32{}, nil, nil, [3]int{}},
		{"single", map[int]float32{57: 1}, nil, nil, [3]int{57, 0, 0}},
		{"third note 108 needs no scan", map[int]float32{96: 1, 100: 1, 108: 1}, nil, nil, [3]int{96, 100, 0}},
		{"continuation past 108 finds nothing", map[int]float32{100: 1, 108: 1}, nil, nil, [3]int{100, 108, 0}},
		{"first note 108", map[int]float32{108: 1}, nil, nil, [3]int{108, 0, 0}},
		{"neither mode", map[int]float32{57: 1}, nil, func(n *CompoundNomination) { n.Chords = false }, [3]int{}},
	}
	for _, c := range cases {
		if got := chordPick(fixtureScores(c.values, c.foreground), c.edit); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestBassCandidates(t *testing.T) {
	tenth := float32(0.1)
	cases := []struct {
		name       string
		values     map[int]float32
		foreground []int
		edit       func(*CompoundNomination)
		want       [3]int
	}{
		{"bass replaces chords", map[int]float32{8: 100, 20: 11, 32: 11, 36: 1}, nil,
			func(n *CompoundNomination) { n.Bass = true }, [3]int{8, 20, 32}},
		{"upper slots are independent", map[int]float32{8: 100, 20: 11, 32: 11, 36: 1}, nil,
			chordBassOnly(func(n *CompoundNomination) { n.UpperOctaves = [2]bool{false, true} }), [3]int{8, 0, 32}},
		{"upper notes skip the base octave constraint", map[int]float32{24: 100, 36: 11, 48: 11}, nil,
			chordBassOnly(func(n *CompoundNomination) { n.Octave = 2 }), [3]int{24, 36, 48}},
		{"failed manual base has no fallback but keeps tails", map[int]float32{12: 3, 24: 3}, nil,
			chordBassOnly(func(n *CompoundNomination) { n.Octave, n.PitchClass = 0, 8 }), [3]int{0, 12, 24}},
		{"zero base threshold uses score 0", map[int]float32{0: 30, 12: 3, 24: 4}, []int{12, 24},
			chordBassOnly(func(n *CompoundNomination) { n.Octave, n.PitchClass = 0, 8 }), [3]int{0, 0, 24}},
		{"binary32 tenth equality fails", map[int]float32{24: 1, 36: tenth, 48: math.Nextafter32(tenth, 1)}, []int{},
			chordManualBass(2, 0), [3]int{24, 0, 48}},
		{"manual option bypasses validation", map[int]float32{23: 5, 24: 10, 36: 2, 48: 2}, nil,
			chordManualBass(2, 0), [3]int{24, 36, 48}},
		{"recording disables the manual option", map[int]float32{23: 5, 24: 10, 36: 2, 48: 2}, nil,
			func(n *CompoundNomination) { chordManualBass(2, 0)(n); n.Recording = true }, [3]int{0, 0, 0}},
		{"neighbor equality fails", map[int]float32{24: 10, 35: 1, 36: 2, 48: 2}, nil,
			chordBassOnly(func(n *CompoundNomination) { n.Octave, n.PitchClass = 2, 0 }), [3]int{24, 0, 48}},
		{"automatic base keeps upper validation", map[int]float32{24: 100, 36: 11, 48: 11}, []int{24},
			chordBassOnly(func(n *CompoundNomination) { n.MeasureContinuously = true }), [3]int{24, 0, 0}},
		{"only the slot above 108 drops", map[int]float32{96: 10, 108: 2}, nil, chordBassOnly(nil), [3]int{96, 108, 0}},
		{"base 108 has no upper slots", map[int]float32{108: 10}, nil, chordBassOnly(nil), [3]int{108, 0, 0}},
		{"subnormal threshold is plain IEEE", map[int]float32{24: 0x1p-126, 36: 1}, nil,
			chordManualBass(2, 0), [3]int{24, 36, 0}},
	}
	for _, c := range cases {
		if got := chordPick(fixtureScores(c.values, c.foreground), c.edit); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestRecognizeChordDirectKinds(t *testing.T) {
	kinds := map[[2]int]ChordKind{
		{2, 5}: ChordSuspendedSecond, {3, 3}: ChordDiminished, {3, 4}: ChordMinor, {4, 2}: ChordMajorFlatFifth,
		{4, 3}: ChordMajor, {4, 4}: ChordAugmented, {5, 2}: ChordSuspendedFourth,
	}
	for pair, kind := range kinds {
		got, ok := RecognizeChord([3]int{36, 36 + pair[0], 36 + pair[0] + pair[1]})
		if !ok || got != (ChordIdentity{Root: 36, Kind: kind}) {
			t.Errorf("%v: got %v %v, want root 36 kind %v", pair, got, ok, kind)
		}
	}
}

func TestRecognizeChordInversionsAndAmbiguity(t *testing.T) {
	cases := []struct {
		notes [3]int
		want  ChordIdentity
	}{
		{[3]int{40, 43, 48}, ChordIdentity{48, ChordMajor}},
		{[3]int{43, 48, 52}, ChordIdentity{48, ChordMajor}},
		{[3]int{39, 43, 48}, ChordIdentity{48, ChordMinor}},
		{[3]int{43, 48, 51}, ChordIdentity{48, ChordMinor}},
		{[3]int{36, 38, 43}, ChordIdentity{36, ChordSuspendedSecond}},
		{[3]int{38, 43, 45}, ChordIdentity{38, ChordSuspendedFourth}},
	}
	for _, c := range cases {
		if got, ok := RecognizeChord(c.notes); !ok || got != c.want {
			t.Errorf("%v: got %v %v, want %v", c.notes, got, ok, c.want)
		}
	}
}

func TestRecognizeChordUnknown(t *testing.T) {
	for _, notes := range [][3]int{
		{0, 0, 0}, {36, 0, 0}, {36, 40, 0}, {36, 37, 38},
		{36, 0, 43}, {36, 40, 48}, {40, 36, 43}, {36, 36, 40}, {36, 48, 60},
	} {
		if got, ok := RecognizeChord(notes); ok {
			t.Errorf("%v: recognized %v, want unknown", notes, got)
		}
	}
}
