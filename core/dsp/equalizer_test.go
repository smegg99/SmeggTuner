package dsp

import (
	"math"
	"testing"

	"smegg.me/smeggtuner/core/tuning"
)

// uncalibratedFloor is what the file and synth paths run on: NewNoiseFloor(0) leaves every band at a
// degenerate 1e-6 and the leak guard is the only reference there is.
func uncalibratedFloor() [tuning.NumNotes]float64 {
	var floor [tuning.NumNotes]float64
	for i := range floor {
		floor[i] = 1e-6
	}
	return floor
}

// A tuner listening to nothing must draw nothing. The numbers are the silent tail of
// tests/fixtures/h-16.wav (loudest band 3e-4, under quietLevel); the equalizer used to divide that by
// leakGuardRatio of itself and draw a firm mountain range in the bass.
func TestTheEqualizerStaysDarkInSilence(t *testing.T) {
	var ce CoarseResult
	for i := range ce.NoteEnergy {
		ce.NoteEnergy[i] = 5e-5
	}
	// Hiss is not flat: it humps in the bass, and that hump is what got drawn.
	for i := 21; i <= 31; i++ {
		ce.NoteEnergy[i] = 3e-4
	}

	for i, db := range equalizerDB(ce, refFloorFor(ce, uncalibratedFloor())) {
		if db != 0 {
			t.Fatalf("band %d drew %.1f dB out of a silent room (loudest note energy 3e-4, "+
				"quietLevel is %v); the strip must stay dark", i, db, quietLevel)
		}
	}
}

// The band drawn tallest is the note that is sounding. h-16.wav is the file that broke it.
func TestTheEqualizerPeaksAtTheSoundingNote(t *testing.T) {
	const sounding = 43 // H3, tuning.MinNote + 43 == midi 59

	var ce CoarseResult
	for i := range ce.NoteEnergy {
		ce.NoteEnergy[i] = 1e-4
	}
	ce.NoteEnergy[sounding] = 0.09
	// the skirt a Goertzel leaves either side of a real reed
	ce.NoteEnergy[sounding-1] = 0.004
	ce.NoteEnergy[sounding+1] = 0.004

	eq := equalizerDB(ce, refFloorFor(ce, uncalibratedFloor()))

	best := 0
	for i, db := range eq {
		if db > eq[best] {
			best = i
		}
	}
	if best != sounding {
		t.Fatalf("tallest band %d (midi %d), want %d (midi %d)",
			best, best+int(tuning.MinNote), sounding, sounding+int(tuning.MinNote))
	}

	// and it reaches the top of the strip, not 47% of it
	if math.Abs(float64(eq[sounding])-EqualizerFullScaleDB) > 0.01 {
		t.Fatalf("the sounding note read %.2f dB, want full scale %.2f",
			eq[sounding], EqualizerFullScaleDB)
	}
}

// EqualizerFullScaleDB is mirrored in frontend/app/composables/useTuner.ts; nothing at runtime relates
// the two, so this catches leakGuardRatio moving out from under the display.
func TestEqualizerFullScaleIsSetByTheLeakGuard(t *testing.T) {
	want := 20 * math.Log10(1/leakGuardRatio)
	if math.Abs(EqualizerFullScaleDB-want) > 0.01 {
		t.Fatalf("EqualizerFullScaleDB is %v but the leak guard puts full scale at %v.\n"+
			"The note strip normalises against this number and keeps a copy in\n"+
			"frontend/app/composables/useTuner.ts. Move all three together.",
			EqualizerFullScaleDB, want)
	}
	if EqualizerFullScaleDB > EqualizerCeilingDB {
		t.Fatalf("full scale %v is above the wire's ceiling %v, so bars would be clipped in transit",
			EqualizerFullScaleDB, EqualizerCeilingDB)
	}
}

// quietLevel belongs to the DISPLAY: equalizerDB floors its reference at it so a silent room draws no
// bars. If it migrated into refFloorFor it would land in the detector's reference and move every threshold.
func TestTheDisplayFloorDoesNotReachTheDetector(t *testing.T) {
	var ce CoarseResult
	for i := range ce.NoteEnergy {
		ce.NoteEnergy[i] = 3e-4 // the same silent frame as above
	}

	for i, ref := range refFloorFor(ce, uncalibratedFloor()) {
		if ref >= quietLevel {
			t.Fatalf("the detector's reference at band %d is %v, at or above quietLevel %v: "+
				"the equalizer's silence floor has leaked into detection", i, ref, quietLevel)
		}
	}
}
