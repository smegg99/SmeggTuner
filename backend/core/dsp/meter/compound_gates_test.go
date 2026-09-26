package meter

import "testing"

func TestCompoundHasNoOrdinaryGates(t *testing.T) {
	// 50 < d <= 58 passes the ordinary gate but not the compound cache gate.
	s := compoundStart(false)
	compoundConsume(&s, compoundSpectrum([4]float32{206, 1, 4, 2}), [3]int{57, 0, 0},
		map[int]CompoundRecord{57: compoundRec(206, 200, 8), 0: {}}, compoundOptions(nil))
	if d := s.Readouts.Panels[0].Current; d <= 50 || d > 58 || s.Readouts.Frequencies[0].Valid {
		t.Fatalf("d=%v valid=%v", d, s.Readouts.Frequencies[0].Valid)
	}
	// A sentinel-valued deviation from a positive peak keeps a valid cache valid.
	s = compoundStart(true)
	compoundConsume(&s, compoundSpectrum([4]float32{188, 1, 4, 1}), [3]int{57, 0, 0},
		map[int]CompoundRecord{57: compoundRec(188, 200, 8), 0: {}}, compoundOptions(nil))
	if s.Readouts.Panels[0].Current != -100 || !s.Readouts.Frequencies[0].Valid {
		t.Fatal("d=-100 must not clear validity")
	}
}

func TestCompoundResetKeepsSharedDim(t *testing.T) {
	s := compoundStart(true)
	compoundConsume(&s, nil, [3]int{57, 69, 81}, map[int]CompoundRecord{57: compoundRec(-1, 200, 8), 69: {}, 81: {}}, compoundOptions(nil))
	if !s.Readouts.FrequencyDimmed {
		t.Fatal("broad reset must not clear the shared frequency dim")
	}
}

func TestCompoundBothFlagsAndChordEnables(t *testing.T) {
	spec := compoundSpectrum([4]float32{201, 1, 4, 1}, [4]float32{401, 1, 5, 1}, [4]float32{801, 1, 3, 1})
	records := map[int]CompoundRecord{57: compoundRec(201, 200, 4), 69: compoundRec(401, 400, 5), 81: compoundRec(801, 800, 6), 0: compoundRec(0, 77, 9)}
	s := compoundStart(true)
	out := compoundConsume(&s, spec, [3]int{57, 69, 0}, records, compoundOptions(func(o *CompoundOptions) { o.Chords, o.RequestedCount = true, 2 }))
	if !out.Call.Table || s.Readouts.Panels[1].Current == 11 {
		t.Fatalf("both flags: call %+v pair %v", out.Call, s.Readouts.Panels[1].Current)
	}
	s = compoundStart(true)
	compoundConsume(&s, spec, [3]int{57, 69, 81}, records, compoundOptions(func(o *CompoundOptions) { o.Bass, o.Chords, o.RequestedCount = false, true, 1 }))
	if s.Graph.Amplitudes != [3]float32{5, 5, 5} || s.Graph.ReferenceBins != [3]float64{200, 400, 800} {
		t.Fatalf("chords enable every graph slot: %+v", s.Graph)
	}
}

func TestCompoundZeroHistoryCallsAreReal(t *testing.T) {
	spec := compoundSpectrum([4]float32{201, 1, 4, 1}, [4]float32{401, 1, 5, 1})
	for _, c := range []struct {
		notes   [3]int
		records map[int]CompoundRecord
	}{
		{[3]int{57, 69, 81}, map[int]CompoundRecord{57: compoundRec(0, 200, 4), 69: compoundRec(401, 400, 5), 81: {}}},
		{[3]int{57, 69, 0}, map[int]CompoundRecord{57: compoundRec(201, 200, 4), 69: compoundRec(401, 400, 5), 0: {}}},
	} {
		s := compoundStart(true)
		s.History = History{LastNotes: [3]int{57, 69, 81}, AcceptedNotes: [3]int{57, 69, 81}, Acquired: true}
		compoundConsume(&s, spec, c.notes, c.records, compoundOptions(nil))
		if s.History.Countdown != 4 || s.History.LastNotes != [3]int{57, 69, 81} || !s.History.Acquired {
			t.Fatalf("%v: %+v", c.notes, s.History)
		}
	}
}

func TestCompoundLevelIncludesOutOfRangeSlots(t *testing.T) {
	spec := compoundSpectrum([4]float32{201, 1, 4, 1}, [4]float32{401, 1, 5, 1}, [4]float32{900, 1, 9, 1})
	records := map[int]CompoundRecord{57: compoundRec(201, 200, 4), 69: compoundRec(401, 400, 5), 81: compoundRec(900, 800, 6)}
	s := compoundStart(true)
	out := compoundConsume(&s, spec, [3]int{57, 69, 81}, records, compoundOptions(nil))
	if out.Slots[2].Deviation <= 50 || s.Level != 9 {
		t.Fatalf("d3=%v level=%v", out.Slots[2].Deviation, s.Level)
	}
}

func TestCompoundCentsPairUsesPrecedingReference(t *testing.T) {
	spec := compoundSpectrum([4]float32{201, 1, 4, 1}, [4]float32{213, 1, 4, 1})
	records := map[int]CompoundRecord{57: compoundRec(201, 200, 4), 58: compoundRec(213, 212, 5), 0: {}}
	s := compoundStart(true)
	out := compoundConsume(&s, spec, [3]int{57, 58, 0}, records, compoundOptions(func(o *CompoundOptions) { o.Hertz, o.RequestedCount = false, 2 }))
	f1, f2 := out.Slots[0].Frequency, out.Slots[1].Frequency
	want := HarmonicBeat(f1, f2, 440, false)
	if want == InvalidBeat || s.Readouts.Panels[1].Current != want || want == HarmonicBeat(f1, f2, compoundTable[58], false) {
		t.Fatalf("pair %v want %v", s.Readouts.Panels[1].Current, want)
	}
	if s.Readouts.Panels[1].Decimals != 1 {
		t.Fatal("cents pair precision is 1")
	}
}

func TestCompoundHistoryKeepsIdentityOfAbsentPeak(t *testing.T) {
	spec := compoundSpectrum([4]float32{201, 1, 4, 1}, [4]float32{401, 1, 5, 1})
	records := map[int]CompoundRecord{57: compoundRec(201, 200, 4), 69: compoundRec(401, 400, 5), 81: compoundRec(0, 800, 6)}
	s := compoundStart(true)
	out := compoundConsume(&s, spec, [3]int{57, 69, 81}, records, compoundOptions(func(o *CompoundOptions) { o.Bass, o.Chords = false, true }))
	if out.Call.Notes != [3]int{57, 69, 81} || out.Call.Deviations[2] != 0 {
		t.Fatalf("call %+v", out.Call)
	}
}
