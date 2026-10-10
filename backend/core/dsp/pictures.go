package dsp

import (
	"math"

	"smegg.me/smeggtuner/backend/core/dsp/meter"
)

// Picture sizes; pictures are drawn from the analysis and never feed it.
const (
	SpectrumColumns      = 256
	BandSpectrumColumns  = 64 // per band or chord tone in Spectra
	SpectrumCents        = 50.0
	SpectrumFloorDB      = 60.0
	WaveformPoints       = 256
	EqualizerCeilingDB   = 60.0
	EqualizerFullScaleDB = EqualizerCeilingDB
)

// equalizer is each note's coarse score over the background in dB, clamped to 0..EqualizerCeilingDB.
func equalizer(a *meter.Analyzer) []float32 {
	scores, background := a.Scores()
	out := make([]float32, len(scores))
	if background <= 0 {
		return out
	}
	for i, s := range scores {
		if s > background {
			out[i] = float32(min(EqualizerCeilingDB, 20*math.Log10(float64(s/background))))
		}
	}
	return out
}

// spectrumPicture is the fine spectrum across +-SpectrumCents around a note, as heights under the loudest column.
func spectrumPicture(a *meter.Analyzer, note int) []float32 {
	out, peak := spectrumColumns(a, note, SpectrumColumns)
	return heights(out, peak)
}

// bandPictures is one BandSpectrumColumns picture per band note on one shared scale; a missing band is flat.
func bandPictures(a *meter.Analyzer, notes []int) [][]float32 {
	out := make([][]float32, len(notes))
	var peak float32
	for i, note := range notes {
		out[i] = make([]float32, BandSpectrumColumns)
		if note > 0 {
			var p float32
			out[i], p = spectrumColumns(a, note, BandSpectrumColumns)
			peak = max(peak, p)
		}
	}
	for i := range out {
		heights(out[i], peak)
	}
	return out
}

// spectrumColumns is the largest fine magnitude in each of n columns across +-SpectrumCents, and their maximum.
func spectrumColumns(a *meter.Analyzer, note, n int) ([]float32, float32) {
	fine := a.Spectrum()
	center := a.ReferenceBin(note)
	out := make([]float32, n)
	edge := func(c float64) float64 {
		return center * math.Pow(2, (c/float64(n)*2-1)*SpectrumCents/1200)
	}
	var peak float32
	for c := range out {
		lo, hi := int(edge(float64(c))), int(math.Ceil(edge(float64(c+1))))
		for k := max(lo, 0); k <= hi && k < len(fine); k++ {
			out[c] = max(out[c], fine[k])
		}
		peak = max(peak, out[c])
	}
	return out, peak
}

// heights maps magnitudes in place onto 0..1 over SpectrumFloorDB below peak.
func heights(out []float32, peak float32) []float32 {
	for c, v := range out {
		if v <= 0 || peak <= 0 {
			out[c] = 0
			continue
		}
		out[c] = float32(max(0, 1+20*math.Log10(float64(v/peak))/SpectrumFloorDB))
	}
	return out
}
