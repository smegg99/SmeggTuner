package dsp

import (
	"math"

	"smegg.me/smeggtuner/core/tuning"
)

// NoiseFloor tracks a per-note ambient energy floor for SNR gating.
type NoiseFloor struct {
	floor     [tuning.NumNotes]float64
	calibLeft float64
}

// NewNoiseFloor with calibSeconds == 0 disables calibration (headless file/synth, where the input
// starts mid-note and folding the tone into the floor would kill detection); the mic path passes ~5.
func NewNoiseFloor(calibSeconds float64) *NoiseFloor {
	nf := &NoiseFloor{calibLeft: calibSeconds}
	for i := range nf.floor {
		nf.floor[i] = 1e-6
	}
	return nf
}

func (nf *NoiseFloor) Calibrating() bool { return nf.calibLeft > 0 }

func (nf *NoiseFloor) Update(e CoarseResult, dt float64) {
	if nf.calibLeft > 0 {
		for i, v := range e.NoteEnergy {
			if v*3 > nf.floor[i] {
				nf.floor[i] = v * 3
			}
		}
		nf.calibLeft -= dt
		// Snap the tail: dt like 0.1 is not exact in binary, so calibration ends after the promised seconds.
		if nf.calibLeft < 1e-9 {
			nf.calibLeft = 0
		}
		return
	}
	// tcSustained is slow on purpose: with the detector's 6x attack gate and the 2.5x hold gate, a
	// constant sound only crosses into background on the ~30 s scale the original tuner documents.
	const tcBackground = 10.0
	const tcSustained = 120.0
	aBg := 1 - math.Exp(-dt/tcBackground)
	aUp := 1 - math.Exp(-dt/tcSustained)
	for i, v := range e.NoteEnergy {
		if v < nf.floor[i]*4 {
			nf.floor[i] += (max(v*3, 1e-6) - nf.floor[i]) * aBg
		} else {
			nf.floor[i] += (v - nf.floor[i]) * aUp
		}
	}
}

func (nf *NoiseFloor) Floor() [tuning.NumNotes]float64 { return nf.floor }
