package dsp

import (
	"sort"

	"smegg.me/smeggtuner/core/tuning"
)

// Aggregate reduces a measurement stream from a recording to one robust result. It votes over the
// sustained region (State == StateRunning, InputLevel >= levelFrac*max tick level) rather than
// falling back to the last tick, which in a hand-pumped recording is the silent tail where detection
// latched onto room noise. Note is the majority; reeds are per-index medians of the majority note's
// ticks with the majority reed count; beats are per-pair medians; Locked is true if any selected tick
// locked. ok=false when the region is empty or has no note.
//
// Reedless state/level heartbeats never enter the note or reed vote but do set the level max behind
// the region threshold, so changing the engine's coarse cadence can move aggregated values slightly.
func Aggregate(ms []Measurement, a4 float64, levelFrac float64) (Measurement, bool) {
	var maxLevel float64
	for i := range ms {
		if ms[i].State != StateRunning {
			continue
		}
		if l := float64(ms[i].InputLevel); l > maxLevel {
			maxLevel = l
		}
	}
	threshold := levelFrac * maxLevel
	var region []*Measurement
	for i := range ms {
		m := &ms[i]
		if m.State != StateRunning || float64(m.InputLevel) < threshold {
			continue
		}
		region = append(region, m)
	}
	if len(region) == 0 {
		return Measurement{}, false
	}

	// plurality vote; ties break toward the note that reached the winning count first
	noteVotes := make(map[tuning.Note]int)
	var note tuning.Note
	bestVotes := 0
	for _, m := range region {
		if !m.Note.Valid() {
			continue
		}
		noteVotes[m.Note]++
		if noteVotes[m.Note] > bestVotes {
			bestVotes = noteVotes[m.Note]
			note = m.Note
		}
	}
	if note == 0 {
		return Measurement{}, false
	}
	var noteTicks []*Measurement
	for _, m := range region {
		if m.Note == note {
			noteTicks = append(noteTicks, m)
		}
	}

	// Majority reed count among the note's ticks. A tick with no reeds gets no vote: it never reached
	// the reeds (a picture drawn over too short a window, or a fine result that resolved no peak), and
	// there are more of those than readings, so letting them vote elects "no reeds".
	reedVotes := make(map[int]int)
	nReeds, bestVotes := 0, 0
	for _, m := range noteTicks {
		n := len(m.Reeds)
		if n == 0 {
			continue
		}
		reedVotes[n]++
		if reedVotes[n] > bestVotes {
			bestVotes = reedVotes[n]
			nReeds = n
		}
	}
	if nReeds == 0 {
		return Measurement{}, false // nothing here ever measured a reed
	}

	var sel []*Measurement
	for _, m := range noteTicks {
		if len(m.Reeds) == nReeds {
			sel = append(sel, m)
		}
	}

	fc := note.Freq(a4)
	rep := sel[len(sel)-1] // representative tick for non-aggregated fields
	out := Measurement{
		Note:     note,
		NoteName: rep.NoteName,
		State:    StateRunning,
		// These belong to the representative tick: the pitch the deviations are measured against and
		// how the reeds were told apart. A display drawing its reference from a zero, or reading a
		// merged pair as separate reeds, is worse than one showing nothing.
		ScalePitch:     rep.ScalePitch,
		ReedsSeparated: rep.ReedsSeparated,
		ReedsFromBeat:  rep.ReedsFromBeat,
		Equalizer:      rep.Equalizer,
		Spectrum:       rep.Spectrum,
	}
	if out.NoteName == "" {
		out.NoteName = note.Name(tuning.NamingCDEFGAB)
	}
	// Lock describes the sustained region only: the tracker can latch a steady room-noise line in the tail.
	for _, m := range sel {
		if m.Locked {
			out.Locked = true
			break
		}
	}

	vals := make([]float64, 0, len(sel))
	for i := 0; i < nReeds; i++ {
		vals = vals[:0]
		for _, m := range sel {
			vals = append(vals, m.Reeds[i].Freq)
		}
		f := median(vals)
		out.Reeds = append(out.Reeds, ReedMeasure{
			Freq:     f,
			DevCents: tuning.Cents(f, fc),
		})
	}

	type pairAcc struct {
		hz    []float64
		depth []float64
		env   int
	}
	pairs := make(map[string]*pairAcc)
	var order []string
	for _, m := range sel {
		for _, b := range m.Beats {
			p := pairs[b.Pair]
			if p == nil {
				p = &pairAcc{}
				pairs[b.Pair] = p
				order = append(order, b.Pair)
			}
			p.hz = append(p.hz, b.Hz)
			p.depth = append(p.depth, b.Depth)
			if b.FromEnvelope {
				p.env++
			}
		}
	}
	// Only the ticks that measured something can witness a beat; most of the region is the reedless heartbeat.
	var measuring int
	for _, m := range sel {
		if len(m.Reeds) > 0 {
			measuring++
		}
	}

	sort.Strings(order)
	for _, name := range order {
		p := pairs[name]
		// A beat most measurements did not see is not a beat: a note starting or dying swings its own amplitude.
		if 2*len(p.hz) <= measuring {
			continue
		}
		hz := median(p.hz)
		out.Beats = append(out.Beats, BeatMeasure{
			Pair:  name,
			Hz:    hz,
			Cents: tuning.Cents(fc+hz, fc),
			// strict majority; a tie keeps the primary spectral path
			FromEnvelope: 2*p.env > len(p.hz),
			Depth:        median(p.depth),
		})
	}

	vals = vals[:0]
	for _, m := range sel {
		vals = append(vals, float64(m.InputLevel))
	}
	out.InputLevel = float32(median(vals))
	return out, true
}

// median returns the middle value of vs, averaging the central pair for even sizes. vs is not modified.
func median(vs []float64) float64 {
	s := append([]float64(nil), vs...)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}
