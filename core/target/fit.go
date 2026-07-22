package target

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"slices"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

// DefaultSpacing is how far apart fitted anchors sit, in semitones: one per
// octave, few enough to stay draggable by hand.
const DefaultSpacing = 12

// DefaultOutlierCents is how far a reed may sit from the trend before the fit drops
// it: five times the reed tolerance, and dropped so it cannot bend the curve
// towards itself.
const DefaultOutlierCents = 5.0

var ErrNoReadings = errors.New("target: no readings to fit")

// Reading is one note of a recorded pass as the fitter needs it: a Take with the
// recording stripped off. session.Pass.Readings builds these.
type Reading struct {
	Note  tuning.Note
	Reeds []dsp.ReedMeasure
	// ReedsMerged says the spectrum did not tell the reeds apart; ReedsFromBeat
	// says they were recovered from the beat anyway (a measurement, not a guess).
	ReedsMerged   bool
	ReedsFromBeat bool
}

type FitOptions struct {
	Name string
	// Spacing between anchors in semitones. 0 means DefaultSpacing.
	Spacing int
	// OutlierCents is the window on DefaultOutlierCents, which 0 means.
	OutlierCents float64
}

// Outlier is a reed left out of the curve; the finished curve is what measures it.
type Outlier struct {
	Note  tuning.Note `json:"note"`
	Reed  int         `json:"reed"`
	Curr  float64     `json:"curr"`  // cents from scale pitch, as recorded
	Goal  float64     `json:"goal"`  // what the fitted curve says at that note
	Error float64     `json:"error"` // Curr - Goal
}

// FitResult is a curve and the diagnosis that came with it.
type FitResult struct {
	Curve    *Curve    `json:"curve"`
	Outliers []Outlier `json:"outliers,omitempty"`
	Used     int       `json:"used"` // readings the curve was fitted to
	// Merged is readings dropped because nothing in them was a reed pitch.
	Merged int `json:"merged"`
}

// Fit recovers the curve an instrument is already holding: per reed, over note
// number, in cents. Each octave's anchor is a median, so one wild reed cannot move
// it; anchors are smoothed against each other, and readings the finished curve
// cannot account for come back in Outliers, not the curve.
//
// A reading whose reeds the spectrum did not separate is dropped and counted in
// Merged, unless ReedsFromBeat, whose frequencies are reed pitches measured off the
// amplitude and fitted like any other. A reed no reading carries comes out flat
// zero. a4 is the reference the pass was recorded against, not the current one.
func Fit(rs []Reading, reedCount int, a4 float64, opts FitOptions) (*FitResult, error) {
	if reedCount < MinReeds || reedCount > MaxReeds {
		return nil, fmt.Errorf("fit: reed count %d out of range %d..%d", reedCount, MinReeds, MaxReeds)
	}
	if len(rs) == 0 {
		return nil, ErrNoReadings
	}
	spacing := opts.Spacing
	if spacing <= 0 {
		spacing = DefaultSpacing
	}
	limit := opts.OutlierCents
	if limit <= 0 {
		limit = DefaultOutlierCents
	}

	res := &FitResult{Curve: NewCurve(opts.Name, reedCount)}
	pts := make([][]point, reedCount)
	var lo, hi tuning.Note

	for _, r := range rs {
		if !r.Note.Valid() {
			continue
		}
		if r.ReedsMerged && !r.ReedsFromBeat {
			res.Merged++
			continue
		}
		res.Used++
		if lo == 0 || r.Note < lo {
			lo = r.Note
		}
		if r.Note > hi {
			hi = r.Note
		}

		ref := r.Note.Freq(a4)
		for i := 0; i < len(r.Reeds) && i < reedCount; i++ {
			c := CurrCents(r.Reeds[i], BandRef(ref, r.Reeds[i].Octave))
			if math.IsNaN(c) || math.IsInf(c, 0) {
				continue
			}
			pts[i] = append(pts[i], point{note: r.Note, bin: int(r.Note) / spacing, curr: c})
		}
	}
	// Every reading was merged: no reed pitch at all. Zero anchors is a legal
	// "no goal" curve; Merged says why.
	if res.Used == 0 {
		return res, nil
	}

	fits := make([][]knot, reedCount)
	for reed := range pts {
		if len(pts[reed]) == 0 {
			continue // no reading carried this reed: it stays flat zero
		}
		ks, bad := fitReed(pts[reed], limit)
		fits[reed] = ks
		for i, out := range bad {
			if out {
				res.Outliers = append(res.Outliers, Outlier{
					Note: pts[reed][i].note,
					Reed: reed,
					Curr: pts[reed][i].curr,
				})
			}
		}
	}

	for _, note := range anchorNotes(fits, lo, hi) {
		a := Anchor{Note: note, Reeds: make([]float64, reedCount)}
		for reed, ks := range fits {
			a.Reeds[reed] = evalKnots(ks, note)
		}
		res.Curve.Anchors = append(res.Curve.Anchors, a)
	}

	// Measure the outliers against the finished curve.
	for i := range res.Outliers {
		o := &res.Outliers[i]
		o.Goal = res.Curve.At(o.Note)[o.Reed]
		o.Error = o.Curr - o.Goal
	}
	slices.SortFunc(res.Outliers, func(a, b Outlier) int {
		return cmp.Or(cmp.Compare(a.Note, b.Note), cmp.Compare(a.Reed, b.Reed))
	})

	if err := res.Curve.Validate(); err != nil {
		return nil, fmt.Errorf("fit: %w", err)
	}
	return res, nil
}
