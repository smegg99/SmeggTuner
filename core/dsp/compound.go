package dsp

import (
	"math"
	"time"

	"smegg.me/smeggtuner/core/tuning"
)

// The compound fine stage: a register spanning octaves gets one band per octave, each resolved on
// its own, the lower ranks' partials subtracted, and a blocked rank refused rather than read off its
// own harmonic. Single-octave registers never come here; the musette machinery in build.go is theirs.

// compoundBand is one octave of the register as this hop measured it.
type compoundBand struct {
	req     OctaveRequest
	center  float64 // Hz, the transposed base shifted by the octave
	zr      ZoomResult
	genuine []Peak
	ghosts  []Peak
	report  BandReport
	// The band's amplitude envelope beat, when it has one: multiplicity the spectrum may not resolve.
	envHz, envDepth float64
	envOK           bool
	// partialImage says the band's lines were the image of the cluster below, not voices; it blocks
	// the phase-lock backfill, which cannot speak for a multi-line source.
	partialImage bool
	// beatHz is the octave beat read off the residual's rotation when the band merged with the
	// partial below: the beat a technician hears between the rank and the rank an octave under,
	// measurable even below what the envelope resolves. Signed: negative is flat of the partial.
	beatHz float64
	beatOK bool
}

// ghostTol is the half-window around a predicted partial within which a line is that partial.
func ghostTol(f float64) float64 {
	return math.Max(ghostFloorHz, f*ghostRelTol)
}

// compound says whether the engine measures octave bands: only when the layout leaves the note's own
// octave, so an all-8' register keeps the single band and its pair splitter.
func (e *Engine) compound() bool {
	for _, r := range e.cfg.Octaves {
		if r.Offset != 0 {
			return true
		}
	}
	return false
}

// compoundBands analyzes one octave band per layout entry around the base pitch fc, drops lines no
// register would own (the cross-band floor), and subtracts the lower ranks' partials. Bands come
// back ascending in center; ok is false when any band could not be analyzed.
func (e *Engine) compoundBands(zoom *Zoom, ring *Ring, fc float64, window time.Duration) ([]compoundBand, bool) {
	reqs := e.cfg.Octaves // ascending: fill normalized them
	bands := make([]compoundBand, len(reqs))
	raw := make([]OctaveBand, len(reqs))
	wide := make([]int, len(reqs))
	var loudest float64
	for i, req := range reqs {
		center := fc * math.Exp2(float64(req.Offset)/12)
		span := math.Max(16, center*0.035)
		zr := zoom.Analyze(ring, center, span, window)
		if !zr.Valid {
			return nil, false
		}
		peaks := FindPeaks(zr, req.Reeds+harmonicSlack, e.lobeWidth())
		for _, p := range peaks {
			if p.Amp > loudest {
				loudest = p.Amp
			}
		}
		bands[i] = compoundBand{req: req, center: center, zr: zr}
		raw[i] = OctaveBand{Offset: req.Offset, Center: center, Reeds: peaks, Valid: true}
		// Trim nothing here: the base resolution reads overflow, and the rank trim waits for the
		// ghost verdicts.
		wide[i] = req.Reeds + harmonicSlack
	}

	// A rank's line anywhere is far louder than an empty band's noise; FindPeaks floors only within
	// its own band, so an octave whose rank is silent would otherwise report that noise as a reed.
	for i := range raw {
		kept := raw[i].Reeds[:0]
		for _, p := range raw[i].Reeds {
			if p.Amp >= crossBandFloor*loudest {
				kept = append(kept, p)
			}
		}
		raw[i].Reeds = kept
	}

	for i, b := range SubtractHarmonics(raw, wide, ghostTol) {
		bands[i].genuine = b.Reeds
		bands[i].ghosts = b.Ghosts
		if hz, depth, ok := EnvelopeBeat(bands[i].zr, minBeatHz, 25); ok {
			bands[i].envHz, bands[i].envDepth, bands[i].envOK = hz, depth, true
		}
	}
	markPartialImages(bands)
	return bands, true
}

