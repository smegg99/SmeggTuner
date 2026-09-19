package meter

import (
	"math"
	"math/bits"
)

// watchSeconds are the short windows a watched note is read through; a note gets the shortest one
// holding watchCycles of its periods, so its band stays apart from the next semitone.
var watchSeconds = [...]float64{0.25, 0.5, 1}

const (
	watchCycles  = 100
	watchDecay   = 0.97 // per pass: a fade over seconds lowers the peak with it, a stop does not
	watchDrop    = 10   // a score under peak/watchDrop is a quiet pass
	watchSilence = 8    // quiet passes with nothing else sounding before a note is let go: longer than a bellows reversal
)

// watch notices that the measured note stopped sounding while the coarse window still holds it.
type watch struct {
	plans  [len(watchSeconds)]analyzerPlan
	hann   [len(watchSeconds)][]float32 // the reference windows are rectangular; this one keeps a semitone's leakage out of the band
	bands  [len(watchSeconds)][ActiveNotes]NoteBand
	note   int
	plan   int
	peak   float32
	quiet  int // quiet passes in a row with another note sounding
	silent int // quiet passes in a row with nothing else sounding
}

// newWatch builds the short windows; prev's plans and Hann tables are taken over where their sizes match.
func newWatch(cfg AnalyzerConfig, prev *watch) watch {
	var w watch
	for i, seconds := range watchSeconds {
		samples := int(math.Round(float64(cfg.SampleRate) * seconds))
		window := WindowPlan{SampleRate: cfg.SampleRate, Samples: samples, FFTSize: 4 << bits.Len(uint(samples-1))}
		var plan *analyzerPlan
		if prev != nil {
			plan = &prev.plans[i]
		}
		w.plans[i] = newAnalyzerPlan(cfg, window, plan)
		if prev != nil && len(prev.hann[i]) == samples {
			w.hann[i] = prev.hann[i]
		} else {
			w.hann[i] = make([]float32, samples)
			for k := range w.hann[i] {
				w.hann[i][k] = float32(0.5 - 0.5*math.Cos(2*math.Pi*float64(k)/float64(samples-1)))
			}
		}
		for k := range w.bands[i] {
			n := k + FirstActiveNote
			w.bands[i][k] = NewNoteBand(cfg.Table[n-1], cfg.Table[n], cfg.Table[n+1], window.FFTSize, cfg.SampleRate, cfg.RateFactor)
		}
	}
	return w
}

// watched is the note the last native event measured, the base of a register or Bass tuple, or while that
// event found none the note still locked; Chords are not watched.
func (a *Analyzer) watched() int {
	note := 0
	switch {
	case !a.last.ok || a.cfg.Chords:
		return 0
	case a.cfg.Bass:
		note = a.last.compound.Working[0]
	case len(a.cfg.Bands) >= 2:
		if len(a.last.notes) > 0 {
			note = a.last.notes[0]
		}
	default:
		note = a.last.frame.Selected
	}
	if note == 0 && a.History.Locked {
		return a.History.AcceptedNotes[0]
	}
	return note
}

// Gone reads the watched note's fundamental and twelfth through a short window and reports whether it has
// stayed under a tenth of its recent peak while another note sounds for cfg.Watch passes in a row, or with
// nothing else sounding for watchSilence passes. Call it once per pass.
func (a *Analyzer) Gone(ring *SampleRing) bool {
	w := &a.watch
	note := a.watched()
	if a.cfg.Watch == 0 || note == 0 {
		w.note = 0
		return false
	}
	if note != w.note {
		w.note, w.plan, w.peak, w.quiet, w.silent = note, len(watchSeconds)-1, 0, 0, 0
		for i, seconds := range watchSeconds {
			if seconds*a.cfg.Table[note] >= watchCycles {
				w.plan = i
				break
			}
		}
	}
	p := &w.plans[w.plan]
	if _, ok := p.window.Read(ring, p.input); !ok {
		return false
	}
	for k, h := range w.hann[w.plan] {
		p.input[k] *= h
	}
	spectrum := Magnitudes(p.spectrum, p.transform.Packed(p.spectrum, p.input), p.bounds)
	var score, other float32
	for k, b := range w.bands[w.plan] {
		v := BandPeakScore(spectrum, b.LowerBin, b.UpperBin)
		if n := k + FirstActiveNote; n == note || n == note+19 {
			score += v
		} else {
			other = max(other, v)
		}
	}
	w.peak = max(score, w.peak*watchDecay)
	switch {
	case score*watchDrop >= w.peak:
		w.quiet, w.silent = 0, 0
	case other*watchDrop >= w.peak:
		w.quiet++
	default:
		w.silent++
	}
	return w.quiet >= a.cfg.Watch || w.silent >= watchSilence
}

// LetGo is what a Gone note costs: the ring keeps only the short window it was judged on and the lock
// drops. The caller runs a native event at once; with Fresh it selects from that window.
func (a *Analyzer) LetGo(ring *SampleRing) {
	ring.Keep(a.watch.plans[a.watch.plan].window.Samples)
	a.History.Locked = false
	a.watch.note = 0
}
