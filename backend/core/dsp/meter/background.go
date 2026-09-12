package meter

// FirstActiveNote and ActiveNotes span native notes 4..108, the population that learning, targets and the background aggregate cover.
const (
	FirstActiveNote = 4
	ActiveNotes     = 105
)

const (
	bgLearningSteps = 8
	bgQuietNotes    = 27
)

// NoteState is one coarse note record over bins Lower..Lower+len(Reference)-1; NaN, infinity and overflow are unsupported.
type NoteState struct {
	Lower           int
	Reference       []float32
	Threshold       float32
	Foreground      bool
	EntrySnapshot   float32
	EntryKnown      bool
	PreviousMaximum float32
	Q               int32
	Quietest        float32 // the lowest PreviousMaximum of a learning pass, for Learning.QuietSeed
	Seed            float32 // the threshold the last learning pass left; 0 before one
}

// NewNoteState is construction plus the first band allocation: zero reference, Q one, no entry snapshot.
func NewNoteState(lower, upper int) NoteState {
	return NoteState{Lower: lower, Reference: make([]float32, upper-lower+1), Q: 1}
}

// Reconfigure reallocates a zero reference for a new band; threshold, flags, snapshot, maximum and Q survive.
func (n *NoteState) Reconfigure(lower, upper int) {
	n.Lower, n.Reference = lower, make([]float32, upper-lower+1)
}

// NoteStatistics is one update's scan; bins are absolute spectrum indices.
type NoteStatistics struct {
	Maximum     float32
	MaximumBin  int
	Minimum     float32
	MinimumBin  int
	Mean        float32
	PeakCount   int
	PeakMaximum float32
	PeakMean    float32
	PeakRatio   float32
}

// AdaptReference steps each bin toward window by 1/150 when rising, else 1/200, in place.
func AdaptReference(reference, window []float32) {
	if len(reference) != len(window) {
		panic("meter: reference and window lengths differ")
	}
	for j, x := range window {
		b := reference[j]
		divisor := float32(200)
		if x > b {
			divisor = 150
		}
		reference[j] = b + (x-b)/divisor
	}
}

// Update adapts the reference, scans the band and its two neighbours, and decides foreground; raw precedes the ratio veto.
func (n *NoteState) Update(spectrum []float32, selected bool) (stats NoteStatistics, raw bool) {
	upper := n.Lower + len(n.Reference) - 1
	AdaptReference(n.Reference, spectrum[n.Lower:upper+1])
	stats.Maximum, stats.MaximumBin = spectrum[n.Lower], n.Lower
	stats.Minimum, stats.MinimumBin = spectrum[n.Lower], n.Lower
	var total, peakTotal float32
	for k := n.Lower; k <= upper; k++ {
		v := spectrum[k]
		if v > stats.Maximum {
			stats.Maximum, stats.MaximumBin = v, k
		}
		if v < stats.Minimum {
			stats.Minimum, stats.MinimumBin = v, k
		}
		total += v
		if v > spectrum[k-1] && v >= spectrum[k+1] {
			peakTotal += v
			stats.PeakCount++
			if v > stats.PeakMaximum {
				stats.PeakMaximum = v
			}
		}
	}
	// The branch reads the old threshold; decay keeps S - S/3, never 2S/3.
	if float32(10*stats.Maximum) <= n.Threshold {
		n.Threshold -= n.Threshold / 3
	} else {
		divisor := float32(50)
		if selected {
			divisor = 100
		}
		n.Threshold = stats.Maximum/divisor + n.Threshold
	}
	stats.Mean = total / float32(len(n.Reference))
	if stats.PeakCount > 0 {
		stats.PeakMean = peakTotal / float32(stats.PeakCount)
	}
	if stats.PeakMean != 0 {
		stats.PeakRatio = stats.PeakMaximum / stats.PeakMean
	}
	raw = float32(float32(n.Q)*stats.PeakMaximum) > n.Threshold
	// The entry snapshot is written before the ratio veto.
	if raw && !n.Foreground {
		n.EntrySnapshot, n.EntryKnown = n.Threshold, true
	}
	n.Foreground = raw && stats.PeakRatio >= 1.5
	n.PreviousMaximum = stats.Maximum
	return stats, raw
}

// Learning is the eight-update learning transaction; Counter is -1 when idle, else the last step 0..7.
type Learning struct {
	Request bool
	Counter int
	// QuietSeed seeds thresholds from each note's quietest learning step, not the last: a note played
	// while learning would otherwise sit under four times its own level for as long as it sounds.
	QuietSeed bool
	// Done is set by the first pass to finish; what it seeded can start a later run (Analyzer.Learned).
	Done bool
}

// NewLearning is the idle transaction.
func NewLearning() Learning { return Learning{Counter: -1} }

// SelectionAllowed is false while a request is pending, even when the counter is active.
func (l Learning) SelectionAllowed() bool { return !l.Request }

// Prepare runs before selection and Update, from previous maxima; it returns the step, 0 when idle, 1 shows and 8 hides the indicator.
func (l *Learning) Prepare(notes *[ActiveNotes]NoteState, spectrum []float32, oldBackground float32) int {
	if l.Counter == -1 {
		if !l.Request {
			return 0
		}
		l.Counter = 0
	}
	l.Counter++
	step := l.Counter
	if step >= 3 {
		for i := range notes {
			bgSeed(&notes[i], spectrum, step == 3, oldBackground, l.QuietSeed)
		}
	}
	if step == bgLearningSteps {
		l.Request, l.Counter, l.Done = false, -1, true
	}
	return step
}

// bgSeed copies the band on step 3; later steps keep the per-bin minimum and reseed the threshold.
func bgSeed(n *NoteState, spectrum []float32, copyBand bool, oldBackground float32, quiet bool) {
	window := spectrum[n.Lower : n.Lower+len(n.Reference)]
	if copyBand {
		copy(n.Reference, window)
		n.Quietest = n.PreviousMaximum
		return
	}
	n.Quietest = min(n.Quietest, n.PreviousMaximum)
	for j, x := range window {
		if x < n.Reference[j] { // not builtin min: a tie keeps the old signed zero
			n.Reference[j] = x
		}
	}
	if len(n.Reference) == 0 {
		return
	}
	seed := n.PreviousMaximum
	if quiet {
		seed = n.Quietest
	}
	n.Threshold = float32(seed*4) + float32(oldBackground*15)
	n.Seed = n.Threshold
}

// AggregateBackground takes maxima of notes 4..108 in order: the mean of those under a quarter of their mean if 27 qualify, else the mean.
func AggregateBackground(maxima [ActiveNotes]float32) float32 {
	var total float32
	for _, v := range maxima {
		total += v
	}
	mean := total / ActiveNotes
	cutoff := mean * 0.25
	var quiet float32
	count := 0
	for _, v := range maxima {
		if v < cutoff {
			quiet += v
			count++
		}
	}
	if count >= bgQuietNotes {
		return quiet / float32(count)
	}
	return mean
}