// markPartialImages catches a cluster the lower band could not resolve masquerading as reeds an
// octave up. Reeds too close to split still beat, and their partials carry the beat magnified: the
// k-th partials of reeds beating at env sit k*env apart, and they often outweigh the fundamentals
// (measured 5x on the low-D 16+8+8+8 take, whose partials once read as three phantom voices). Upper
// lines spaced exactly k*env are that image, not voices, and go to the band's ghosts. A genuine
// rank below never trips this: one reed has no deep envelope beat.
func markPartialImages(bands []compoundBand) {
	for i := 1; i < len(bands); i++ {
		j := nearestLined(bands, i)
		if j < 0 {
			continue
		}
		lo := &bands[j]
		if !lo.envOK || lo.envDepth < reedBeatDepth {
			continue
		}
		// The lattice is looked for over every line the band held - the subtraction may already have
		// taken the middle of the image as an ordinary ghost, and a broken adjacency would hide it.
		spacing := float64(int(1)<<((bands[i].req.Offset-lo.req.Offset)/12)) * lo.envHz
		all := append(append([]Peak(nil), bands[i].genuine...), bands[i].ghosts...)
		for a := 1; a < len(all); a++ {
			for b := a; b > 0 && all[b].Freq < all[b-1].Freq; b-- {
				all[b], all[b-1] = all[b-1], all[b]
			}
		}
		matched := map[float64]bool{}
		for a := 0; a+1 < len(all); a++ {
			if math.Abs((all[a+1].Freq-all[a].Freq)-spacing) <= partialImageTol {
				matched[all[a].Freq], matched[all[a+1].Freq] = true, true
			}
		}
		if len(matched) == 0 {
			continue
		}
		bands[i].partialImage = true
		var keep []Peak
		for _, p := range bands[i].genuine {
			if matched[p.Freq] {
				bands[i].ghosts = append(bands[i].ghosts, p)
			} else {
				keep = append(keep, p)
			}
		}
		bands[i].genuine = keep
		bands[i].ghosts = sortByAmp(bands[i].ghosts)
	}
}

// nearestLined is the closest band below i with a line to measure against.
func nearestLined(bands []compoundBand, i int) int {
	for j := i - 1; j >= 0; j-- {
		if len(bands[j].genuine) > 0 {
			return j
		}
	}
	return -1
}

// resolveBase interprets the tracked note as each octave of the layout and keeps the interpretation
// whose bands look most like the declared register: every rank found scores, lines a band cannot own
// count against. The detector tracks the lowest sounding fundamental, so on a tie the tracked note
// is read as the lowest declared rank - with one fundamental and only its harmonics above, nothing
// acoustic can say which rank is sounding, and the register's own lowest is the honest default.
func (e *Engine) resolveBase(zoom *Zoom, ring *Ring, tracked tuning.Note, window time.Duration) (tuning.Note, []compoundBand, bool) {
	var bestNote tuning.Note
	var bestBands []compoundBand
	bestScore := math.Inf(-1)

	tried := map[tuning.Note]bool{}
	for _, req := range e.cfg.Octaves {
		base := tracked - tuning.Note(req.Offset)
		if tried[base] || !base.Valid() {
			continue
		}
		tried[base] = true
		fc := base.Transpose(e.cfg.Transpose).Freq(e.cfg.A4)
		bands, ok := e.compoundBands(zoom, ring, fc, window)
		if !ok {
			continue
		}
		score := 0.0
		for _, b := range bands {
			g := len(b.genuine)
			score += float64(min(g, b.req.Reeds)) - 0.5*float64(max(0, g-b.req.Reeds))
			// The envelope is multiplicity evidence the spectrum may lack: a band declared one rank
			// does not beat itself, and a multi-rank band short of lines is still audibly several.
			if deep := b.envOK && b.envDepth >= reedBeatDepth; deep && g < b.req.Reeds {
				score += 0.5
			} else if deep && b.req.Reeds == 1 && g <= 1 {
				score -= 0.5
			}
		}
		// Strictly-better keeps the first (lowest-rank) interpretation on a tie: offsets ascend, so
		// the lowest offset builds the highest base and is tried first.
		if score > bestScore {
			bestScore, bestNote, bestBands = score, base, bands
		}
	}
	if bestBands == nil {
		return 0, nil, false
	}
	return bestNote, bestBands, true
}

