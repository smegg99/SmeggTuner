package dsp

import (
	"math"

	"smegg.me/smeggtuner/core/tuning"
)

// refFloorFor is the floor the detector and equalizer both measure against: the learned one, but
// never below leakGuardRatio of the loudest note in the frame. A sounding reed leaks into every
// neighbouring band, and measured against ambient alone that leakage clears the gate; flooring the
// reference puts it back underneath. With calibration off the raw floor is a degenerate 1e-6 and the
// same guard keeps hiss from reading as music.
func refFloorFor(ce CoarseResult, floor [tuning.NumNotes]float64) [tuning.NumNotes]float64 {
	ref := floor
	var maxE float64
	for _, v := range ce.NoteEnergy {
		if v > maxE {
			maxE = v
		}
	}
	if g := maxE * leakGuardRatio; g > 0 {
		for i := range ref {
			if ref[i] < g {
				ref[i] = g
			}
		}
	}
	return ref
}

// equalizerDB turns the coarse energies into the note strip's bars.
//
// The reference is both relative and absolute. refFloorFor keeps it above leakGuardRatio of the
// loudest note in the frame, which is right for detection but wrong for a picture: in silence it
// divides the room's hiss by a fraction of itself and hands back a full-scale bar. So the display's
// own copy is also floored at quietLevel, the level the engine calls silence. That decision is made
// here, on the display's floor; the detector's reference is untouched.
func equalizerDB(ce CoarseResult, floor [tuning.NumNotes]float64) []float32 {
	out := make([]float32, tuning.NumNotes)
	for i := range ce.NoteEnergy {
		ref := floor[i]
		if ref < quietLevel {
			ref = quietLevel
		}
		db := 20 * math.Log10((ce.NoteEnergy[i]+1e-12)/(ref+1e-12))
		if db < 0 {
			db = 0
		}
		if db > EqualizerCeilingDB {
			db = EqualizerCeilingDB
		}
		out[i] = float32(db)
	}
	return out
}

func snrAt(e CoarseResult, floor [tuning.NumNotes]float64, n tuning.Note) float64 {
	i := int(n - tuning.MinNote)
	if i < 0 || i >= tuning.NumNotes {
		return 0
	}
	return e.NoteEnergy[i] / floor[i]
}
