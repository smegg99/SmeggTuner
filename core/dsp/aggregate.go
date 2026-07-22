package dsp

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"smegg.me/smeggtuner/core/tuning"
)

// Aggregate reduces a measurement stream from a recording to one robust result. It votes over the
// sustained region (State == StateRunning, InputLevel at least levelFrac of the 90th-percentile
// running level) rather than falling back to the last tick, which in a hand-pumped recording is the
// silent tail where detection latched onto room noise. Note is the majority; reeds are per-index medians of the majority note's
// ticks with the majority reed count; beats are per-pair medians; Locked is true if any selected tick
// locked. ok=false when the region is empty or has no note.
//
// Reedless state/level heartbeats never enter the note or reed vote but do set the level max behind
// the region threshold, so changing the engine's coarse cadence can move aggregated values slightly.
func Aggregate(ms []Measurement, a4 float64, levelFrac float64) (Measurement, bool) {
	// The level reference is the 90th percentile of the running ticks, not the loudest one: a
	// hand-pumped recording opens with an attack several times the sustained level, and a threshold
	// hung off that spike disqualifies the whole note (hohner16+8A4 in sounds/ does exactly this).
	var levels []float64
	for i := range ms {
		if ms[i].State == StateRunning {
			levels = append(levels, float64(ms[i].InputLevel))
		}
	}
	var ref float64
	if len(levels) > 0 {
		sort.Float64s(levels)
		ref = levels[len(levels)*9/10]
	}
	threshold := levelFrac * ref
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

	// Plurality vote among the ticks that measured reeds; ties break toward the note that reached
	// the winning count first. Reedless ticks vote only when nothing measured: they are pictures and
	// heartbeats, and in compound mode a picture names the sounding note while the reading names the
	// key - letting pictures vote would split one note between its two names.
	voters := region
	var measured []*Measurement
	for _, m := range region {
		if len(m.Reeds) > 0 {
			measured = append(measured, m)
		}
	}
	if len(measured) > 0 {
		voters = measured
	}
	noteVotes := make(map[tuning.Note]int)
	var note tuning.Note
	bestVotes := 0
	for _, m := range voters {
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
		Bands:          rep.Bands,
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
		// Each reed is read against its own octave's pitch; the band structure holds across the
		// selected ticks (same note, same reed count), so the representative's octave stands for all.
		oct := rep.Reeds[i].Octave
		out.Reeds = append(out.Reeds, ReedMeasure{
			Freq:     f,
			DevCents: tuning.Cents(f, bandPitch(fc, oct)),
			Octave:   oct,
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
		// A beat is heard where its upper reed sounds, so its cents are taken at that reed's octave.
		ref := bandPitch(fc, pairHiOctave(name, out.Reeds))
		out.Beats = append(out.Beats, BeatMeasure{
			Pair:  name,
			Hz:    hz,
			Cents: tuning.Cents(ref+hz, ref),
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

// bandPitch shifts the note's pitch to a reed's own octave.
func bandPitch(fc float64, octave int) float64 {
	if octave == 0 {
		return fc
	}
	return fc * math.Exp2(float64(octave)/12)
}

// pairHiOctave is the octave of a beat's upper reed, read from its "lo-hi" 1-based pair name; zero
// when the name or index does not resolve, which is every single-band measurement.
func pairHiOctave(pair string, reeds []ReedMeasure) int {
	_, after, ok := strings.Cut(pair, "-")
	if !ok {
		return 0
	}
	hi, err := strconv.Atoi(after)
	if err != nil || hi < 1 || hi > len(reeds) {
		return 0
	}
	return reeds[hi-1].Octave
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
