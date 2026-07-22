package dsp

import (
	"math"
	"math/cmplx"
	"time"

	"gonum.org/v1/gonum/dsp/fourier"
)

// ZoomResult is one fine-stage spectrum of a narrow band around Center.
type ZoomResult struct {
	Center     float64      // heterodyne center Hz
	Rate       float64      // decimated sample rate
	Spec       []float64    // magnitude spectrum of zero-padded FFT
	BinHz      float64      // Hz per spectrum bin
	MinHz      float64      // absolute Hz of Spec[0]
	TimeSeries []complex128 // decimated baseband signal (for envelope/phase)
	Phases     []float64    // phase per Spec bin
	Valid      bool         // false if ring had too little data
}

// Zoom heterodynes a narrow band to baseband and resolves it with a heavily zero-padded FFT, giving
// sub-0.05 Hz peak positions. Every intermediate buffer is held here and reused; only what a
// ZoomResult carries away is freshly made, because the engine keeps it.
type Zoom struct {
	sampleRate int
	ffts       map[int]*fourier.CmplxFFT
	scratch    []float64
	// The heterodyne writes into one of these and the decimator ping-pongs between them.
	bufA, bufB []complex128
	padded     []complex128
	coeffs     []complex128
	hann       []float64
}

func NewZoom(sampleRate int) *Zoom {
	return &Zoom{sampleRate: sampleRate, ffts: map[int]*fourier.CmplxFFT{}}
}

// Analyze heterodynes the newest window of samples to baseband at centerHz, low-pass filters and
// decimates so Rate >= 8*spanHz, Hann-windows, zero-pads x8, runs a complex FFT and returns the band
// [centerHz-spanHz, centerHz+spanHz].
func (z *Zoom) Analyze(ring *Ring, centerHz, spanHz float64, window time.Duration) ZoomResult {
	sr := float64(z.sampleRate)
	need := int(window.Seconds() * sr)
	if need <= 0 || spanHz <= 0 {
		return ZoomResult{Valid: false}
	}
	if len(z.scratch) < need {
		z.scratch = make([]float64, need)
	}
	got := ring.Tail(need, z.scratch[:need])
	if float64(got) < 0.9*float64(need) {
		return ZoomResult{Valid: false}
	}
	data := z.scratch[:got]

	// heterodyne to baseband
	a := grow(&z.bufA, got)
	b := grow(&z.bufB, got/2+1)
	heterodyne(data, a, -2*math.Pi*centerHz/sr)

	// decimate by 2 until the rate is just above 8*span, alternating between the two buffers: a pass
	// may not write into the one it is reading.
	bb := a
	inA := true
	rate := sr
	for rate/2 >= 8*spanHz && len(bb) > 512 {
		if inA {
			bb = lowpassDecimate2(bb, b[:0])
		} else {
			bb = lowpassDecimate2(bb, a[:0])
		}
		inA = !inA
		rate /= 2
	}

	n := len(bb)
	if n < 3 {
		return ZoomResult{Valid: false}
	}

	// TimeSeries outlives this call (the engine holds the previous result to refine phase against), so
	// it cannot be a view into a buffer the next hop overwrites. It is the one baseband copy made.
	series := make([]complex128, n)
	copy(series, bb)

	// Hann window + zero-pad x8 + FFT; the window is applied on the way into the FFT buffer so
	// TimeSeries keeps a clean envelope.
	if len(z.hann) != n {
		z.hann = Hann(n)
	}
	hw := z.hann

	nfft := 1
	for nfft < n*8 {
		nfft *= 2
	}
	padded := grow(&z.padded, nfft)
	clear(padded)
	for i, v := range series {
		padded[i] = v * complex(hw[i], 0)
	}
	fft, ok := z.ffts[nfft]
	if !ok {
		fft = fourier.NewCmplxFFT(nfft)
		z.ffts[nfft] = fft
	}
	coeffs := fft.Coefficients(grow(&z.coeffs, nfft), padded)

	binHz := rate / float64(nfft)
	// band [-span, +span] around center; FFT layout: bins 0..nfft/2-1 positive, nfft/2..nfft-1 negative
	bins := int(spanHz/binHz) + 1
	if bins > nfft/2-1 {
		bins = nfft/2 - 1
	}
	// Magnitudes are normalized by the window's sample count, so a line's amplitude does not depend
	// on how far this band's rate was decimated: bands zoomed at different centers compare honestly
	// (the compound stage's cross-band floor and the harmonic profiles both do).
	spec := make([]float64, 2*bins+1)
	phases := make([]float64, 2*bins+1)
	norm := 1 / float64(n)
	for k := -bins; k <= bins; k++ {
		idx := k
		if idx < 0 {
			idx += nfft
		}
		spec[k+bins] = cmplx.Abs(coeffs[idx]) * norm
		phases[k+bins] = cmplx.Phase(coeffs[idx])
	}
	return ZoomResult{
		Center:     centerHz,
		Rate:       rate,
		Spec:       spec,
		BinHz:      binHz,
		MinHz:      centerHz - float64(bins)*binHz,
		TimeSeries: series,
		Phases:     phases,
		Valid:      true,
	}
}
