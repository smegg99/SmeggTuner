package dsp

import (
	"math"
	"sort"
	"time"
)

// Octave analysis measures a compound register one octave at a time. The fine stage centres a single
// ~60 cent band on the played note and is blind to a 16' an octave below or a 4' an octave above; the
// zoom heterodynes to any centre, so the fix is to place a band where each foot sounds and resolve
// each on its own. This is the banding only; SubtractHarmonics does the harmonic bookkeeping.

// OctaveRequest asks for one octave of a compound register: where it sounds relative to the note, and
// how many lines to look for there. Several 8' ranks share Offset 0 and are told apart by Reeds > 1.
type OctaveRequest struct {
	Offset int // semitones from the note: -12 a 16', 0 an 8', +12 a 4'
	Reeds  int // lines to resolve in this octave (the ranks that share it)
}

// OctaveBand is one octave's result: the band that was placed, and the lines found in it.
type OctaveBand struct {
	Offset int     // the request's offset, so a caller can map a band back to its foot
	Center float64 // Hz the band was centred on: base * 2^(Offset/12)
	Reeds  []Peak  // lines found, strongest-refined, ascending in frequency
	Valid  bool    // false when the ring held too little to analyse this band
}

// AnalyzeOctaves resolves each octave of a compound register independently. Each request places a
// band at baseHz*2^(Offset/12) and finds up to Reeds lines in it, kept minSepHz apart, over the given
// window. Bands come back in request order. It owns no state and mutates nothing shared.
func AnalyzeOctaves(z *Zoom, ring *Ring, baseHz float64, reqs []OctaveRequest, minSepHz float64, window time.Duration) []OctaveBand {
	bands := make([]OctaveBand, 0, len(reqs))
	for _, req := range reqs {
		center := baseHz * math.Exp2(float64(req.Offset)/12)
		span := math.Max(16, center*0.035)
		band := OctaveBand{Offset: req.Offset, Center: center}

		reeds := req.Reeds
		if reeds < 1 {
			reeds = 1
		}
		if zr := z.Analyze(ring, center, span, window); zr.Valid {
			band.Reeds = FindPeaks(zr, reeds, minSepHz)
			band.Valid = true
		}
		bands = append(bands, band)
	}
	return bands
}

// How many partials of a solved rank to predict into the bands above. 2..8 reaches three octaves up,
// which covers any register a card has columns for; odd partials fall between the octave bands.
const harmonicMax = 8

// Extra peaks to look for per band beyond its genuine ranks, so a leaked partial standing beside a
// fundamental is actually found before it can be subtracted.
const harmonicSlack = 2

// AnalyzeCompound reads a compound register and hands back each octave's own reeds, the ranks below
// subtracted out. Each band is over-resolved (its ranks plus slack) so a lower rank's leaked partial
// is seen, then SubtractHarmonics removes those partials bottom-up. tolFor is the drift prior.
func (z *Zoom) AnalyzeCompound(ring *Ring, baseHz float64, reqs []OctaveRequest, minSepHz float64, window time.Duration, tolFor func(freq float64) float64) []OctaveBand {
	raw := make([]OctaveRequest, len(reqs))
	ranks := make([]int, len(reqs))
	for i, r := range reqs {
		ranks[i] = r.Reeds
		raw[i] = OctaveRequest{Offset: r.Offset, Reeds: r.Reeds + harmonicSlack}
	}
	bands := AnalyzeOctaves(z, ring, baseHz, raw, minSepHz, window)
	return SubtractHarmonics(bands, ranks, tolFor)
}

// SubtractHarmonics removes, bottom-up, the partials a lower rank leaks into the bands above. The
// lowest rank is spectrally clean, so it is solved first and its partials predicted upward; each band
// is then cleaned of what the ranks below it predict before adding its own.
//
// ranks[i] is how many genuine ranks band i has. A peak is dropped as a ghost only while doing so
// leaves the band at least that many: a partial falling exactly on a genuine reed cannot be told from
// it in one spectrum, so the coincident case keeps the reed. tolFor is the half-window, in Hz, around
// a predicted partial - wide enough to catch an inharmonic partial, tight enough not to swallow a
// detuned reed - and the caller sets it because that width is the bellows-drift prior.
func SubtractHarmonics(bands []OctaveBand, ranks []int, tolFor func(freq float64) float64) []OctaveBand {
	// Traverse lowest band first, whatever order they arrived in; return them in that same order.
	order := make([]int, len(bands))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return bands[order[a]].Center < bands[order[b]].Center })

	out := make([]OctaveBand, len(bands))
	copy(out, bands)

	var predicted []float64 // partial frequencies leaked from the ranks already solved
	for _, idx := range order {
		want := 1
		if idx < len(ranks) {
			want = ranks[idx]
		}
		if want < 0 {
			want = 0
		}

		var genuine, ghosts []Peak
		for _, p := range out[idx].Reeds {
			if isHarmonic(p.Freq, predicted, tolFor) {
				ghosts = append(ghosts, p)
			} else {
				genuine = append(genuine, p)
			}
		}

		// A ghost sitting on a real reed must not starve the band below its rank count: give the
		// strongest ghosts back until it has what the register says it has.
		sort.SliceStable(ghosts, func(a, b int) bool { return ghosts[a].Amp > ghosts[b].Amp })
		for len(genuine) < want && len(ghosts) > 0 {
			genuine = append(genuine, ghosts[0])
			ghosts = ghosts[1:]
		}

		// Keep the `want` strongest, returned ascending in frequency.
		sort.SliceStable(genuine, func(a, b int) bool { return genuine[a].Amp > genuine[b].Amp })
		if len(genuine) > want {
			genuine = genuine[:want]
		}
		sort.SliceStable(genuine, func(a, b int) bool { return genuine[a].Freq < genuine[b].Freq })
		out[idx].Reeds = genuine

		// This band's own reeds leak upward in turn.
		for _, r := range genuine {
			for m := 2; m <= harmonicMax; m++ {
				predicted = append(predicted, r.Freq*float64(m))
			}
		}
	}
	return out
}

func isHarmonic(f float64, predicted []float64, tolFor func(freq float64) float64) bool {
	for _, h := range predicted {
		if math.Abs(f-h) <= tolFor(h) {
			return true
		}
	}
	return false
}

// CentsWindow is a convenient tolFor for SubtractHarmonics: a half-window that is the same number of
// cents at every frequency, which is how a reed's drift scales.
func CentsWindow(cents float64) func(freq float64) float64 {
	return func(freq float64) float64 {
		return freq * (math.Exp2(cents/1200) - 1)
	}
}
