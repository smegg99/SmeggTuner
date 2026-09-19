package meter

import (
	"math"
	"slices"
	"testing"
)

type targetsPoint struct {
	note  int
	value float64
}

func targetsCurve(points ...targetsPoint) NoteCurve {
	var c NoteCurve
	for i := range c {
		c[i] = Missing
	}
	for _, p := range points {
		c[p.note-FirstActiveNote] = p.value
	}
	return c
}

func targetsCovered(c NoteCoverage) int {
	n := 0
	for _, f := range c {
		if f {
			n++
		}
	}
	return n
}

func targetsExpect(t *testing.T, got []float64, want ...float64) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("targets %v, want %v", got, want)
	}
}

func TestSlotTargetsEmptyCurve(t *testing.T) {
	v, c := BuildSlotTargets(targetsCurve(), NoteCoverage{}, nil)
	if v != targetsCurve() || c != (NoteCoverage{}) {
		t.Fatalf("empty curve built %v %v", v, c)
	}
}

func TestSlotTargetsSingletonHasNoExtrapolation(t *testing.T) {
	v, c := BuildSlotTargets(targetsCurve(targetsPoint{57, 3}), NoteCoverage{}, nil)
	if v != targetsCurve(targetsPoint{57, 3}) || targetsCovered(c) != 1 || !c[57-FirstActiveNote] {
		t.Fatalf("singleton built %v %v", v, c)
	}
}

func TestSlotTargetsEndpointsAreInclusive(t *testing.T) {
	v, c := BuildSlotTargets(targetsCurve(targetsPoint{4, -52}, targetsPoint{108, 52}), NoteCoverage{}, nil)
	targetsExpect(t, v[:3], -52, -51, -50)
	targetsExpect(t, v[ActiveNotes-3:], 50, Missing, Missing)
	if targetsCovered(c) != ActiveNotes {
		t.Fatalf("coverage %v", c)
	}
}

func TestSlotTargetsMissingAndLowerAreNotAnchors(t *testing.T) {
	controls := targetsCurve(targetsPoint{4, -100}, targetsPoint{5, -99}, targetsPoint{7, -98}, targetsPoint{9, -96})
	controls[20] = math.NaN()
	v, c := BuildSlotTargets(controls, NoteCoverage{}, nil)
	targetsExpect(t, v[:7], -99, -99, -99, -98, -97, -96, -99)
	if targetsCovered(c) != 3 || v[20] != Missing {
		t.Fatalf("coverage %v, NaN slot %v", c, v[20])
	}
}

func TestSlotTargetsInterpolateRawControlsBeforeRejection(t *testing.T) {
	v, c := BuildSlotTargets(targetsCurve(targetsPoint{4, 0}, targetsPoint{8, 100}, targetsPoint{12, 0}), NoteCoverage{}, nil)
	targetsExpect(t, v[:9], 0, 25, 50, -99, -99, -99, 50, 25, 0)
	if targetsCovered(c) != 9 {
		t.Fatalf("coverage %v", c)
	}
}

func TestSlotTargetsUpperThresholdIsStrictAndOneSided(t *testing.T) {
	for _, c := range []struct{ control, want float64 }{
		{50, 50}, {math.Nextafter(50, math.Inf(1)), Missing}, {-98, -98}, {-60, -60},
	} {
		if v, _ := BuildSlotTargets(targetsCurve(targetsPoint{4, c.control}), NoteCoverage{}, nil); v[0] != c.want {
			t.Fatalf("control %v built %v, want %v", c.control, v[0], c.want)
		}
	}
}

func TestSlotTargetsFinalEndpointIsNotPinned(t *testing.T) {
	v, _ := BuildSlotTargets(targetsCurve(targetsPoint{4, 0}, targetsPoint{14, 1}), NoteCoverage{}, nil)
	if v[10] != 0.9999999999999999 {
		t.Fatalf("final endpoint %v", v[10])
	}
	v, _ = BuildSlotTargets(targetsCurve(targetsPoint{4, 0}, targetsPoint{14, 1}, targetsPoint{15, 2}), NoteCoverage{}, nil)
	targetsExpect(t, v[10:12], 1, 2)
}

func TestSlotTargetsPreserveCoverageOutsideSpans(t *testing.T) {
	var coverage NoteCoverage
	coverage[ActiveNotes-1] = true
	v, c := BuildSlotTargets(targetsCurve(targetsPoint{4, 80}), coverage, nil)
	if v != targetsCurve() || !c[0] || !c[ActiveNotes-1] || targetsCovered(c) != 2 || coverage[0] {
		t.Fatalf("built %v %v, input %v", v, c, coverage)
	}
}

func TestTargetTablesCoverageIsUnionOfRejectedSpans(t *testing.T) {
	r := BuildTargets([]NoteCurve{
		targetsCurve(targetsPoint{4, 100}, targetsPoint{6, 100}), targetsCurve(targetsPoint{8, 100}), targetsCurve(),
	}, nil)
	for slot, values := range r.Slots {
		if values != targetsCurve() {
			t.Fatalf("slot %d built %v", slot, values)
		}
	}
	if !slices.Equal(r.Coverage[:6], []bool{true, true, true, false, true, false}) || targetsCovered(r.Coverage) != 4 {
		t.Fatalf("coverage %v", r.Coverage)
	}
	if empty := BuildTargets([]NoteCurve{targetsCurve(), targetsCurve(), targetsCurve()}, nil); targetsCovered(empty.Coverage) != 0 {
		t.Fatalf("full rebuild kept coverage %v", empty.Coverage)
	}
}

func TestTargetTablesNoteIndexesNativeNotes(t *testing.T) {
	r := BuildTargets([]NoteCurve{
		targetsCurve(targetsPoint{4, 1}), targetsCurve(targetsPoint{108, 2}), targetsCurve(targetsPoint{108, 3}),
	}, nil)
	if targets, covered := r.Note(4, 3); !slices.Equal(targets, []float64{1, Missing, Missing}) || !covered {
		t.Fatalf("note 4: %v %v", targets, covered)
	}
	if targets, covered := r.Note(108, 3); !slices.Equal(targets, []float64{Missing, 2, 3}) || !covered {
		t.Fatalf("note 108: %v %v", targets, covered)
	}
	if _, covered := r.Note(57, 3); covered {
		t.Fatal("note 57 covered")
	}
}
