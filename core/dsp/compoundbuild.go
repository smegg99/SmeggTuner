package dsp

import (
	"math"
	"time"

	"smegg.me/smeggtuner/core/tuning"
)

// compoundFine is one full-window compound analysis, ready to be measured and locked on.
type compoundFine struct {
	base  tuning.Note // the played key, untransposed: what the measurement is named after
	fc    float64     // the transposed base's scale pitch: ScalePitch and the band grid
	bands []compoundBand
	freqs []float64 // flat, bands ascending then lines ascending, clock-corrected
	specI int       // which band the picture is drawn from
}

// compoundAnalyze runs the compound stage for the tracked note: resolve which rank the note is,
// band and judge the register, and flatten the lines for the lock. The base resolution is cached
// while the note holds (re-checked once a second); a pinned note names the key itself, so it skips
// the resolution and the bands hang off it directly. now is seconds of samples consumed - the
// clock the residual-angle trackers run on.
func (e *Engine) compoundAnalyze(zoom *Zoom, ring *Ring, tracked tuning.Note, window time.Duration, now float64) (compoundFine, bool) {
	if e.compFor != tracked {
		// A new note is a new register position: the angle trackers may not carry across.
		for i := range e.compTracks {
			e.compTracks[i] = bandTrack{}
		}
		e.compFor, e.compBase, e.compAge = tracked, 0, 0
	}

	var base tuning.Note
	var bands []compoundBand

	switch {
	case e.cfg.ManualNote != 0:
		base = tracked
	case e.compBase != 0 && e.compAge < 12:
		e.compAge++
		base = e.compBase
	default:
		b, bs, ok := e.resolveBase(zoom, ring, tracked, window)
		if !ok {
			return compoundFine{}, false
		}
		base, bands = b, bs
		e.compBase, e.compAge = base, 0
	}

	fc := base.Transpose(e.cfg.Transpose).Freq(e.cfg.A4)
	if bands == nil {
		var ok bool
		bands, ok = e.compoundBands(zoom, ring, fc, window)
		if !ok {
			return compoundFine{}, false
		}
	}
	e.compoundVerdicts(bands, base, now)

	cf := compoundFine{base: base, fc: fc, bands: bands}
	for i, b := range bands {
		if b.req.Offset == 0 {
			cf.specI = i
		}
		for _, p := range b.genuine {
			cf.freqs = append(cf.freqs, p.Freq*(1-e.cfg.ClockPPM*1e-6))
		}
	}
	return cf, len(cf.freqs) > 0
}

// buildCompound turns one compound analysis into a Measurement: reeds against their own octave's
// pitch, beats within each band and across neighbouring bands - the beat a technician hears between
// a rank and the harmonic of the rank an octave under it.
func (e *Engine) buildCompound(cf compoundFine, ce CoarseResult,
	floor [tuning.NumNotes]float64, level float64, locked bool, srcAt float64) Measurement {

	m := Measurement{
		SourceAt:   srcAt,
		Note:       cf.base,
		NoteName:   cf.base.Name(e.cfg.Scale),
		Locked:     locked,
		ScalePitch: cf.fc,
		InputLevel: float32(level),
		State:      StateRunning,
		Equalizer:  equalizerDB(ce, floor),
		Waveform:   e.waveform(),
	}

	separated := true
	fi := 0
	prevLined := -1 // band index of the last band that put lines in the flat list
	for bi, b := range cf.bands {
		m.Bands = append(m.Bands, b.report)
		if b.report.Found < b.report.Ranks {
			separated = false
		}
		n := len(b.genuine)
		if n == 0 {
			continue
		}
		first := fi
		for _, p := range b.genuine {
			f := p.Freq * (1 - e.cfg.ClockPPM*1e-6)
			m.Reeds = append(m.Reeds, ReedMeasure{
				Freq:     f,
				DevCents: tuning.Cents(f, b.center),
				Octave:   b.req.Offset,
			})
			fi++
		}
		for k := first + 1; k < fi; k++ {
			hz := cf.freqs[k] - cf.freqs[k-1]
			m.Beats = append(m.Beats, BeatMeasure{
				Pair: pairName(k - 1), Hz: hz,
				Cents: tuning.Cents(b.center+hz, b.center),
			})
		}
		// A multi-rank band whose cluster would not split still beats; the envelope is the reading.
		if n < b.req.Reeds && b.envOK && b.envDepth >= reedBeatDepth {
			m.Beats = append(m.Beats, BeatMeasure{
				Pair: pairName(fi - 1), Hz: b.envHz,
				Cents:        tuning.Cents(b.center+b.envHz, b.center),
				FromEnvelope: true, Depth: b.envDepth,
			})
		}
		if prevLined >= 0 {
			if beat, ok := e.crossBeat(cf.bands[prevLined], cf.bands[bi], cf.freqs[first-1], cf.freqs[first]); ok {
				beat.Pair = pairName(first - 1)
				m.Beats = append(m.Beats, beat)
			}
		}
		prevLined = bi
	}
	m.ReedsSeparated = separated

	m.Spectrum = e.spectrumFor(cf.bands[cf.specI].zr, cf.bands[cf.specI].center)
	return m
}

// crossBeat is the beat between the last line of the band below and the first line of the band
// above: the line against the partial the lower rank lays beside it. Resolved apart, it is their
// distance. Merged into one line, it is read indirectly - off the band's amplitude envelope when
// the swing is fast enough to resolve, else off the residual's rotation (the verdict already
// measured it) - but only when the band holds one declared rank, because under several the
// indirect readings blend the octave beat with the band's own.
func (e *Engine) crossBeat(lo, hi compoundBand, fLo, fHi float64) (BeatMeasure, bool) {
	pred := fLo * math.Exp2(float64(hi.req.Offset-lo.req.Offset)/12)
	hz := fHi - pred
	if math.Abs(hz) >= e.lobeWidth() {
		return BeatMeasure{Hz: hz, Cents: tuning.Cents(hi.center+hz, hi.center)}, true
	}
	if hi.req.Reeds != 1 {
		return BeatMeasure{}, false
	}
	if ehz, depth, ok := e.envelopeBeat(hi.zr); ok {
		return BeatMeasure{
			Hz: ehz, Cents: tuning.Cents(hi.center+ehz, hi.center),
			FromEnvelope: true, Depth: depth,
		}, true
	}
	if hi.beatOK {
		return BeatMeasure{
			Hz: hi.beatHz, Cents: tuning.Cents(hi.center+hi.beatHz, hi.center),
			FromEnvelope: true,
		}, true
	}
	return BeatMeasure{}, false
}

// compoundShape is the report-twice gate's fingerprint of a compound result: what changed the shape
// must hold for two windows, including which base the note resolved to and each band's line count.
func compoundShape(cf compoundFine, beats int) reedShape {
	s := reedShape{reeds: len(cf.freqs), beats: beats, base: cf.base}
	for i, b := range cf.bands {
		if i >= len(s.found) {
			break
		}
		s.found[i] = int8(b.report.Found)
	}
	return s
}
