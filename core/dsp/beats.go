package dsp

import (
	"math"
	"sort"

	"gonum.org/v1/gonum/dsp/fourier"
)

// BeatEstimate is one beat (tremolo) frequency between two reeds.
type BeatEstimate struct {
	Hz           float64 // beat rate
	FromEnvelope bool    // true if measured from the amplitude envelope
}

// BeatsFromPeaks returns the adjacent-pair frequency differences of peaks (assumed sorted ascending).
// FromEnvelope is false for all entries.
func BeatsFromPeaks(peaks []Peak) []BeatEstimate {
	if len(peaks) < 2 {
		return nil
	}
	out := make([]BeatEstimate, 0, len(peaks)-1)
	for i := 1; i < len(peaks); i++ {
		out = append(out, BeatEstimate{Hz: peaks[i].Freq - peaks[i-1].Freq})
	}
	return out
}

// EnvelopeBeat measures the dominant amplitude-modulation frequency of |zr.TimeSeries| within
// [minHz, maxHz]: the fallback for reeds too close for FindPeaks to separate, where the merged peak
// still beats. The envelope is de-meaned, Hann-windowed, zero-padded x4 and transformed; the dominant
// in-band bin is refined by parabolic interpolation. ok is false when no bin clears 4x the in-band
// median or the modulation depth is negligible next to the carrier.
func EnvelopeBeat(zr ZoomResult, minHz, maxHz float64) (hz, depth float64, ok bool) {
	if !zr.Valid || len(zr.TimeSeries) < 64 || zr.Rate <= 0 || maxHz <= minHz {
		return 0, 0, false
	}
	n := len(zr.TimeSeries)
	env := make([]float64, n)
	var mean float64
	for i, c := range zr.TimeSeries {
		env[i] = math.Hypot(real(c), imag(c))
		mean += env[i]
	}
	mean /= float64(n)
	// remove DC before windowing: the envelope's large offset would otherwise swamp the sub-Hz bins
	hw := Hann(n)
	var wsum float64
	for i := range env {
		env[i] = (env[i] - mean) * hw[i]
		wsum += hw[i]
	}
	// magnitude the removed DC would have had at bin 0; a modulation of depth m lands at ~(m/2)*carrier
	carrier := mean * wsum

	nfft := 1
	for nfft < n*4 {
		nfft *= 2
	}
	padded := make([]float64, nfft)
	copy(padded, env)
	fft := fourier.NewFFT(nfft)
	coeffs := fft.Coefficients(nil, padded)
	binHz := zr.Rate / float64(nfft)

	lo := int(minHz / binHz)
	if lo < 1 {
		lo = 1
	}
	hi := int(maxHz / binHz)
	// keep bestBin+1 addressable for the parabolic refinement
	if hi > len(coeffs)-2 {
		hi = len(coeffs) - 2
	}
	if hi < lo {
		return 0, 0, false
	}
	mags := make([]float64, 0, hi-lo+1)
	bestBin, bestMag := -1, 0.0
	for k := lo; k <= hi; k++ {
		m := math.Hypot(real(coeffs[k]), imag(coeffs[k]))
		mags = append(mags, m)
		if m > bestMag {
			bestBin, bestMag = k, m
		}
	}
	if bestBin <= 0 || len(mags) < 8 {
		return 0, 0, false
	}
	sorted := append([]float64(nil), mags...)
	sort.Float64s(sorted)
	median := sorted[len(sorted)/2]
	// significance: 4x the in-band median AND at least ~2% depth. The depth floor matters: a flat
	// envelope has a near-zero median, so the ratio test alone would pass on rounding noise.
	if bestMag < 4*median || bestMag < 0.01*carrier {
		return 0, 0, false
	}
	la := math.Log(math.Hypot(real(coeffs[bestBin-1]), imag(coeffs[bestBin-1])) + 1e-30)
	lb := math.Log(bestMag + 1e-30)
	lc := math.Log(math.Hypot(real(coeffs[bestBin+1]), imag(coeffs[bestBin+1])) + 1e-30)
	den := la - 2*lb + lc
	var d float64
	if den != 0 {
		d = 0.5 * (la - lc) / den
	}
	// An AM sideband of depth m lands at ~(m/2)*carrier, so the depth is twice the sideband over the
	// carrier: two reeds beating swing it far, one reed on a moving bellows barely stirs it.
	depth = 2 * bestMag / carrier
	hz = (float64(bestBin) + d) * binHz
	// Interpolation can carry the answer out of the band the caller asked for; a caller excluding
	// bellows wobble by asking for nothing under 0.6 Hz means it.
	if hz < minHz || hz > maxHz {
		return 0, 0, false
	}

	// Refuse a HARMONIC of a beat too slow to report. The envelope of a pair is |cos|, which carries a
	// second harmonic at twice the beat: a pair under the floor has its fundamental outside the band
	// and its harmonic inside it. The fundamental of |cos| is stronger, so more energy at half the
	// answer than at the answer means the answer is a harmonic. Only an answer whose half is under the
	// floor is asked; a real 2 Hz beat halves to a legal 1 Hz and is never doubted.
	if half := hz / 2; half < minHz {
		if binMag(coeffs, half/binHz) > bestMag {
			return 0, 0, false
		}
	}
	return hz, depth, true
}

// binMag is the magnitude at a fractional bin, taken at the nearer of the two: a probe, not a measurement.
func binMag(coeffs []complex128, bin float64) float64 {
	k := int(math.Round(bin))
	if k < 0 || k >= len(coeffs) {
		return 0
	}
	return math.Hypot(real(coeffs[k]), imag(coeffs[k]))
}
