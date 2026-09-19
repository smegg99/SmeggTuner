package meter

import "testing"

func selPanics(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s: expected a panic", name)
		}
	}()
	f()
}

func TestBandPeakScore(t *testing.T) {
	cases := []struct {
		values       []float32
		lower, upper int
		want         float32
	}{
		{[]float32{0, 1, 5, 0}, 1, 2, 5},
		{[]float32{0, 5, 0}, 1, 1, 5},
		{[]float32{0, 5, 5, 0}, 1, 2, 5},
		{[]float32{0, 5, 5, 0}, 2, 2, 0},
		{[]float32{0, 1, 2, 3}, 1, 2, 0},
		{[]float32{0, 2, 0, 11, 0}, 1, 3, 11},
	}
	for _, c := range cases {
		if got := BandPeakScore(c.values, c.lower, c.upper); got != c.want {
			t.Errorf("BandPeakScore(%v, %d, %d) = %v, want %v", c.values, c.lower, c.upper, got, c.want)
		}
	}
	selPanics(t, "missing upper neighbour", func() { BandPeakScore([]float32{0, 5}, 1, 1) })
}

func TestAutomaticScan(t *testing.T) {
	auto := func(s *ScoreTable) int { return s.Automatic(0, -1, -1) }
	if auto(fixtureScores(nil, nil)) != 0 || auto(fixtureScores(map[int]float32{57: 100}, []int{})) != 0 {
		t.Fatal("no score or no foreground selects nothing")
	}
	if got := auto(fixtureScores(map[int]float32{20: 10, 60: 10}, nil)); got != 20 {
		t.Fatalf("equal scores keep the lower note: %d", got)
	}
	for strength, want := range map[float32]int{299: 20, 300: 20, 301: 60} {
		if got := auto(fixtureScores(map[int]float32{20: 10, 60: strength}, nil)); got != want {
			t.Errorf("30x replacement is strict: %v -> %d", strength, got)
		}
	}
	if got := auto(fixtureScores(map[int]float32{20: 1, 40: 20, 60: 31}, nil)); got != 60 {
		t.Fatalf("threshold follows the last accepted score: %d", got)
	}
	if got := auto(fixtureScores(map[int]float32{20: 1, 40: 31, 60: 900}, nil)); got != 40 {
		t.Fatalf("threshold follows the last accepted score: %d", got)
	}
	if got := auto(fixtureScores(map[int]float32{20: 1000, 60: 1}, []int{60})); got != 60 {
		t.Fatalf("a rejected candidate leaves the threshold: %d", got)
	}
	for _, neighbour := range []int{56, 58} {
		for center, want := range map[float32]int{9: 0, 10: 0, 11: 57} {
			if got := auto(fixtureScores(map[int]float32{57: center, neighbour: 5}, []int{57})); got != want {
				t.Errorf("2x neighbour %d is strict: %v -> %d", neighbour, center, got)
			}
		}
	}
	if got := auto(fixtureScores(map[int]float32{57: 100, 58: 60}, []int{57})); got != 0 {
		t.Fatalf("ineligible neighbours still compete: %d", got)
	}
}

func TestAutomaticRangeAndConstraints(t *testing.T) {
	auto := func(s *ScoreTable) int { return s.Automatic(0, -1, -1) }
	if got := auto(fixtureScores(map[int]float32{3: 100, 4: 201}, nil)); got != 4 {
		t.Fatalf("record 3 is the lower neighbour of note 4: %d", got)
	}
	if got := auto(fixtureScores(map[int]float32{108: 1}, nil)); got != 108 {
		t.Fatalf("note 108 is in range: %d", got)
	}
	if got := auto(fixtureScores(map[int]float32{108: 2, 109: 1}, []int{108})); got != 0 {
		t.Fatalf("score 109 is the upper neighbour of note 108: %d", got)
	}
	starts := fixtureScores(map[int]float32{4: 1, 20: 1}, nil)
	for start, want := range map[int]int{-100: 4, 20: 20, 21: 0} {
		if got := starts.Automatic(start, -1, -1); got != want {
			t.Errorf("start %d is inclusive and clamped to 4: %d", start, got)
		}
	}
	selPanics(t, "start above 108", func() { starts.Automatic(109, -1, -1) })
	notes := fixtureScores(map[int]float32{45: 1, 57: 1, 60: 1}, nil)
	for _, c := range []struct{ octave, pitch, want int }{{4, -1, 57}, {5, -1, 60}, {-1, 9, 45}, {4, 9, 57}} {
		if got := notes.Automatic(0, c.octave, c.pitch); got != c.want {
			t.Errorf("octave %d pitch %d: %d, want %d", c.octave, c.pitch, got, c.want)
		}
	}
}

