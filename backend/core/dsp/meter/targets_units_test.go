package meter

import (
	"math"
	"testing"
)

func targetsNominal(hz float64) *NoteCurve {
	var f NoteCurve
	for i := range f {
		f[i] = hz
	}
	return &f
}

func TestTargetHertzScaleIsWidenedBinary32(t *testing.T) {
	if TargetHertzScale != float64(float32(0.06)) || TargetHertzScale == 0.06 || math.Float64bits(TargetHertzScale) != 0x3faeb851e0000000 {
		t.Fatalf("TargetHertzScale = %v", TargetHertzScale)
	}
}

func TestSlotTargetsHertzInterpolatesBeforeConversion(t *testing.T) {
	nominal := targetsNominal(440)
	nominal[1], nominal[2] = 480, 520
	v, _ := BuildSlotTargets(targetsCurve(targetsPoint{4, 1}, targetsPoint{6, 3}), NoteCoverage{}, nominal)
	for i, raw := range []float64{1, 2, 3} {
		if want := (raw * 100) / (nominal[i] * TargetHertzScale); v[i] != want {
			t.Fatalf("note %d: %v, want %v", i+FirstActiveNote, v[i], want)
		}
	}
	if v[1] == (v[0]+v[2])/2 {
		t.Fatal("converted endpoints were interpolated")
	}
}

func TestSlotTargetsHertzConvertsBeforeRejection(t *testing.T) {
	v, _ := BuildSlotTargets(targetsCurve(targetsPoint{4, 100}), NoteCoverage{}, targetsNominal(10_000))
	if want := 10_000 / (10_000 * TargetHertzScale); v[0] != want || v[0] <= 0 || v[0] >= 50 {
		t.Fatalf("converted %v, want %v", v[0], want)
	}
}

func TestSlotTargetsHertzKeepsValuesBelowMissing(t *testing.T) {
	v, c := BuildSlotTargets(targetsCurve(targetsPoint{4, -1}), NoteCoverage{}, targetsNominal(10))
	if want := -100 / (10 * TargetHertzScale); v[0] != want || v[0] >= Missing || !c[0] {
		t.Fatalf("converted %v covered %v, want %v", v[0], c[0], want)
	}
}

func TestSlotTargetsHertzCanCollideWithMissing(t *testing.T) {
	scale := TargetHertzScale
	v, c := BuildSlotTargets(targetsCurve(targetsPoint{4, -98}), NoteCoverage{}, targetsNominal((100/scale)*(98.0/99)))
	if v[0] != Missing || !c[0] {
		t.Fatalf("converted %v covered %v", v[0], c[0])
	}
}

func TestTargetTablesReuseNominalForEverySlot(t *testing.T) {
	r := BuildTargets([]NoteCurve{
		targetsCurve(targetsPoint{57, 1}), targetsCurve(targetsPoint{57, 2}), targetsCurve(targetsPoint{57, 3}),
	}, targetsNominal(440))
	targets, _ := r.Note(57, 3)
	for slot, raw := range []float64{1, 2, 3} {
		if want := (raw * 100) / (440 * TargetHertzScale); targets[slot] != want {
			t.Fatalf("slot %d: %v, want %v", slot+1, targets[slot], want)
		}
	}
}

func TestSlotTargetsOutsideModelDomainArePlainArithmetic(t *testing.T) {
	v, c := BuildSlotTargets(targetsCurve(targetsPoint{4, math.MaxFloat64}), NoteCoverage{}, targetsNominal(440))
	if v[0] != Missing || !c[0] {
		t.Fatalf("overflowed target %v covered %v", v[0], c[0])
	}
	v, _ = BuildSlotTargets(targetsCurve(targetsPoint{4, 0}, targetsPoint{108, math.SmallestNonzeroFloat64 * 1e10}), NoteCoverage{}, nil)
	if v[0] != 0 || v[1] == 0 || v[1] >= 0x1p-1022 {
		t.Fatalf("subnormal span %v %v", v[0], v[1])
	}
}
