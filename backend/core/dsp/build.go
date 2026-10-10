// backend/core/dsp/build.go
package dsp

import (
	"fmt"
	"strconv"

	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

var frameKinds = map[meter.FrameStatus]FrameKind{
	meter.FrameNoCoarse: FrameWaiting, meter.FrameNoNote: FrameNoNote, meter.FrameHeld: FrameHeld,
	meter.FrameNoFine: FrameWaiting, meter.FrameIneligible: FrameIneligible, meter.FrameMeasured: FrameMeasured,
}

// absentPanel stands where a layout has no readout, such as the pair of reeds 1 and 3.
var absentPanel = Panel{Value: meter.Missing, Target: meter.Missing, Decimals: 1, Dimmed: true, Shown: "-", Absent: true}

// heartbeat carries the meters and the current screen state between analyses.
func (r *runner) heartbeat(e *Engine, state EngineState) Measurement {
	a := r.analyzer
	ro := &a.Display.Readouts
	m := Measurement{State: state, Lock: Lock{Visible: a.Display.VisibleLock, Internal: a.History.Locked,
		Acquired: a.History.Acquired, Countdown: a.History.Countdown, Recorded: a.Display.Recorded}, Learning: a.Learning(),
		NeedlesDimmed: ro.FrequencyDimmed, InputLevel: r.level(), Waveform: e.waveform(), SourceAt: r.srcAt,
		Input: r.input.last, Unreliable: r.input.unreliable(),
		Reeds: []ReedMeasure{}, Beats: []BeatMeasure{}, Spectrum: []float32{}, Equalizer: []float32{}}
	for k, id := range r.ids {
		if k > 0 {
			p := absentPanel
			if r.pairs[k-1] && r.panels {
				p = panel(ro.Panels[2*id-1])
			}
			m.Panels = append(m.Panels, p)
		}
		m.Panels = append(m.Panels, panel(ro.Panels[2*id]))
		f := ro.Frequencies[id]
		m.Needles = append(m.Needles, Needle{Dev: f.Deviation, Freq: f.Frequency, Valid: f.Valid})
	}
	return m
}

func panel(p meter.Panel) Panel {
	return Panel{Value: p.Current, Target: p.Target, Signed: p.Signed, Decimals: p.Decimals, Dimmed: p.Dimmed,
		Text: p.Annotation, Shown: shown(p.Current, p.Decimals, p.Signed)}
}

// measurement is one analysis event: the screen state plus this event's fresh slot values.
func (r *runner) measurement(e *Engine, f meter.Frame) Measurement {
	m := r.heartbeat(e, r.state())
	m.Frame = frameKinds[f.Status]
	m.Equalizer = equalizer(r.analyzer)
	m.Taken = f.Taken
	if f.Note <= 0 {
		return m
	}
	_, keyAbove := r.cfg.bands()
	key := f.Note + keyAbove
	note := tuning.Note(key + 12)
	// A register can shift an analyzed note beyond the supported played-key range.
	if !note.Valid() {
		m.Frame, m.Taken = FrameNoNote, false
		m.Lock.Visible, m.Lock.Acquired, m.Lock.Recorded = false, false, false
		m.NeedlesDimmed = true
		for i := range m.Panels {
			m.Panels[i].Dimmed = true
		}
		return m
	}
	table := r.cfg.table()
	m.Note, m.ScalePitch = note, table[key]
	m.NoteName = m.Note.Name(r.cfg.Scale)
	if f.Status != meter.FrameMeasured && f.Status != meter.FrameIneligible {
		return m
	}
	m.Spectrum = spectrumPicture(r.analyzer, f.Note)
	if notes := r.bandNotes(f); notes != nil {
		m.Spectra = bandPictures(r.analyzer, notes)
	}
	octaves := r.cfg.slotOctaves()
	for k, id := range r.ids {
		reed := ReedMeasure{Reed: id + 1, Dev: meter.Missing, Octave: octaves[k]}
		if id < len(f.Refined.Bins) && f.Refined.Bins[id] > 0 {
			note := f.SlotNotes[id]
			reed.Present, reed.Dev, reed.Octave = true, f.Refined.Deviations[id], note-key
			reed.Freq = meter.FrequencyHz(table[note], reed.Dev)
		}
		m.Reeds = append(m.Reeds, reed)
	}
	for k, ok := range r.pairs {
		if ok {
			m.Beats = append(m.Beats, pairBeat(m.Reeds, k+1))
		}
	}
	if r.cfg.Chords {
		m.Label = r.analyzer.Display.Label.Text
	}
	return m
}

// bandNotes is the note of each band a chord or register measures, for one spectrum per tone; nil for one band.
func (r *runner) bandNotes(f meter.Frame) []int {
	switch {
	case r.cfg.Chords:
		return f.Notes
	case !r.cfg.spans():
		return nil
	case r.cfg.dirkBass():
		notes := make([]int, len(r.ids))
		for k, id := range r.ids {
			notes[k] = f.Notes[id]
		}
		return notes
	}
	return f.Notes
}

// AdjacentBeats is the beat in Hz of every neighbouring slot pair whose reed boxes are neighbours too.
func AdjacentBeats(reeds []ReedMeasure) []BeatMeasure {
	out := make([]BeatMeasure, 0, max(0, len(reeds)-1))
	for i := 1; i < len(reeds); i++ {
		if reeds[i].Reed-reeds[i-1].Reed <= 1 {
			out = append(out, pairBeat(reeds, i))
		}
	}
	return out
}

// pairBeat is slots i-1 and i by the reference's nearest-harmonic rule, named by their 1-based positions.
func pairBeat(reeds []ReedMeasure, i int) BeatMeasure {
	lo, hi := reeds[i-1], reeds[i]
	beat := BeatMeasure{Pair: fmt.Sprintf("%d-%d", i, i+1)}
	if lo.Present && hi.Present {
		beat.Present, beat.Hz = true, meter.HarmonicBeat(lo.Freq, hi.Freq, 0, true)
	}
	return beat
}

// shown prints a readout like the reference panel: an unsigned negative or a missing value is "-",
// otherwise the binary32-prequantized value with a leading "+" where the reference shows one.
func shown(value float64, decimals int, signed bool) string {
	if !signed && value < 0 || value <= meter.Missing {
		return "-"
	}
	q, plus := meter.QuantizeReadout(value, decimals, signed)
	text := strconv.FormatFloat(float64(q), 'f', decimals, 32)
	if plus && q > 0 {
		text = "+" + text
	}
	return text
}
