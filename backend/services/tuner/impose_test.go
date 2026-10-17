package tuner

import (
	"slices"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// musette is a three-column goal from note 60 to 72: the first column flat, the third sharp.
func musette(t *testing.T) *target.Curve {
	t.Helper()
	c := target.NewCurve("Musette", 3)
	c.ExtrapolateLeft, c.ExtrapolateRight = false, false
	for _, note := range []tuning.Note{60, 72} {
		for col, dev := range []float64{-5, 0, 5} {
			if err := c.Set(note, col, dev); err != nil {
				t.Fatal(err)
			}
		}
	}
	return c
}

// slotTarget is the engine's target for a slot at a played key; the engine itself moves it onto the analysis note.
func slotTarget(c dsp.EngineConfig, slot int, key tuning.Note) (float64, bool) {
	i := int(key) - 12 - meter.FirstActiveNote
	return c.Targets[slot][i], c.TargetCoverage[i]
}

// A register short of the card reads each slot's own column at the played key.
func TestTheCurveBecomesTheEngineTargets(t *testing.T) {
	c := impose(dsp.EngineConfig{A4: 440, ReedCount: 1}, defaultPanel, imposed{
		a4: 440, reeds: 2, banks: "L,M2", cardBanks: "L,M1,M2", curve: musette(t),
	})
	if !slices.Equal(c.Octaves, []dsp.OctaveRequest{{Offset: -12, Reeds: 1}, {Offset: 0, Reeds: 1}}) || c.Bass {
		t.Fatalf("layout = %+v (bass %v), want a 16' band under an 8' one on the treble side", c.Octaves, c.Bass)
	}
	if len(c.Targets) != 2 {
		t.Fatalf("targets for %d slots, want the register's 2", len(c.Targets))
	}
	for _, key := range []tuning.Note{60, 66, 72} {
		low, covered := slotTarget(c, 0, key)
		high, _ := slotTarget(c, 1, key)
		if !covered || low != -5 || high != 5 {
			t.Fatalf("key %d = %v / %v (covered %v), want the L column's -5 and the M2 column's 5", key, low, high, covered)
		}
	}
	// Outside the anchors with extrapolation off the reference has no target, so neither does the engine.
	if v, covered := slotTarget(c, 0, 59); covered || v != meter.Missing {
		t.Fatalf("key 59 = %v (covered %v), want Missing and uncovered", v, covered)
	}
}

// Without a curve the engine gets no targets; facing the bass side the feet become the layout.
func TestABassBenchWithoutACurve(t *testing.T) {
	c := impose(dsp.EngineConfig{A4: 440, ReedCount: 1}, defaultPanel, imposed{a4: 440, reeds: 2, bassFeet: "16,8", cardFeet: "16,8,4"})
	if c.Targets != nil {
		t.Fatalf("targets = %d slots with no curve, want none", len(c.Targets))
	}
	if !c.Bass || !slices.Equal(c.Octaves, []dsp.OctaveRequest{{Offset: 0, Reeds: 1}, {Offset: 12, Reeds: 1}}) {
		t.Fatalf("layout = %+v (bass %v), want the 16' base and the 8' an octave up", c.Octaves, c.Bass)
	}
	if c.ReedCount != 2 {
		t.Fatalf("reed count = %d, want the ladder's 2", c.ReedCount)
	}
}

// An all-8' register stays one band of the register's reeds, and its targets follow the bank columns.
func TestAnEightFootRegisterIsOneBand(t *testing.T) {
	c := impose(dsp.EngineConfig{A4: 440, ReedCount: 1}, defaultPanel, imposed{
		a4: 440, reeds: 2, banks: "M1,M3", cardBanks: "M1,M2,M3", curve: musette(t),
	})
	if c.Octaves != nil || c.ReedCount != 2 || len(c.Targets) != 2 {
		t.Fatalf("layout = %+v, %d reeds, %d target slots, want one band of 2", c.Octaves, c.ReedCount, len(c.Targets))
	}
	if v, _ := slotTarget(c, 1, 60); v != 5 {
		t.Fatalf("slot 2 target = %v, want the M3 column's 5", v)
	}
}

// A unison register runs on the reference reed boxes of its banks: M2 and M3 of an MMM card are boxes 2 and 3.
func TestUnisonRegisterRunsOnItsReedBoxes(t *testing.T) {
	for _, tc := range []struct {
		card, banks string
		want        []int
		count       int
	}{
		{"M1,M2,M3", "M2,M3", []int{2, 3}, 2},
		{"L,M1,M2,M3", "M3", []int{3}, 1},
		{"M1,M2,M3", "M1,M2,M3", []int{1, 2, 3}, 3},
		{"M1,M2,M3,M4", "M3,M4", nil, 2},
		{"L,M1,M2", "L,M2", nil, 2},
	} {
		c := impose(dsp.EngineConfig{A4: 440, ReedCount: 1}, defaultPanel,
			imposed{a4: 440, reeds: len(splitBanks(tc.banks)), banks: tc.banks, cardBanks: tc.card})
		if !slices.Equal(c.Reeds, tc.want) || c.ReedCount != tc.count {
			t.Errorf("%s of %s = boxes %v / %d reeds, want %v / %d", tc.banks, tc.card, c.Reeds, c.ReedCount, tc.want, tc.count)
		}
	}
}

// A session decides over the panel and the app's own transpose: chords set with no session open do not reach its engine.
func TestASessionOverridesThePanel(t *testing.T) {
	pn := panel{mode: ModeChords, reeds: boxesOf([]int{1}), octaves: boxesOf([]int{2})}
	c := impose(dsp.EngineConfig{A4: 440, ReedCount: 1, Transpose: 4, ManualNote: 60}, pn, imposed{a4: 442, transpose: 7, reeds: 2})
	if c.Chords || c.ReedCount != 2 || c.Transpose != 7 || c.ManualNote != 60 || c.A4 != 442 {
		t.Fatalf("session engine = chords %v / %d reeds / T %d / note %d / A4 %v", c.Chords, c.ReedCount, c.Transpose, c.ManualNote, c.A4)
	}
}
