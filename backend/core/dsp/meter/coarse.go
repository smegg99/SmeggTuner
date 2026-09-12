package meter

import "slices"

// CoarseState is the coarse table between refreshes. Notes[i] is note i+4; Scores keeps both boundary neighbours.
type CoarseState struct {
	Notes      [ActiveNotes]NoteState
	Scores     [110]float32
	Background float32
	Learning   Learning
	// Follow lets a selected note's threshold fall with its level while it stays well above the background,
	// so a note that goes on more quietly is still heard; the reference keeps the threshold it reached.
	Follow bool
	// Fresh refreshes the table before selecting from it, the last selection counting as selected; the
	// reference selects from the table the previous refresh left, half a second behind the audio.
	Fresh  bool
	chosen []int // the last selection, which Fresh refreshes as selected
}

// NewCoarseState is construction: idle learning and NewNoteState for each band of notes 4..108.
func NewCoarseState(bands *[ActiveNotes]NoteBand) CoarseState {
	s := CoarseState{Learning: NewLearning()}
	for i, b := range bands {
		s.Notes[i] = NewNoteState(b.LowerBin, b.UpperBin)
	}
	return s
}

// CoarseFrame is one refresh: the analysis-local note and the learning step, 0 when idle.
type CoarseFrame struct {
	Selected     int
	LearningStep int
}

// Refresh prepares learning, selects from the stored table, then refreshes notes 4..108 and the background from spectrum.
func (s *CoarseState) Refresh(spectrum []float32, d Nomination) CoarseFrame {
	frame := CoarseFrame{LearningStep: s.Learning.Prepare(&s.Notes, spectrum, s.Background)}
	s.before(spectrum)
	if s.Learning.SelectionAllowed() {
		table := ScoreTable{Scores: s.Scores}
		for i := range s.Notes {
			table.Foreground[i+FirstActiveNote] = s.Notes[i].Foreground
		}
		frame.Selected = table.Ordinary(d)
	}
	s.after(spectrum, frame.Selected)
	return frame
}

// followFloor is the band maximum, in background aggregates, under which Follow leaves a threshold alone
// when no learning pass has seeded the band; a seeded band's floor is its seed.
const followFloor = 15

// update is NoteState.Update for note i+4, with Follow applied first.
func (s *CoarseState) update(i int, spectrum []float32, selected bool) NoteStatistics {
	n := &s.Notes[i]
	if s.Follow && selected && n.PreviousMaximum > 0 {
		peak := spectrum[n.Lower]
		for _, v := range spectrum[n.Lower : n.Lower+len(n.Reference)] {
			peak = max(peak, v)
		}
		floor := n.Seed
		if floor == 0 {
			floor = followFloor * s.Background
		}
		if peak < n.PreviousMaximum && peak > floor {
			n.Threshold *= peak / n.PreviousMaximum
		}
	}
	stats, _ := n.Update(spectrum, selected)
	return stats
}

// refresh updates notes 4..108 and the background; a note in chosen is the selected one.
func (s *CoarseState) refresh(spectrum []float32, chosen []int) {
	var maxima [ActiveNotes]float32
	for i := range s.Notes {
		note := i + FirstActiveNote
		stats := s.update(i, spectrum, slices.Contains(chosen, note))
		s.Scores[note], maxima[i] = stats.PeakMaximum, stats.Maximum
	}
	s.Background = AggregateBackground(maxima)
}

// before is the Fresh refresh ahead of selection; after records the selection and, in the reference
// order, refreshes with it.
func (s *CoarseState) before(spectrum []float32) {
	if s.Fresh {
		s.refresh(spectrum, s.chosen)
	}
}

func (s *CoarseState) after(spectrum []float32, chosen ...int) {
	s.chosen = append(s.chosen[:0], chosen...)
	if !s.Fresh {
		s.refresh(spectrum, s.chosen)
	}
}
