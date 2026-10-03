package meter

import "testing"

func TestClockProgressFollowsTheSpreadThenTheTimer(t *testing.T) {
	p := NewClockProgress()
	steps := []struct {
		status ClockStatus
		spread float64
		wall   float64
		want   int
	}{
		{ClockCalibrating, 0.01, 1, 5},     // the widest spread so far: no progress yet, floor 5
		{ClockCalibrating, 0.005, 2, 26},   // 50*(0.01-0.005)/(0.01-0.0005)
		{ClockCalibrating, 0.008, 3, 26},   // a wider spread never takes the bar back
		{ClockCalibrating, 0.001, 4, 47},   // 47.4
		{ClockCalibrating, 0.0004, 10, 50}, // settled: the timer starts at 50
		{ClockCalibrating, 0.0009, 15, 50}, // unsettled again: 47.9 from the spread, the bar holds
		{ClockCalibrating, 0.0003, 20, 73}, // the timer kept running: 50+45*10/20 = 72.5, half up
		{ClockCalibrating, 0.0001, 30, 95},
		{ClockCalibrating, 0.0001, 60, 95}, // capped until the outcome
		{ClockCalibrated, 0, 61, 100},
	}
	for i, s := range steps {
		if got := p.Next(ClockReport{Status: s.status, Spread: s.spread}, s.wall); got != s.want {
			t.Fatalf("step %d: bar %d, want %d", i, got, s.want)
		}
	}
}

func TestClockProgressOutcomes(t *testing.T) {
	p := NewClockProgress()
	if got := p.Next(ClockReport{Status: ClockPreparing}, 1); got != 50 {
		t.Fatalf("preparing with no spread runs the timer from 50, got %d", got)
	}
	if got := p.Next(ClockReport{Status: ClockCalibrated}, 2); got != 0 {
		t.Fatalf("calibrated without ever seeing a spread shows %d, want 0", got)
	}
	p = NewClockProgress()
	p.Next(ClockReport{Status: ClockCalibrating, Spread: 0.002}, 1)
	if got := p.Next(ClockReport{Status: ClockUnreliable}, 2); got != 0 {
		t.Fatalf("unreliable shows %d, want 0", got)
	}
}
