package meter

const (
	TargetUpperLimit = 50.0
	// TargetHertzScale is binary32 0.06 widened, not the binary64 0.06 of the measurement coordinate.
	TargetHertzScale = 0x1.eb851ep-5
)

// NoteCurve holds one value per native note 4..108, in that order.
type NoteCurve [ActiveNotes]float64

// NoteCoverage is the per-note written flag shared by all three slots.
type NoteCoverage [ActiveNotes]bool

// TargetTables is a completed full refresh: three slot curves and their shared coverage.
type TargetTables struct {
	Slots    []NoteCurve
	Coverage NoteCoverage
}

// BuildSlotTargets rebuilds one slot from controls above Missing; nil nominal keeps linear units, else it holds F(n) in Hz.
func BuildSlotTargets(controls NoteCurve, coverage NoteCoverage, nominal *NoteCurve) (NoteCurve, NoteCoverage) {
	var targets NoteCurve
	for i := range targets {
		targets[i] = Missing
	}
	left := -1
	for right, anchor := range controls {
		if anchor > Missing {
			if left < 0 {
				left = right
			}
			targetsSpan(&targets, &coverage, left, right, controls[left], anchor, nominal)
			left = right
		}
	}
	return targets, coverage
}

// targetsSpan accumulates from the raw left anchor; the right endpoint is not pinned to its control.
func targetsSpan(targets *NoteCurve, coverage *NoteCoverage, left, right int, from, to float64, nominal *NoteCurve) {
	step := 0.0
	if right != left {
		step = (to - from) / float64(right-left)
	}
	value := from
	for i := left; i <= right; i++ {
		current := value
		value += step
		coverage[i] = true
		if nominal != nil {
			current = (current * 100) / (nominal[i] * TargetHertzScale)
		}
		if current > TargetUpperLimit {
			current = Missing
		}
		targets[i] = current
	}
}

// BuildTargets is the full refresh: coverage starts clear and gathers every slot's spans.
func BuildTargets(controls []NoteCurve, nominal *NoteCurve) TargetTables {
	t := TargetTables{Slots: make([]NoteCurve, len(controls))}
	for slot := range controls {
		t.Slots[slot], t.Coverage = BuildSlotTargets(controls[slot], t.Coverage, nominal)
	}
	return t
}

// Note returns n slot targets and the coverage of native note 4..108; slots beyond the tables are Missing.
func (t *TargetTables) Note(note, n int) ([]float64, bool) {
	i := note - FirstActiveNote
	out := make([]float64, n)
	for slot := range out {
		out[slot] = Missing
		if slot < len(t.Slots) {
			out[slot] = t.Slots[slot][i]
		}
	}
	return out, t.Coverage[i]
}
