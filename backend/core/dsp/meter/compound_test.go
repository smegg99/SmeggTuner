package meter

import (
	"reflect"
	"slices"
	"testing"
)

var compoundTable = func() *NoteTable {
	t := EqualTemperament(440, 0)
	return &t
}()

// compoundConsume adapts note-keyed fixtures to the per-slot record array; absent keys are zero records.
type compoundFixture struct {
	Display
	History History
}

func compoundConsume(s *compoundFixture, spectrum []float32, notes [3]int, m map[int]CompoundRecord, o CompoundOptions) CompoundOutcome {
	var records [3]CompoundRecord
	for slot, note := range notes {
		records[slot] = m[note]
	}
	return ConsumeCompound(&s.Display, &s.History, spectrum, notes, records, o)
}

var compoundNames = [12]string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func compoundOptions(edit func(*CompoundOptions)) CompoundOptions {
	o := CompoundOptions{Bass: true, Hertz: true, RequestedCount: 3, Table: compoundTable,
		Annotate: func(note int) string { return compoundNames[note%12] }}
	if edit != nil {
		edit(&o)
	}
	return o
}

func compoundStart(valid bool) compoundFixture {
	var s compoundFixture
	s.Readouts = NewReadouts(3)
	for i := range s.Readouts.Panels {
		s.Readouts.Panels[i] = Panel{Current: 10 + float64(i), Target: 20 + float64(i), Decimals: 2,
			Enabled: i%2 == 1, Dimmed: true, Annotation: "old", Octave: 7}
	}
	for i := range s.Readouts.Frequencies {
		s.Readouts.Frequencies[i] = FrequencySlot{Deviation: 30, Frequency: 999, Position: 4, Motion: 5, Valid: valid}
	}
	s.Readouts.FrequencyDimmed, s.Readouts.Peaks = true, []int{9, 9, 9}
	s.Graph = CompoundGraph{Length: 4096, Amplitudes: [3]float32{1, 2, 3}, Peaks: [3]int{7, 8, 9},
		Offsets: [3]float64{.1, .2, .3}, ReferenceBins: [3]float64{11, 12, 13}}
	return s
}

// compoundSpectrum places (bin, left, centre, right) triples into a zero array.
func compoundSpectrum(peaks ...[4]float32) []float32 {
	values := make([]float32, 1200)
	for _, p := range peaks {
		i := int(p[0])
		values[i-1], values[i], values[i+1] = p[1], p[2], p[3]
	}
	return values
}

func compoundRec(peak int, bin float64, count int) CompoundRecord {
	return CompoundRecord{Peak: peak, ReferenceBin: bin, BinCount: count}
}

func TestCompoundSlotArithmeticAndPanel(t *testing.T) {
	s := compoundStart(true)
	out := compoundConsume(&s, compoundSpectrum([4]float32{203, 1.5, 4, 2.25}), [3]int{57, 0, 0},
		map[int]CompoundRecord{57: compoundRec(203, 200, 8), 0: compoundRec(0, 0, 0)}, compoundOptions(nil))
	d := DeviationUnits(203+Interpolate(1.5, 4, 2.25), 200)
	c := out.Slots[0]
	if c.Deviation != d || c.Amplitude != 4 || c.Reference != 440 || c.Frequency != FrequencyHz(440, d) {
		t.Fatalf("slot %+v", c)
	}
	p := s.Readouts.Panels[0]
	if p.Current != d || !p.Signed || p.Decimals != 1 || p.Dimmed || p.Annotation != "A" || p.Octave != 4 || p.Target != 20 || p.Enabled {
		t.Fatalf("panel %+v", p)
	}
	if !slices.Equal(s.Readouts.Peaks, []int{203, 0, 0}) {
		t.Fatalf("published peaks %v", s.Readouts.Peaks)
	}
}

