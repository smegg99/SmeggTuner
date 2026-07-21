package dsp

import (
	"math"
	"sort"
)

// Peak is one spectral peak found in a ZoomResult.
type Peak struct {
	Freq float64 // absolute Hz
	Amp  float64 // linear magnitude
}

// A candidate is a line of its own only if the spectrum falls between it and the stronger peak beside
// it, to at least this fraction of the weaker. A single reed's rippled shoulders make local maxima at
// 15-25% of the peak; what they lack is a valley, and two reeds always leave one.
const peakValleyRatio = 0.5

// How loud a line has to be, against the loudest one, to be another reed.
const reedPeakFloor = 0.35

// FindPeaks returns up to maxPeaks spectral lines, strongest first, each at least minSepHz from the
// others and separated by a real valley, refined by parabolic interpolation and returned ascending.
func FindPeaks(zr ZoomResult, maxPeaks int, minSepHz float64) []Peak {
	if !zr.Valid || len(zr.Spec) < 3 || maxPeaks <= 0 {
		return nil
	}
	var maxAmp float64
	for _, v := range zr.Spec {
		if v > maxAmp {
			maxAmp = v
		}
	}
	if maxAmp <= 0 {
		return nil
	}

	type candidate struct {
		Peak
		bin int
	}

	// Reeds in one register sound within a few dB of each other, so a line far quieter than the
	// loudest is the bellows, not a reed. This also clears the Hann sidelobe floor (31 dB down) wide.
	thresh := reedPeakFloor * maxAmp
	var cands []candidate
	for i := 1; i < len(zr.Spec)-1; i++ {
		v := zr.Spec[i]
		// strict on one side so a flat two-bin top yields one candidate
		if v < thresh || v < zr.Spec[i-1] || v <= zr.Spec[i+1] {
			continue
		}
		la := math.Log(zr.Spec[i-1] + 1e-30)
		lb := math.Log(v + 1e-30)
		lc := math.Log(zr.Spec[i+1] + 1e-30)
		den := la - 2*lb + lc
		var d float64
		if den != 0 {
			d = 0.5 * (la - lc) / den
		}
		cands = append(cands, candidate{
			Peak: Peak{Freq: zr.MinHz + (float64(i)+d)*zr.BinHz, Amp: v},
			bin:  i,
		})
	}

	// valleyBetween is the lowest the spectrum gets between two bins.
	valleyBetween := func(a, b int) float64 {
		if a > b {
			a, b = b, a
		}
		low := math.Inf(1)
		for i := a; i <= b; i++ {
			if zr.Spec[i] < low {
				low = zr.Spec[i]
			}
		}
		return low
	}

	sort.Slice(cands, func(a, b int) bool { return cands[a].Amp > cands[b].Amp })

	var out []candidate
	for _, c := range cands {
		ok := true
		for _, o := range out {
			if math.Abs(o.Freq-c.Freq) < minSepHz {
				ok = false
				break
			}
			if valleyBetween(o.bin, c.bin) > peakValleyRatio*c.Amp {
				ok = false // a shoulder of o, not a line of its own
				break
			}
		}
		if ok {
			out = append(out, c)
			if len(out) == maxPeaks {
				break
			}
		}
	}

	sort.Slice(out, func(a, b int) bool { return out[a].Freq < out[b].Freq })
	peaks := make([]Peak, len(out))
	for i, c := range out {
		peaks[i] = c.Peak
	}
	return peaks
}

func RefinePhase(prev, cur ZoomResult, f float64, hopSeconds float64) float64 {
	if !prev.Valid || !cur.Valid || hopSeconds <= 0 ||
		prev.BinHz != cur.BinHz || prev.MinHz != cur.MinHz || prev.Center != cur.Center ||
		prev.Rate != cur.Rate || len(prev.TimeSeries) != len(cur.TimeSeries) {
		return f
	}
	if cur.BinHz <= 0 {
		return f
	}
	bin := int(math.Round((f - cur.MinHz) / cur.BinHz))
	if bin < 0 || bin >= len(cur.Phases) || bin >= len(prev.Phases) {
		return f
	}
	dphi := cur.Phases[bin] - prev.Phases[bin]
	// The heterodyne LO restarts at phase zero every Analyze, so bin phase tracks absolute frequency
	// f, not baseband f-Center; hopSeconds must come from exact sample counts (error is f*dt/hop).
	expected := 2 * math.Pi * f * hopSeconds
	dev := dphi - expected
	// wrap into [-pi, pi]
	dev = math.Mod(dev, 2*math.Pi)
	if dev > math.Pi {
		dev -= 2 * math.Pi
	} else if dev < -math.Pi {
		dev += 2 * math.Pi
	}
	return f + dev/(2*math.Pi*hopSeconds)
}
