package meter

// ScoreTable is the stored coarse state the selector reads; Scores[109] is the upper neighbour of note 108 (fine record 0).
type ScoreTable struct {
	Scores     [110]float32
	Foreground [109]bool
}

// Nomination is the caller state at ordinary dispatch, read after the lock countdown step.
type Nomination struct {
	Octave              int // -1 when unset
	PitchClass          int // -1 when unset; otherwise the manual note wins
	Locked              bool
	AcceptedNote        int
	Recording           bool // the reference record index is set
	MeasureContinuously bool // skips manual validation while not recording
}

// BandPeakScore is the largest local maximum in inclusive [lower, upper], or zero when there is none.
func BandPeakScore(magnitudes []float32, lower, upper int) float32 {
	var score float32
	for k := lower; k <= upper; k++ {
		if v := magnitudes[k]; v > magnitudes[k-1] && v >= magnitudes[k+1] && v > score {
			score = v
		}
	}
	return score
}

// Automatic scans max(4, start)..108 upward, replacing the winner only above 30x its score; start above 108 panics.
func (t *ScoreTable) Automatic(start, octave, pitchClass int) int {
	if start > 108 {
		panic("meter: automatic note scan cannot start above note 108")
	}
	winner, accepted := 0, float32(0)
	for n := max(4, start); n <= 108; n++ {
		if octave >= 0 && n/12 != octave || pitchClass >= 0 && n%12 != pitchClass {
			continue
		}
		if t.Foreground[n] && t.Scores[n] > 30*accepted && t.eligible(n) {
			winner, accepted = n, t.Scores[n]
		}
	}
	return winner
}

// Ordinary picks the manual note, else the locked note, else scans; a failed nomination returns 0 and callers skip it while learning.
func (t *ScoreTable) Ordinary(d Nomination) int {
	switch {
	case d.PitchClass >= 0:
		return t.nominated(12*d.Octave+d.PitchClass, !d.Recording && d.MeasureContinuously)
	case d.Locked:
		return t.nominated(d.AcceptedNote, false)
	}
	return t.Automatic(0, d.Octave, d.PitchClass)
}

// nominated keeps a nonpositive identity unvalidated, as the native caller does.
func (t *ScoreTable) nominated(note int, bypass bool) int {
	if note <= 0 || bypass || t.eligible(note) {
		return note
	}
	return 0
}

func (t *ScoreTable) eligible(n int) bool {
	s := t.Scores[n]
	return t.Foreground[n] && s > t.Scores[n-1]+t.Scores[n-1] && s > t.Scores[n+1]+t.Scores[n+1]
}