func TestAutomaticRoundsTheProductToBinary32(t *testing.T) {
	low, high := float32(1.0000001192092896), float32(30.000003814697266)
	if float64(high) <= 30*float64(low) {
		t.Fatal("high must exceed the unrounded product")
	}
	if got := fixtureScores(map[int]float32{20: low, 60: high}, nil).Automatic(0, -1, -1); got != 20 {
		t.Fatalf("high equals the rounded 30x product and must not replace: %d", got)
	}
}

func TestOrdinaryDispatch(t *testing.T) {
	strong := map[int]float32{20: 1000, 57: 1}
	locked := func(note int, edit func(*Nomination)) Nomination {
		d := fixtureUnset()
		d.Locked, d.AcceptedNote = true, note
		if edit != nil {
			edit(&d)
		}
		return d
	}
	manual := func(d Nomination) Nomination { d.Octave, d.PitchClass = 4, 9; return d }
	if got := fixtureScores(map[int]float32{20: 1, 57: 20}, nil).Ordinary(fixtureUnset()); got != 20 {
		t.Fatalf("unlocked dispatch scans: %d", got)
	}
	if got := fixtureScores(strong, nil).Ordinary(locked(57, nil)); got != 57 {
		t.Fatalf("a valid lock beats a stronger competitor: %d", got)
	}
	failing := fixtureScores(strong, []int{20})
	for _, d := range []Nomination{locked(57, nil), locked(57, func(d *Nomination) { d.MeasureContinuously = true }), locked(0, nil)} {
		if got := failing.Ordinary(d); got != 0 {
			t.Errorf("a failed lock never rescans or bypasses: %+v -> %d", d, got)
		}
	}
	if got := fixtureScores(strong, nil).Ordinary(manual(locked(20, nil))); got != 57 {
		t.Fatalf("manual wins over lock: %d", got)
	}
	if got := failing.Ordinary(manual(fixtureUnset())); got != 0 {
		t.Fatalf("a failed manual note never rescans: %d", got)
	}
	if got := fixtureScores(map[int]float32{57: 1}, nil).Ordinary(locked(57, func(d *Nomination) { d.Octave = 3 })); got != 57 {
		t.Fatalf("lock ignores the octave constraint: %d", got)
	}
}

func TestOrdinaryManualBypassAndIdentity(t *testing.T) {
	empty := fixtureScores(nil, nil)
	d := Nomination{Octave: 4, PitchClass: 9, MeasureContinuously: true}
	if got := empty.Ordinary(d); got != 57 {
		t.Fatalf("MeasureContinuously bypasses validation while not recording: %d", got)
	}
	d.Recording = true
	if got := empty.Ordinary(d); got != 0 {
		t.Errorf("recording validates the manual note: %d", got)
	}
	if got := fixtureScores(map[int]float32{57: 100}, nil).Ordinary(Nomination{}); got != 0 {
		t.Fatalf("manual identity zero is kept: %d", got)
	}
	if got := empty.Ordinary(Nomination{Octave: -1, PitchClass: 9}); got != -3 {
		t.Fatalf("a nonpositive manual identity is returned unvalidated: %d", got)
	}
}
