package dsp

// decimateWave thins a block to WaveformPoints into the engine's own buffer, allocating nothing.
// Each point is the largest-magnitude sample in its stride, kept with its sign; taking every nth
// sample would alias the reed's frequency. Points stay at full scale (-1..1) so the trace's height
// is the signal's level, not stretched to fill the panel.
func (e *Engine) decimateWave(samples []float32) {
	n := len(samples)
	if n == 0 {
		e.wave = [WaveformPoints]float32{}
		return
	}
	for i := range e.wave {
		lo := i * n / WaveformPoints
		hi := (i + 1) * n / WaveformPoints
		if hi <= lo {
			hi = lo + 1 // a block shorter than the trace: points repeat
		}
		if hi > n {
			hi = n
		}
		best := samples[lo]
		mag := abs32(best)
		for j := lo + 1; j < hi; j++ {
			if a := abs32(samples[j]); a > mag {
				best, mag = samples[j], a
			}
		}
		e.wave[i] = best
	}
}

// waveform copies the decimated block out for a measurement to carry: the run loop overwrites its
// buffer on the next block, and a measurement is read on another goroutine.
func (e *Engine) waveform() []float32 {
	out := make([]float32, WaveformPoints)
	copy(out, e.wave[:])
	return out
}

func abs32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
