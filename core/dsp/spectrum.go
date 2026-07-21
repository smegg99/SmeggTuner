package dsp

import (
	"math"

	"smegg.me/smeggtuner/core/tuning"
)

// How far under the loudest line in the band the display reaches. A reed, its bellows sidebands and
// their noise live inside 60 dB; reaching deeper only draws deeper nulls.
const spectrumFloorDB = -SpectrumFloorDB

// How much of the reference height survives one fine result when the band has gone quieter (time
// constant near 0.6 s). The reference rises at once and falls slowly, so a dying note is not
// renormalised back up to full height.
const spectrumPeakRelease = 0.66

// spectrumFor lays the analysis band onto the columns a display draws: the +-SpectrumCents view
// around the note, in decibels under the loudest line, as heights in 0..1. The binning, log and
// smoothing happen here rather than in the UI so the frontend does not redo them on every repaint.
func (e *Engine) spectrumFor(zr ZoomResult, fc float64) []float32 {
	n := len(zr.Spec)
	if n == 0 || fc <= 0 {
		return nil
	}

	// A new note is a new band: a reference height held over would flatten a quieter one or lift a louder one off frame.
	if fc != e.specFc {
		e.specFc, e.specPeak = fc, 0
	}

	const step = 2 * SpectrumCents / SpectrumColumns
	var mag [SpectrumColumns]float64

	// Each column takes the loudest bin inside it, not the bin at its middle: a reed's lobe is a few
	// bins wide, and stepping over it would draw the reed short or miss it and invent one next door.
	for i, v := range zr.Spec {
		hz := zr.MinHz + float64(i)*zr.BinHz
		if hz <= 0 {
			continue
		}
		cents := tuning.Cents(hz, fc)
		if cents < -SpectrumCents || cents > SpectrumCents {
			continue
		}
		c := min(SpectrumColumns-1, int((cents+SpectrumCents)/step))
		if v > mag[c] {
			mag[c] = v
		}
	}

	// A band coarser than the columns leaves gaps; hold the last value across them so the trace stays
	// continuous without inventing a shape. The gaps are narrower than a pixel at any real width.
	var held, peak float64
	for c, v := range mag {
		if v > 0 {
			held = v
		}
		mag[c] = held
		peak = math.Max(peak, held)
	}

	if peak > e.specPeak {
		e.specPeak = peak
	} else {
		e.specPeak = e.specPeak*spectrumPeakRelease + peak*(1-spectrumPeakRelease)
	}

	// Decibels first, then the kernel: on a log scale the nulls between lines are cliffs to the floor,
	// so smoothing magnitudes would turn a clean reed into a comb of spikes. The kernel is short
	// enough that the reed's own lobe keeps its height.
	var db [SpectrumColumns]float64
	for c, v := range mag {
		db[c] = spectrumHeight(v, e.specPeak)
	}

	out := make([]float32, SpectrumColumns)
	last := SpectrumColumns - 1
	for c := range out {
		a := db[max(0, c-2)]
		b := db[max(0, c-1)]
		d := db[min(last, c+1)]
		f := db[min(last, c+2)]
		out[c] = float32((a + 2*b + 3*db[c] + 2*d + f) / 9)
	}
	return out
}

// spectrumHeight puts one magnitude on the frame: 1 at the loudest line in the band, 0 at the floor.
func spectrumHeight(mag, ref float64) float64 {
	if mag <= 0 || ref <= 0 {
		return 0
	}
	h := 1 - 20*math.Log10(mag/ref)/spectrumFloorDB
	return math.Min(1, math.Max(0, h))
}
