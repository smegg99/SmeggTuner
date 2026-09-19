package meter

// UnknownChord is the native unknown label; an empty label is a distinct known value.
const UnknownChord = "?"

// ChordRetention is the Chords filter memory; ticks are millisecond reads of one uint32 clock.
type ChordRetention struct {
	Notes      [3]int
	Label      string
	ChangeTick uint32
	LastTick   uint32
}

// ChordRetained is one filter outcome; ExpiredStored says the expiry branch ran, not that memory is empty.
type ChordRetained struct {
	Notes            [3]int
	Label            string
	RetainedPrevious bool
	ExpiredStored    bool
}

// retentionRelation is the positional (candidate, stored) equality set; it is not set overlap.
var retentionRelation = [7][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}, {1, 2}, {2, 1}, {2, 2}}

// Filter runs one Chords filter call; tick is called once per native tick read, and ages wrap as uint32.
func (r *ChordRetention) Filter(notes [3]int, label string, tick func() uint32) ChordRetained {
	out := ChordRetained{Notes: notes, Label: label}
	change := r.ChangeTick
	if notes != r.Notes {
		matches := 0
		for _, pair := range retentionRelation {
			if notes[pair[0]] == r.Notes[pair[1]] {
				matches++
			}
		}
		out.RetainedPrevious = matches >= 2 && tick()-change > 750
		if !out.RetainedPrevious {
			change = tick()
		}
	}
	if out.RetainedPrevious {
		out.Notes, out.Label = r.Notes, r.Label
	}
	stored, storedLabel := r.Notes, r.Label
	if tick()-r.LastTick > 3000 {
		out.ExpiredStored = true
		stored, storedLabel = [3]int{}, ""
	}
	var last uint32
	if out.Label == UnknownChord {
		out.Notes = [3]int{}
		change = 0
	} else {
		if change == 0 {
			change = tick()
		}
		stored, storedLabel = out.Notes, out.Label
		last = tick()
	}
	*r = ChordRetention{Notes: stored, Label: storedLabel, ChangeTick: change, LastTick: last}
	return out
}
