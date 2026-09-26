package meter

// CompoundCoarse is one Chords/Bass coarse stage: raw candidates, the post-retention tuple and whether selection ran.
type CompoundCoarse struct {
	Raw          [3]int
	Working      [3]int
	Label        string
	Published    bool
	LearningStep int
}

// RefreshCompound is learning preparation, selection from the stored table, Chords retention, then a refresh
// that marks exactly the working tuple's notes as selected. While learning, nothing is selected or published.
func (s *CoarseState) RefreshCompound(spectrum []float32, n CompoundNomination, r *ChordRetention,
	label func(ChordIdentity) string, tick func() uint32) CompoundCoarse {
	out := CompoundCoarse{LearningStep: s.Learning.Prepare(&s.Notes, spectrum, s.Background)}
	s.before(spectrum)
	if s.Learning.SelectionAllowed() {
		table := ScoreTable{Scores: s.Scores}
		for i := range s.Notes {
			table.Foreground[i+FirstActiveNote] = s.Notes[i].Foreground
		}
		out.Raw = table.CompoundCandidates(n)
		out.Working = out.Raw
		if n.Chords {
			text := UnknownChord
			if id, ok := RecognizeChord(out.Raw); ok {
				text = label(id)
			}
			kept := r.Filter(out.Raw, text, tick)
			out.Working, out.Label = kept.Notes, kept.Label
		}
		out.Published = true
	}
	s.after(spectrum, out.Working[:]...)
	return out
}