func TestCompoundRecordZeroIsConsumed(t *testing.T) {
	alt := *compoundTable
	alt[0] = 300
	s := compoundStart(true)
	out := compoundConsume(&s, compoundSpectrum([4]float32{203, 1, 4, 1}, [4]float32{306, 1, 4, 1}), [3]int{57, 0, 0},
		map[int]CompoundRecord{57: compoundRec(203, 200, 8), 0: compoundRec(306, 300, 9)}, compoundOptions(func(o *CompoundOptions) {
			o.Table, o.Bass, o.Chords = &alt, false, true
		}))
	if !out.Slots[1].Computed || out.Slots[1].Reference != 300 || s.Readouts.Panels[2].Octave != 0 {
		t.Fatalf("record zero slot %+v", out.Slots[1])
	}
}

func TestCompoundCacheGate(t *testing.T) {
	for _, c := range []struct {
		peak  int
		valid bool
	}{{206, true}, {194, true}, {207, false}, {188, false}} {
		s := compoundStart(false)
		compoundConsume(&s, compoundSpectrum([4]float32{float32(c.peak), 1, 4, 1}), [3]int{57, 0, 0},
			map[int]CompoundRecord{57: compoundRec(c.peak, 200, 8), 0: {}}, compoundOptions(nil))
		f := s.Readouts.Frequencies[0]
		if f.Valid != c.valid || (c.valid && (f.Position != 0 || s.Readouts.FrequencyDimmed)) || (!c.valid && f.Frequency != 999) {
			t.Fatalf("peak %d cache %+v dim %v", c.peak, f, s.Readouts.FrequencyDimmed)
		}
	}
}

func TestCompoundResets(t *testing.T) {
	s := compoundStart(true)
	s.Readouts.FrequencyDimmed = false
	out := compoundConsume(&s, compoundSpectrum([4]float32{201, 1, 4, 1}), [3]int{57, 69, 0},
		map[int]CompoundRecord{57: compoundRec(201, 200, 8), 69: compoundRec(-1, 400, 8), 0: {}}, compoundOptions(nil))
	if !slices.Equal(out.Resets, []int{1}) || out.GraphPublished || out.Call.Table {
		t.Fatalf("outcome %+v", out)
	}
	if s.Readouts.Panels[0].Current != Missing || s.Readouts.Frequencies[0].Valid || s.Readouts.FrequencyDimmed {
		t.Fatal("broad reset must clear panels and validity but keep the shared dim")
	}
	if s.Readouts.Frequencies[0].Deviation != DeviationUnits(201, 200) || s.Graph.ReferenceBins != [3]float64{} || s.Graph.Peaks != [3]int{7, 8, 9} {
		t.Fatal("reset keeps numbers and other graph fields")
	}
}

func TestCompoundBassBeatsUseLocalFrequencies(t *testing.T) {
	records := map[int]CompoundRecord{33: compoundRec(275, 250, 8), 45: compoundRec(501, 500, 8), 57: compoundRec(1003, 1000, 8)}
	spec := compoundSpectrum([4]float32{275, 1, 4, 1}, [4]float32{501, 1, 4, 1}, [4]float32{1003, 1, 4, 1})
	s := compoundStart(true)
	out := compoundConsume(&s, spec, [3]int{33, 45, 57}, records, compoundOptions(nil))
	f := [3]float64{out.Slots[0].Frequency, out.Slots[1].Frequency, out.Slots[2].Frequency}
	if s.Readouts.Frequencies[0].Frequency != 999 {
		t.Fatal("out-of-range slot must keep its cache")
	}
	if s.Readouts.Panels[1].Current != HarmonicBeat(f[0], f[1], 110, true) || s.Readouts.Panels[3].Current != HarmonicBeat(f[1], f[2], 220, true) {
		t.Fatal("pairs must use local frequencies and the preceding reference")
	}
	p := s.Readouts.Panels[1]
	if p.Signed || p.Decimals != 2 || p.Dimmed || p.Annotation != "old" || p.Octave != 7 || p.Target != 21 {
		t.Fatalf("pair panel %+v", p)
	}
	s = compoundStart(true)
	compoundConsume(&s, spec, [3]int{33, 45, 57}, records, compoundOptions(func(o *CompoundOptions) { o.Bass, o.Chords = false, true }))
	if s.Readouts.Panels[1].Current != 11 || s.Readouts.Panels[3].Current != 13 {
		t.Fatal("chords must not write pairs")
	}
}

