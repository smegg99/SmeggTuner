package meter

import "math"

// NoteTable holds reference frequencies by note index (MIDI-12; 57 is A4), without any storage offset.
type NoteTable [110]float64

// EqualTemperament is the reference table for a given A4 and transposition: band n sounds at note n+transpose.
// Index 0 is 0 Hz.
func EqualTemperament(a4 float64, transpose int) NoteTable {
	var t NoteTable
	for i := 1; i < len(t); i++ {
		t[i] = a4 * math.Pow(2, float64(i+transpose-57)/12)
	}
	return t
}

// NoteBand is one note's search band in a transform; frequencies exclude the common table offset.
type NoteBand struct {
	ReferenceHz  float64
	ReferenceBin float64
	NearestBin   int
	LowerBin     int
	UpperBin     int
}

// NewNoteBand is defined for 0 < lower < note < upper and a positive even fftSize; a zero rateFactor counts as 1.
func NewNoteBand(lower, note, upper float64, fftSize, sampleRate int, rateFactor float64) NoteBand {
	if rateFactor == 0 {
		rateFactor = 1
	}
	half := float64(fftSize / 2)
	clock := float64(sampleRate) * rateFactor
	reference := (float64(half*note) + float64(half*note)) / clock
	return NoteBand{
		ReferenceHz:  note,
		ReferenceBin: reference,
		NearestBin:   bandRound(reference),
		LowerBin:     bandRound((note + lower) * half / clock),
		UpperBin:     bandRound((note+upper)*half/clock) - 1,
	}
}

func (b NoteBand) BinCount() int { return b.UpperBin - b.LowerBin + 1 }

// LockSpanUnits is the widest history span that still locks; a nonpositive count skips the test.
func (b NoteBand) LockSpanUnits() float64 { return 400 / float64(b.BinCount()) }

// LockSpanHz is LockSpanUnits in hertz, a deduction rather than a displayed tolerance.
func (b NoteBand) LockSpanHz() float64 { return float64(b.ReferenceHz * 0.0006 * b.LockSpanUnits()) }

func bandRound(x float64) int { return int(math.Trunc(x + 0.5)) }
