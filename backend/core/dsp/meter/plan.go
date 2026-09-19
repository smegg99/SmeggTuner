// backend/core/dsp/meter/plan.go
package meter

// analyzerPlan is one analysis window: its size, FFT, magnitude bounds and the buffers it reads into.
type analyzerPlan struct {
	window    WindowPlan
	transform *Transform
	bounds    MagnitudeBounds
	input     []float32
	spectrum  []float32
}

// newAnalyzerPlan builds the plan for w; prev's transform and buffers are taken over, cleared, when
// its FFT size matches, since the analyzer that owned them is discarded.
func newAnalyzerPlan(cfg AnalyzerConfig, w WindowPlan, prev *analyzerPlan) analyzerPlan {
	bounds, _ := NewMagnitudeBounds(w.FFTSize, cfg.SampleRate, cfg.RateFactor, &cfg.Table)
	p := analyzerPlan{window: w, bounds: bounds}
	if prev == nil || prev.transform == nil || prev.transform.size != w.FFTSize {
		p.transform, p.input, p.spectrum = NewTransform(w.FFTSize), make([]float32, w.FFTSize), make([]float32, w.FFTSize)
		return p
	}
	p.transform, p.input, p.spectrum = prev.transform, prev.input, prev.spectrum
	clear(p.transform.work)
	clear(p.transform.scratch)
	clear(p.input)
	clear(p.spectrum)
	return p
}
