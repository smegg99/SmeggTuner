package meter

import "math"

// Acceptance constants of the reference estimator, in seconds and percent.
const (
	clockHistoryFrom  = 0.75
	clockShortWindow  = 5.0
	clockLongWindow   = 20.0
	clockMinTau       = 20.75
	clockMinCount     = 6
	clockMinSpan      = 19.0
	clockMaxRange     = 0.001  // 10 ppm
	clockMaxDrift     = 0.0005 // 5 ppm
	clockTimeout      = 500.0
	clockTimeoutRange = 0.01 // a timed-out session wider than this over 5 s is unreliable
	clockHistory      = 2048
)

// clockEntry is one history point, kept in binary32 as the reference keeps it.
type clockEntry struct{ e, t float32 }

// clockRing holds the newest clockHistory entries.
type clockRing struct {
	buf  [clockHistory]clockEntry
	next int
	n    int
}

func (r *clockRing) reset() { r.next, r.n = 0, 0 }

func (r *clockRing) push(e, t float64) {
	r.buf[r.next] = clockEntry{float32(e), float32(t)}
	r.next = (r.next + 1) % clockHistory
	r.n = min(r.n+1, clockHistory)
}

// since visits the entries with time >= from, oldest first.
func (r *clockRing) since(from float64, visit func(e, t float64)) {
	for i := range r.n {
		x := r.buf[(r.next-r.n+i+clockHistory)%clockHistory]
		if float64(x.t) >= from {
			visit(float64(x.e), float64(x.t))
		}
	}
}

// clockWindow is the count, mean, value range and time span of the entries since a time.
type clockWindow struct {
	n, mean, spread, span, slope float64
}

func (r *clockRing) window(from float64) clockWindow {
	var n, se, st, stt, ste float64
	elo, ehi, tlo, thi := math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)
	r.since(from, func(e, t float64) {
		n++
		se, st, stt, ste = se+e, st+t, stt+t*t, ste+t*e
		elo, ehi, tlo, thi = min(elo, e), max(ehi, e), min(tlo, t), max(thi, t)
	})
	if n == 0 {
		return clockWindow{}
	}
	w := clockWindow{n: n, mean: se / n, spread: ehi - elo, span: thi - tlo}
	if den := n*stt - st*st; n >= 2 && math.Abs(den) > 1e-12 {
		w.slope = (n*ste - st*se) / den
	}
	return w
}

// clockAcceptance is the reference's half-second acceptance estimator over the rate since the reference point.
type clockAcceptance struct{ hist clockRing }

// clockStep is one acceptance decision; value is set once done.
type clockStep struct {
	spread  float64 // range of the estimates over the last 5 s
	done    bool
	timeout bool
	value   float64
}

// step takes the estimate e at tau seconds after the reference and wall seconds into the session.
func (a *clockAcceptance) step(e, tau, wall float64) clockStep {
	if tau >= clockHistoryFrom {
		a.hist.push(e, tau)
	}
	s := clockStep{spread: a.hist.window(tau - clockShortWindow).spread}
	if wall >= clockTimeout {
		s.done, s.timeout, s.value = true, true, round3(e)
		return s
	}
	w := a.hist.window(tau - clockLongWindow)
	if tau >= clockMinTau && w.n >= clockMinCount && w.span >= clockMinSpan &&
		w.spread <= clockMaxRange && math.Abs(w.slope)*w.span <= clockMaxDrift {
		s.done, s.value = true, round3(w.mean)
	}
	return s
}

// round3 rounds to 0.001 with ties away from zero; NaN stays NaN and a tiny negative gives -0.
func round3(x float64) float64 {
	m := math.Floor(math.Abs(x)*1000+0.5) / 1000
	if x < 0 {
		return -m
	}
	return m
}
