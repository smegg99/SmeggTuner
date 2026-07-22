package dsp

import (
	"math"
	"time"

	"smegg.me/smeggtuner/core/tuning"
)

// Profiling: with a solo-rank register pulled, every partial in the octaves above belongs to the
// one reed sounding, so its amplitude over the fundamental can be measured and kept. That ratio is
// the reed's voice; compound tuning reads it back through profileFor to judge, by amplitude, the
// case phase cannot - a rank tuned dead onto the partial beats too slowly for anything to rotate.

// attachProfile measures the reading's own 2nd and 4th partials onto its one reed. Only a
// single-reed reading in profiling mode is measured: with more than one reed sounding no partial
// is unmistakably anyone's.
func (e *Engine) attachProfile(m *Measurement, zoom *Zoom, ring *Ring, window time.Duration) {
	if !e.cfg.ProfileHarmonics || len(m.Reeds) != 1 {
		return
	}
	f, amp := m.Reeds[0].Freq, reedAmp(m, zoom, ring, window)
	if f <= 0 || amp <= 0 {
		return
	}
	m.Reeds[0].Harmonics = []float64{
		e.partialRatio(zoom, ring, 2*f, amp, window),
		e.partialRatio(zoom, ring, 4*f, amp, window),
	}
}

// reedAmp is the reading's fundamental amplitude, measured on the same window the partials are, so
// the ratio owes nothing to the bellows.
func reedAmp(m *Measurement, zoom *Zoom, ring *Ring, window time.Duration) float64 {
	f := m.Reeds[0].Freq
	zr := zoom.Analyze(ring, f, math.Max(16, f*0.035), window)
	if !zr.Valid {
		return 0
	}
	var amp float64
	for _, p := range FindPeaks(zr, 1, 1) {
		amp = p.Amp
	}
	return amp
}

// partialRatio is the amplitude of the line at an exact multiple of the fundamental, over the
// fundamental's own; zero when nothing stands there. The line must sit within a lobe of the exact
// multiple - a reed's partials have no choice.
func (e *Engine) partialRatio(zoom *Zoom, ring *Ring, fk, fundAmp float64, window time.Duration) float64 {
	zr := zoom.Analyze(ring, fk, math.Max(16, fk*0.035), window)
	if !zr.Valid || fundAmp <= 0 {
		return 0
	}
	best := 0.0
	for _, p := range FindPeaks(zr, 3, e.lobeWidth()) {
		if math.Abs(p.Freq-fk) < e.lobeWidth() && p.Amp > best {
			best = p.Amp
		}
	}
	return best / fundAmp
}

// profileFor is the calibrated voice of the rank at offset sounding note, when the calibration
// sweep taught one. A ratio under profileMinRatio is no profile: dividing by a vanished harmonic
// would call anything at all a second voice.
func (e *Engine) profileFor(offset int, note tuning.Note, k int) (float64, bool) {
	for _, p := range e.cfg.Profiles {
		if p.Offset != offset || p.Note != note {
			continue
		}
		r := p.R2
		if k == 4 {
			r = p.R4
		}
		return r, r >= profileMinRatio
	}
	return 0, false
}