// compoundVerdicts trims each band to its declared ranks and rules on the reed-or-partial question.
//
// A single-rank band whose one line sits within a lobe of the rank below's partial is the hard
// case: reed, partial, or both merged, and one window cannot say which when the beat is slow. The
// residual's angle across hops can (see HarmonicPLV): a partial's angle holds still, a reed's
// rotates at the beat - so the band is ruled independent when its angle walks, and the walk rate IS
// the octave beat, reported even below what the envelope resolves. When the calibration sweep has
// taught the rank below's voice (RankProfile), amplitude judges too, and it works at zero beat: a
// line standing profileExcess over the calibrated partial is more than the partial can be. Until
// any of that has spoken, the phase-locking value alone rules. A band short of several ranks keeps
// the plain value rule: its residual mixes lines and the angle speaks for no one.
func (e *Engine) compoundVerdicts(bands []compoundBand, base tuning.Note, now float64) {
	for i := range bands {
		b := &bands[i]
		want := b.req.Reeds
		if len(b.genuine) > want {
			b.genuine = topPeaks(b.genuine, want)
		}

		src, k := plvSource(bands, i)
		if src == nil {
			b.report = BandReport{Octave: b.req.Offset, Ranks: want, Found: len(b.genuine)}
			continue
		}
		srcFreq := strongestPeak(src.genuine).Freq

		ghostOnly := false
		switch cand, isGhost, solo := soloCandidate(b, srcFreq*float64(k), e.lobeWidth()); {
		case solo:
			plv := HarmonicPLV(src.zr, srcFreq, k, b.zr)
			var rate float64
			var settled bool
			locked := false
			if i < len(e.compTracks) {
				tr := &e.compTracks[i]
				if angle, ok := ResidualAngle(src.zr, srcFreq, k, b.zr, cand.Freq); ok {
					tr.observe(angle, now)
				}
				if plv >= plvLocked {
					tr.locked = true
				}
				rate, settled, locked = tr.rate, tr.settled, tr.locked
			}
			// The calibrated voice, when the sweep taught one, judges the amplitude side and
			// replaces the warm value rule: matched amplitude is the partial whatever the value
			// dips to, excess amplitude is a second voice however slowly it beats.
			byAmp, ampRuled := false, false
			if r, ok := e.profileFor(src.req.Offset, base+tuning.Note(src.req.Offset), k); ok {
				byAmp = cand.Amp >= profileExcess*r*strongestPeak(src.genuine).Amp
				ampRuled = true
			}
			// A band that has ever locked this note is never called independent on a mere dip - a
			// blocked rank's dying bellows collapses the value too. Once locked it takes the
			// sustained rotation; never locked, the dip is a beat too fast to lock at all.
			independent := byAmp ||
				(settled && math.Abs(rate) >= rotIndependent) ||
				(!ampRuled && !settled && !locked && plv < plvWarmIndependent)
			switch {
			case independent:
				if isGhost {
					b.genuine = backfillGhosts(b.genuine, b.ghosts, want)
				}
				if settled {
					b.beatHz, b.beatOK = rate/(2*math.Pi), true
				}
			case isGhost:
				ghostOnly = true
			default:
				b.ghosts = append(b.ghosts, cand)
				b.genuine = nil
				ghostOnly = true
			}

		case b.partialImage:
			// The ghosts are a cluster's image; the phase lock cannot speak for a source of
			// several lines, so nothing comes back.
			ghostOnly = len(b.genuine) == 0

		case len(b.genuine) < want && len(b.ghosts) > 0:
			if HarmonicPLV(src.zr, srcFreq, k, b.zr) < plvLocked {
				b.genuine = backfillGhosts(b.genuine, b.ghosts, want)
			} else {
				ghostOnly = len(b.genuine) == 0
			}
		}

		b.report = BandReport{Octave: b.req.Offset, Ranks: want, Found: len(b.genuine), GhostOnly: ghostOnly}
	}
}

