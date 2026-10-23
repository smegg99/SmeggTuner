package tests

import (
	"fmt"
	"math"
	"sort"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// minShare of analyses must name a note for it to be the reading; an attack may open on an analysis or two.
const minShare = 2.0 / 3

// heard is the note most analyses named, and the share of analyses that named it.
func heard(all []dsp.Measurement) (tuning.Note, float64) {
	count := map[tuning.Note]int{}
	var best tuning.Note
	frames := measured(all)
	for _, m := range frames {
		count[m.Note]++
		if count[m.Note] > count[best] || count[m.Note] == count[best] && m.Note < best {
			best = m.Note
		}
	}
	if len(frames) == 0 {
		return 0, 0
	}
	return best, float64(count[best]) / float64(len(frames))
}

// reads reports whether a run's reading is note.
func reads(all []dsp.Measurement, note tuning.Note) bool {
	got, share := heard(all)
	return got == note && share >= minShare
}

// sounding is the analyses on note at half the run's peak input level or more, which drops bellows reversals.
func sounding(all []dsp.Measurement, note tuning.Note) []dsp.Measurement {
	var peak float32
	for _, m := range all {
		peak = max(peak, m.InputLevel)
	}
	var out []dsp.Measurement
	for _, m := range measured(all, note) {
		if m.InputLevel >= peak/2 {
			out = append(out, m)
		}
	}
	return out
}

// sustained is a slot's median frequency over frames where it sounded, and how many frames that was.
func sustained(frames []dsp.Measurement, slot int) (float64, int) {
	var f []float64
	for _, m := range frames {
		if slot < len(m.Reeds) && m.Reeds[slot].Present {
			f = append(f, m.Reeds[slot].Freq)
		}
	}
	if len(f) == 0 {
		return 0, 0
	}
	sort.Float64s(f)
	return f[len(f)/2], len(f)
}

// cents is how far freq sits from nominal.
func cents(freq, nominal float64) float64 { return 1200 * math.Log2(freq/nominal) }

func name(n tuning.Note) string { return n.Name(tuning.NamingCDEFGAB) }

// slotsText prints a measurement's slots and beats for a log line.
func slotsText(m dsp.Measurement) string {
	out := ""
	for i, r := range m.Reeds {
		if !r.Present {
			out += fmt.Sprintf(" %d:-", i+1)
			continue
		}
		out += fmt.Sprintf(" %d:%.3fHz@%+d", i+1, r.Freq, r.Octave)
	}
	for _, b := range m.Beats {
		if b.Present {
			out += fmt.Sprintf(" %s:%.3fHz", b.Pair, b.Hz)
		}
	}
	return out
}

// summary condenses a run for a failure message: the measured notes in order, run-length encoded.
func summary(all []dsp.Measurement) string {
	out, last, n := "", tuning.Note(-1), 0
	flush := func() {
		if n > 0 {
			out += fmt.Sprintf(" %sx%d", name(last), n)
		}
	}
	for _, m := range measured(all) {
		if m.Note != last {
			flush()
			last, n = m.Note, 0
		}
		n++
	}
	flush()
	note, share := heard(all)
	return fmt.Sprintf("%s on %.0f%% of analyses:%s", name(note), 100*share, out)
}
