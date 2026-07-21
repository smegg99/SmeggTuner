package report

// Graph geometry is computed here so the template only prints coordinates and the SVG stays self-contained.

import (
	"fmt"
	"math"
	"strings"

	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
)

// Plot dimensions in SVG user units (aspect ratio only, scaled by the stylesheet).
const (
	graphWidth  = 800.0
	graphHeight = 250.0
	padLeft     = 44.0
	padRight    = 10.0
	padTop      = 12.0
	padBottom   = 26.0
	// minSpan stops a well-tuned instrument's cent of noise being magnified to full height.
	minSpan = 20.0
)

// reedColors stay distinct in greyscale; eight because a curve may be eight reeds wide.
var reedColors = []string{
	"#1f4e9c", "#b3401a", "#1e7a45", "#7a2d8f",
	"#8a6a00", "#0f6d78", "#a01c46", "#3d4a56",
}

// Dot is one measured reed on the plot.
type Dot struct {
	X, Y  float64
	InTol bool
	Note  string
	Cents float64
}

// Series is one reed: the goal across the range and what was measured.
type Series struct {
	Reed int
	// Label is this reed's legend name (Report.Head), so legend and column heads agree.
	Label string
	Color string
	// Line is the goal as SVG polyline points; empty when the session has no curve.
	Line string
	Dots []Dot
}

// Tick is one axis label.
type Tick struct {
	Pos   float64
	Label string
}

// Graph is the plot: notes across, cents up.
type Graph struct {
	Width, Height            float64
	Left, Right, Top, Bottom float64
	ZeroY                    float64
	Series                   []Series
	XTicks                   []Tick
	YTicks                   []Tick
	HasCurve                 bool
	// MergedNotes counts notes with no dots (reeds never separated) so the caption can explain the gap.
	MergedNotes int
}

// graph plots the pass against the curve, or nil when there is nothing to draw; merged notes contribute no dots but are counted.
func graph(r *Report, c *target.Curve) *Graph {
	if len(r.Rows) == 0 || len(r.Reeds) == 0 {
		return nil
	}

	lo, hi := r.Summary.MinNote, r.Summary.MaxNote
	if hi < lo {
		return nil
	}

	minC, maxC := 0.0, 0.0
	for _, row := range r.Rows {
		for _, cell := range row.Reeds {
			if !cell.Present {
				continue
			}
			minC = math.Min(minC, math.Min(cell.Curr, cell.Goal))
			maxC = math.Max(maxC, math.Max(cell.Curr, cell.Goal))
		}
	}
	if c != nil {
		for n := lo; n <= hi; n++ {
			for _, v := range c.At(n) {
				minC, maxC = math.Min(minC, v), math.Max(maxC, v)
			}
		}
	}
	minC, maxC = padRange(minC, maxC)

	g := &Graph{
		Width:    graphWidth,
		Height:   graphHeight,
		Left:     padLeft,
		Right:    graphWidth - padRight,
		Top:      padTop,
		Bottom:   graphHeight - padBottom,
		HasCurve: c != nil,
	}

	// A single-note pass is drawn down the middle rather than divided by zero.
	x := func(n float64) float64 {
		if hi == lo {
			return (g.Left + g.Right) / 2
		}
		f := (n - float64(lo)) / float64(hi-lo)
		return g.Left + f*(g.Right-g.Left)
	}
	y := func(cents float64) float64 {
		f := (cents - minC) / (maxC - minC)
		return g.Bottom - f*(g.Bottom-g.Top)
	}
	g.ZeroY = y(0)

	for _, reed := range r.Reeds {
		s := Series{
			Reed:  reed,
			Label: r.Head(reed - 1),
			Color: reedColors[(reed-1)%len(reedColors)],
		}
		if c != nil {
			var pts []string
			for n := lo; n <= hi; n++ {
				goal := c.At(n)
				if reed-1 >= len(goal) {
					continue
				}
				pts = append(pts, fmt.Sprintf("%.1f,%.1f", x(float64(n)), y(goal[reed-1])))
			}
			s.Line = strings.Join(pts, " ")
		}
		for _, row := range r.Rows {
			if reed-1 >= len(row.Reeds) {
				continue
			}
			cell := row.Reeds[reed-1]
			if !cell.Present {
				continue
			}
			s.Dots = append(s.Dots, Dot{
				X:     x(float64(row.Note)),
				Y:     y(cell.Curr),
				InTol: cell.InTol,
				Note:  row.Name,
				Cents: cell.Curr,
			})
		}
		g.Series = append(g.Series, s)
	}

	for _, row := range r.Rows {
		if row.Merged {
			g.MergedNotes++
		}
	}

	// One X label per octave (each C), plus the range ends when too short to hold one.
	for n := lo; n <= hi; n++ {
		if int(n)%12 == 0 {
			g.XTicks = append(g.XTicks, Tick{Pos: x(float64(n)), Label: n.Name(tuning.NamingCDEFGAB)})
		}
	}
	if len(g.XTicks) == 0 {
		g.XTicks = append(g.XTicks,
			Tick{Pos: x(float64(lo)), Label: r.Summary.MinName},
			Tick{Pos: x(float64(hi)), Label: r.Summary.MaxName})
	}
	for _, v := range yTicks(minC, maxC) {
		// Nudge a floating-point near-zero to 0 so the axis never prints "-0".
		if math.Abs(v) < 1e-9 {
			v = 0
		}
		g.YTicks = append(g.YTicks, Tick{Pos: y(v), Label: fmt.Sprintf("%+.0f", v)})
	}
	return g
}