// soloCandidate is the one line of a single-rank band sitting within a lobe of the predicted
// partial - too close for the spectrum to have told the two apart, so the verdict machinery must.
// isGhost says which side of the subtraction it landed on.
func soloCandidate(b *compoundBand, pred, lobe float64) (Peak, bool, bool) {
	if b.req.Reeds != 1 || len(b.genuine)+len(b.ghosts) != 1 {
		return Peak{}, false, false
	}
	if len(b.ghosts) == 1 {
		if math.Abs(b.ghosts[0].Freq-pred) < lobe {
			return b.ghosts[0], true, true
		}
		return Peak{}, false, false
	}
	if math.Abs(b.genuine[0].Freq-pred) < lobe {
		return b.genuine[0], false, true
	}
	return Peak{}, false, false
}

// bandTrack follows one band's residual angle across hops. The rate is the median of rates taken
// over a lag of several hops: a true beat advances every pair alike, while a bellows transient
// steps the angle once and contaminates only the pairs that straddle it - an endpoint fit would
// carry that step for the whole baseline.
type bandTrack struct {
	at  [rotSpan]float64
	ang [rotSpan]float64 // unwrapped
	n   int // total samples ever; ring position is n % rotSpan
	// the last verdict, held between samples so a gated-out transient does not blank it
	rate    float64
	settled bool
	// locked latches once the band's phase-locking value reaches plvLocked this note: from then on
	// only sustained rotation may call the band independent.
	locked bool
}

// rotSpan is how many hops the rotation baseline reaches back (~2.7 s at the fine cadence) and
// rotLag how far apart each rate pair sits.
const (
	rotSpan = 32
	rotLag  = 8
)

// rotSettleSec is how long the angle must have been watched before its rate is believed.
const rotSettleSec = 1.2

func (t *bandTrack) observe(angle, now float64) {
	u := angle
	if t.n > 0 {
		prev := t.ang[(t.n-1)%rotSpan]
		u = prev + math.Remainder(u-prev, 2*math.Pi)
	}
	t.at[t.n%rotSpan], t.ang[t.n%rotSpan] = now, u
	t.n++

	kept := min(t.n, rotSpan)
	if kept <= rotLag {
		return
	}
	rates := make([]float64, 0, kept-rotLag)
	for s := t.n - kept; s+rotLag < t.n; s++ {
		a, b := s%rotSpan, (s+rotLag)%rotSpan
		if dt := t.at[b] - t.at[a]; dt > 0 {
			rates = append(rates, (t.ang[b]-t.ang[a])/dt)
		}
	}
	if len(rates) == 0 {
		return
	}
	oldest := (t.n - kept) % rotSpan
	t.rate = median(rates)
	t.settled = len(rates) >= rotSpan/2 && now-t.at[oldest] >= rotSettleSec
}

// plvSource is the nearest band below i with a line to lock against, and the partial number that
// lands it in band i. Offsets are whole octaves apart, so the ratio is a power of two.
func plvSource(bands []compoundBand, i int) (*compoundBand, int) {
	j := nearestLined(bands, i)
	if j < 0 {
		return nil, 0
	}
	return &bands[j], 1 << ((bands[i].req.Offset - bands[j].req.Offset) / 12)
}

func strongestPeak(ps []Peak) Peak {
	best := ps[0]
	for _, p := range ps[1:] {
		if p.Amp > best.Amp {
			best = p
		}
	}
	return best
}

// topPeaks keeps the n strongest, returned ascending in frequency.
func topPeaks(ps []Peak, n int) []Peak {
	return backfillGhosts(nil, sortByAmp(ps), n)
}

func sortByAmp(ps []Peak) []Peak {
	out := append([]Peak(nil), ps...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].Amp > out[j-1].Amp; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