func TestCompoundHistoryAndGraph(t *testing.T) {
	spec := compoundSpectrum([4]float32{201, 1, 4, 1}, [4]float32{401, 1, 5, 1}, [4]float32{801, 1, 3, 1})
	records := map[int]CompoundRecord{57: compoundRec(201, 200, 4), 69: compoundRec(401, 400, 5), 81: compoundRec(801, 800, 6), 0: compoundRec(0, 77, 9)}
	cases := []struct {
		notes  [3]int
		edit   func(*CompoundOptions)
		table  bool
		counts [3]int
	}{
		{[3]int{57, 69, 81}, func(o *CompoundOptions) { o.Bass, o.Chords = false, true }, true, [3]int{4, 5, 6}},
		{[3]int{57, 69, 0}, func(o *CompoundOptions) { o.Bass, o.Chords = false, true }, false, [3]int{}},
		{[3]int{57, 69, 0}, func(o *CompoundOptions) { o.RequestedCount = 2 }, true, [3]int{4, 5, 9}},
		{[3]int{57, 69, 0}, nil, false, [3]int{}},
		{[3]int{57, 0, 0}, func(o *CompoundOptions) { o.RequestedCount = 2 }, false, [3]int{}},
	}
	for _, c := range cases {
		s := compoundStart(true)
		out := compoundConsume(&s, spec, c.notes, records, compoundOptions(c.edit))
		if out.Call.Table != c.table || out.Call.BinCounts != c.counts || !out.GraphPublished {
			t.Fatalf("%v: call %+v", c.notes, out.Call)
		}
	}
	s := compoundStart(true)
	s.Level, s.History.Inhibited = 4.5, true
	out := compoundConsume(&s, spec, [3]int{57, 69, 81}, records, compoundOptions(func(o *CompoundOptions) { o.RequestedCount = 2 }))
	if out.Call.Accepted || s.Level != 5 || s.Graph.Amplitudes != [3]float32{5, 5, 0} || s.Graph.ReferenceBins != [3]float64{200, 400, 0} || !s.Graph.Scratch {
		t.Fatalf("graph %+v level %v", s.Graph, s.Level)
	}
	if s.Graph.Peaks != [3]int{201, 401, 801} || s.Graph.Offsets != [3]float64{} {
		t.Fatalf("graph peaks/offsets %+v", s.Graph)
	}
	s = compoundStart(true)
	records[57] = compoundRec(0, 200, 4)
	out = compoundConsume(&s, spec, [3]int{57, 69, 81}, records, compoundOptions(nil))
	if out.GraphPublished || s.Graph.Length != 4096 || out.Call.Table || !out.Slots[1].Computed {
		t.Fatalf("zero first peak %+v", out)
	}
}

func TestCompoundFirstIdentityZero(t *testing.T) {
	s := compoundStart(true)
	before := compoundStart(true).Readouts
	s.History = History{LastNotes: [3]int{57, 69, 81}, AcceptedNotes: [3]int{57, 69, 81}, Locked: true, Acquired: true}
	out := compoundConsume(&s, nil, [3]int{0, 69, 81}, nil, compoundOptions(nil))
	if !reflect.DeepEqual(s.Readouts, before) || out.GraphPublished || out.Call != (HistoryCall{}) {
		t.Fatalf("outcome %+v", out)
	}
	if s.History.Countdown != 4 || !s.History.Acquired || !s.VisibleLock {
		t.Fatalf("history %+v visible %v", s.History, s.VisibleLock)
	}
}
