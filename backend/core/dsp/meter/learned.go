// backend/core/dsp/meter/learned.go
package meter

// Learned is what a finished learning pass leaves: each note's reference band and seeded threshold, and the
// background aggregate, for the coarse spectrum it was learned on.
type Learned struct {
	plan       learnedPlan
	references [ActiveNotes][]float32
	seeds      [ActiveNotes]float32
	background float32
}

// learnedPlan is what Carry compares: a background learned under another coarse spectrum does not apply.
type learnedPlan struct {
	rate         int
	factor       float64
	table        NoteTable
	hum50, hum60 bool
}

func (a *Analyzer) learnedPlan() learnedPlan {
	return learnedPlan{a.cfg.SampleRate, a.cfg.RateFactor, a.cfg.Table, a.cfg.Hum50, a.cfg.Hum60}
}

// Learned copies the last finished learning pass, or nil while one runs or before any has.
func (a *Analyzer) Learned() *Learned {
	if a.Learning() || !a.coarse.Learning.Done {
		return nil
	}
	l := &Learned{plan: a.learnedPlan(), background: a.coarse.Background}
	for i, n := range a.coarse.Notes {
		l.references[i] = append([]float32(nil), n.Reference...)
		l.seeds[i] = n.Seed
	}
	return l
}

// Adopt seeds a fresh analyzer as a learning pass on the same input would; false when l was learned under another plan.
func (a *Analyzer) Adopt(l *Learned) bool {
	if l == nil || l.plan != a.learnedPlan() {
		return false
	}
	for i := range a.coarse.Notes {
		n := &a.coarse.Notes[i]
		if len(n.Reference) != len(l.references[i]) {
			return false
		}
	}
	for i := range a.coarse.Notes {
		n := &a.coarse.Notes[i]
		copy(n.Reference, l.references[i])
		n.Threshold, n.Seed = l.seeds[i], l.seeds[i]
	}
	a.coarse.Background, a.coarse.Learning.Done = l.background, true
	return true
}
