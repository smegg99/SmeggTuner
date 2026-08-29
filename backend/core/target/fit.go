package target

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"slices"

	"smegg.me/smeggtuner/backend/core/tuning"
)

// DefaultSpacing is how far apart fitted anchors sit, in semitones: one per
// octave, few enough to stay draggable by hand.
const DefaultSpacing = 12

// DefaultOutlier is how far a slot may sit from the trend before the fit drops
// it, in reference units: five times the reed tolerance, and dropped so it cannot
// bend the curve towards itself.
const DefaultOutlier = 5.0

var ErrNoReadings = errors.New("target: no readings to fit")

type FitOptions struct {
	Name string
	// Spacing between anchors in semitones. 0 means DefaultSpacing.
	Spacing int
	// Outlier is the rejection window in reference units; 0 means DefaultOutlier.
	Outlier float64
}

// Outlier is a slot left out of the curve; the finished curve is what measures it.
type Outlier struct {
	Note  tuning.Note `json:"note"`
	Reed  int         `json:"reed"`
	Curr  float64     `json:"curr"`  // reference units from scale pitch, as recorded
	Goal  float64     `json:"goal"`  // what the fitted curve says at that note
	Error float64     `json:"error"` // Curr - Goal
}

// FitResult is a curve and the diagnosis that came with it.
type FitResult struct {
	Curve    *Curve    `json:"curve"`
	Outliers []Outlier `json:"outliers,omitempty"`
	Used     int       `json:"used"` // readings that carried at least one present slot
}

// Fit recovers the curve an instrument is already holding: per slot position, over note
// number, in reference units. Each octave's anchor is a median, so one wild reed cannot move
// it; anchors are smoothed against each other, and readings the finished curve cannot
// account for come back in Outliers, not the curve. A slot no reading carries comes out
// flat zero.
func Fit(rs []Reading, reedCount int, opts FitOptions) (*FitResult, error) {
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
	limit := opts.Outlier
	if limit <= 0 {
		limit = DefaultOutlier
	}

	res := &FitResult{Curve: NewCurve(opts.Name, reedCount)}
	pts := make([][]point, reedCount)
	var lo, hi tuning.Note

	for _, r := range rs {
		if !r.Note.Valid() {
			continue
		}
		heard := false
		for i := 0; i < len(r.Reeds) && i < reedCount; i++ {
			d := r.Reeds[i].Dev
			if !r.Reeds[i].Present || math.IsNaN(d) || math.IsInf(d, 0) {
				continue
			}
			pts[i] = append(pts[i], point{note: r.Note, bin: int(r.Note) / spacing, curr: d})
			heard = true
		}
		if !heard {
			continue
		}
		res.Used++
		if lo == 0 || r.Note < lo {
			lo = r.Note
		}
		if r.Note > hi {
			hi = r.Note
		}
	}
	// Nothing sounded: zero anchors is the legal "no goal" curve.
	if res.Used == 0 {
		return res, nil
	}

	fits := make([][]knot, reedCount)
	for reed := range pts {
		if len(pts[reed]) == 0 {
			continue // no reading carried this slot: it stays flat zero
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
		o.Goal, _ = goalAt(res.Curve.At(o.Note), o.Reed)
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
