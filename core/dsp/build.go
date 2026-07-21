package dsp

import (
	"smegg.me/smeggtuner/core/tuning"
)

// hopState classifies a hop from its level alone, so tooQuiet is reachable in auto mode.
func hopState(level float64, clipped bool) EngineState {
	switch {
	case clipped:
		return StateTooLoud
	case level < quietLevel:
		return StateTooQuiet
	default:
		return StateRunning
	}
}

// publish emits m, or the latched measurement while frozen. Only fine results are latched: freezing
// must show the last real reading, not the reedless heartbeat that happened to arrive last.
func (e *Engine) publish(m Measurement) {
	if e.frozen && e.latched.State != "" {
		frozen := e.latched
		frozen.State = StateFrozen
		e.emit(frozen)
		return
	}
	if len(m.Reeds) > 0 {
		e.latched = m
	}
	e.emit(m)
}

// stateMeasurement is the lightweight tick: state, level, equalizer and waveform, with no reeds, beats or spectrum.
func (e *Engine) stateMeasurement(state EngineState, ce CoarseResult,
	floor [tuning.NumNotes]float64, level float64, srcAt float64) Measurement {
	return Measurement{
		State:      state,
		InputLevel: float32(level),
		Equalizer:  equalizerDB(ce, floor),
		Waveform:   e.waveform(),
		SourceAt:   srcAt,
	}
}

func (e *Engine) buildMeasurement(note tuning.Note, fc float64,
	freqs []float64, peaks []Peak, zr ZoomResult, ce CoarseResult,
	floor [tuning.NumNotes]float64, level float64, locked bool, srcAt float64) Measurement {

	m := Measurement{
		SourceAt:   srcAt,
		Note:       note,
		NoteName:   note.Name(e.cfg.Scale),
		Locked:     locked,
		ScalePitch: fc,
		InputLevel: float32(level),
		State:      StateRunning,
		Equalizer:  equalizerDB(ce, floor),
		Waveform:   e.waveform(),
		Reeds:      reedsAt(fc, freqs),
	}
	// Two lines closer than the window's main lobe are the shape of one, not two. A Hann window
	// spreads a single sinusoid over 4/T, and a wobbling reed makes that lobe lumpy, so the reeds have
	// to clear a whole lobe of each other to be reported apart and to have a beat taken from them.
	resLimit := e.lobeWidth()
	apart := len(freqs) > 0
	for i := 1; i < len(freqs); i++ {
		if freqs[i]-freqs[i-1] < resLimit {
			apart = false
		}
	}

	switch {
	case apart && len(freqs) == e.cfg.ReedCount && len(freqs) >= 2:
		// Every reed asked for, every one of them its own line.
		m.ReedsSeparated = true
		for i, b := range BeatsFromPeaks(peaksFromFreqs(freqs)) {
			m.Beats = append(m.Beats, BeatMeasure{
				Pair:         pairName(i),
				Hz:           b.Hz,
				Cents:        tuning.Cents(fc+b.Hz, fc),
				FromEnvelope: false,
			})
		}

	case e.cfg.ReedCount > 1:
		// Fewer lines than reeds asked for: either the reeds beat too slowly to pull apart (the
		// envelope still hears them and the peaks are lobe positions) or there are not that many reeds.
		hz, depth, ok := e.envelopeBeat(zr)
		if !ok {
			m.ReedsSeparated = apart
			break
		}
		fit, split := e.splitPair(zr, peaks, hz)
		if !split && depth < reedBeatDepth {
			// A swing too shallow to be two reeds on its own, and nothing else standing behind it.
			m.ReedsSeparated = apart
			break
		}
		m.ReedsSeparated = false
		if split {
			// The peak is the pair seen through too short a window; the beat knows the reeds better.
			lo := fit.Lo * (1 - e.cfg.ClockPPM*1e-6)
			hi := fit.Hi * (1 - e.cfg.ClockPPM*1e-6)
			m.Reeds = reedsAt(fc, []float64{lo, hi})
			m.ReedsFromBeat = true
		}
		beat := beatOf(fit, hz)
		m.Beats = append(m.Beats, BeatMeasure{
			Pair: "1-2", Hz: beat,
			Cents:        tuning.Cents(fc+beat, fc),
			FromEnvelope: true,
			Depth:        depth,
		})

	default:
		m.ReedsSeparated = apart
	}

	m.Spectrum = e.spectrumFor(zr, fc)
	return m
}

// reedsAt turns frequencies into reeds against the note's scale pitch. Nil in, nil out.
func reedsAt(fc float64, freqs []float64) []ReedMeasure {
	var out []ReedMeasure
	for _, f := range freqs {
		out = append(out, ReedMeasure{Freq: f, DevCents: tuning.Cents(f, fc)})
	}
	return out
}

func peaksFromFreqs(freqs []float64) []Peak {
	out := make([]Peak, len(freqs))
	for i, f := range freqs {
		out[i] = Peak{Freq: f}
	}
	return out
}

func pairName(i int) string {
	return string(rune('1'+i)) + "-" + string(rune('2'+i))
}
