package dsp

import (
	"math"

	"smegg.me/smeggtuner/core/tuning"
)

// CoarseResult holds a linear amplitude estimate per note, index = note - tuning.MinNote.
type CoarseResult struct {
	NoteEnergy [tuning.NumNotes]float64
}

// Coarse measures per-note energy with a Goertzel filter at each note's exact frequency. Window
// length per note is min(2.0, max(12 cycles, 0.10 s)) so low notes get enough cycles and high notes
// stay responsive.
type Coarse struct {
	sampleRate int
	a4         float64
	winLen     [tuning.NumNotes]int
	coeff      [tuning.NumNotes]float64
	maxWin     int
	tail       []float64
}

func NewCoarse(sampleRate int, a4 float64) *Coarse {
	c := &Coarse{sampleRate: sampleRate}
	c.SetA4(a4)
	return c
}

func (c *Coarse) SetA4(a4 float64) {
	c.a4 = a4
	const minCycles = 12.0
	sr := float64(c.sampleRate)
	c.maxWin = 0
	for i := 0; i < tuning.NumNotes; i++ {
		f := (tuning.MinNote + tuning.Note(i)).Freq(a4)
		win := minCycles / f
		if win < 0.10 {
			win = 0.10
		}
		if win > 2.0 {
			win = 2.0
		}
		n := int(win * sr)
		c.winLen[i] = n
		if n > c.maxWin {
			c.maxWin = n
		}
		c.coeff[i] = 2 * math.Cos(2*math.Pi*f/sr)
	}
	c.tail = make([]float64, c.maxWin)
}

func (c *Coarse) Analyze(ring *Ring) CoarseResult {
	var res CoarseResult
	n := ring.Tail(c.maxWin, c.tail)
	if n == 0 {
		return res
	}
	data := c.tail[:n]
	for i := 0; i < tuning.NumNotes; i++ {
		win := c.winLen[i]
		if win > n {
			win = n
		}
		seg := data[n-win:]
		var s1, s2 float64
		coeff := c.coeff[i]
		for _, x := range seg {
			s0 := x + coeff*s1 - s2
			s2, s1 = s1, s0
		}
		power := s1*s1 + s2*s2 - coeff*s1*s2
		if power < 0 {
			power = 0
		}
		res.NoteEnergy[i] = 2 * math.Sqrt(power) / float64(win)
	}
	return res
}
