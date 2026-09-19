package meter

// History is the eight-frame stability reducer behind the lock latches.
// Locked is the internal lock, Acquired the held-result latch; neither is the visible lock.
type History struct {
	LastNotes     [3]int
	AcceptedNotes [3]int
	Values        [3][8]float64
	WriteIndex    int
	Countdown     int
	Locked        bool
	Acquired      bool
	Inhibited     bool
	Settle        int  // native analyses a tuple must stay selected before it may lock; 0 is the reference
	Release       bool // a new nonempty tuple drops the lock at once; the reference holds it for the countdown
	Picks         int  // native analyses the current tuple has stayed selected
}

// BeginAnalysis runs once per handler event before any measurement.
func (h *History) BeginAnalysis() {
	h.Picks++
	if h.Countdown > 0 {
		h.Countdown--
		if h.Countdown == 0 {
			h.Locked = false
		}
	}
}

// Evaluate inserts one frame. A changed tuple clears history and re-arms the countdown
// before the inhibit and all-zero early returns, which leave LastNotes unchanged.
func (h *History) Evaluate(notes [3]int, deviations [3]float64, binCounts [3]int) bool {
	if notes != h.LastNotes {
		h.Values = [3][8]float64{}
		h.WriteIndex = 0
		h.Countdown = 4
		h.Picks = 1
		h.Locked = h.Locked && (!h.Release || notes == [3]int{})
	}
	if h.Inhibited || notes == [3]int{} {
		return false
	}
	for slot := range h.Values {
		value := 0.0
		if notes[slot] != 0 {
			value = deviations[slot]
		}
		h.Values[slot][h.WriteIndex] = value
	}
	h.WriteIndex = (h.WriteIndex + 1) % 8
	accepted := true
	for slot, count := range binCounts {
		if count > 0 && !(historySpan(h.Values[slot]) <= 400/float64(count)) {
			accepted = false
		}
	}
	h.LastNotes = notes
	if accepted && h.Picks < h.Settle {
		accepted = false
	}
	if accepted {
		h.Locked = true
		h.Acquired = true
		h.AcceptedNotes = notes
	}
	return accepted
}

// historySpan reads the physical array, not chronological order: after a zero,
// the upper bound is max(100, values after the last zero).
func historySpan(values [8]float64) float64 {
	low, high := values[0], values[0]
	lastZero := -1
	for i, v := range values {
		low = min(low, v)
		high = max(high, v)
		if v == 0 {
			lastZero = i
		}
	}
	if lastZero >= 0 {
		high = 100
		for _, v := range values[lastZero+1:] {
			high = max(high, v)
		}
	}
	return high - low
}
