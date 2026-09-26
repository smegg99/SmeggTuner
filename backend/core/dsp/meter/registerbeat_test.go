package meter

import (
	"math"
	"testing"
)

// An octave pair is read at its upper slot: a 16' at -9 under an 8' at 0, against goals -9 and 0, is on goal.
func TestAnOctavePairReadsAtTheUpperPitch(t *testing.T) {
	table := EqualTemperament(440, 0)
	references := []float64{table[48], table[60]}
	deviations, targets := []float64{-9, 0}, []float64{-9, 0}
	for _, hertz := range []bool{false, true} {
		r := NewReadouts(3)
		r.Slots(references, deviations, targets, true, false, hertz)
		pair := r.Panels[1]
		if math.Abs(pair.Current) > 1e-6 {
			t.Fatalf("hertz %v: pair residual = %v, want 0 for both reeds on goal", hertz, pair.Current)
		}
		want := 9.0
		if hertz {
			want = 9 * table[60] * 0.0006
		}
		if math.Abs(pair.Target-want) > 1e-9 {
			t.Fatalf("hertz %v: pair goal = %v, want %v", hertz, pair.Target, want)
		}
	}
}
