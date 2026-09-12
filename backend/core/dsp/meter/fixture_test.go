package meter

// fixtureScores marks every scored note below 109 foreground unless foreground is non-nil.
func fixtureScores(scores map[int]float32, foreground []int) *ScoreTable {
	var t ScoreTable
	for note, value := range scores {
		t.Scores[note] = value
		if foreground == nil && note < 109 {
			t.Foreground[note] = true
		}
	}
	for _, note := range foreground {
		t.Foreground[note] = true
	}
	return &t
}

func fixtureUnset() Nomination { return Nomination{Octave: -1, PitchClass: -1} }
