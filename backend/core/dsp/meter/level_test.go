package meter

import (
	"math"
	"testing"
)

func levelRun(value int16, n int) []int16 {
	out := make([]int16, n)
	for i := range out {
		out[i] = value
	}
	return out
}

func TestLevelLatchNeedsEightEqualSamplesAtFiveHundred(t *testing.T) {
	cases := []struct {
		name  string
		chunk []int16
		want  bool
	}{
		{"eight at 500", levelRun(500, 8), true},
		{"eight at -500", levelRun(-500, 8), true},
		{"seven", levelRun(32767, 7), false},
		{"eight at 499", levelRun(499, 8), false},
		{"a changed value restarts the run", append(levelRun(600, 7), append([]int16{601}, levelRun(601, 6)...)...), false},
		{"a quiet sample ends the run", append(levelRun(600, 4), append([]int16{0}, levelRun(600, 4)...)...), false},
		{"the run after a change", append([]int16{700}, levelRun(600, 8)...), true},
	}
	for _, c := range cases {
		m := NewLevelMeter(20)
		got, done := m.Add(c.chunk)
		if !done {
			got, _ = m.Add(make([]int16, 10))
		}
		if got.Clipped != c.want {
			t.Errorf("%s: clipped %v", c.name, got.Clipped)
		}
	}
}

func TestLevelIntervalsEndOnTheChunkThatReachesHalfASecond(t *testing.T) {
	m := NewLevelMeter(1000) // 500 frames an interval
	if _, done := m.Add(levelRun(1000, 300)); done {
		t.Fatal("300 of 500 frames")
	}
	got, done := m.Add(levelRun(-2000, 300))
	want := float32(math.Sqrt((300*1000*1000 + 300*2000*2000) / 600.0))
	if !done || got.RMS != want || got.Peak != 2000 || !got.Clipped {
		t.Fatalf("first interval %+v want rms %v", got, want)
	}
	// The 100-frame overshoot counts toward the next interval; the latch does not carry.
	if _, done := m.Add(levelRun(3, 399)); done {
		t.Fatal("the overshoot counts toward the next interval")
	}
	got, done = m.Add([]int16{4})
	if !done || got.Clipped || got.Peak != 4 {
		t.Fatalf("second interval %+v", got)
	}
	m.Add(levelRun(9000, 499))
	m.Reset()
	if got, done = m.Add(levelRun(5, 499)); done {
		t.Fatalf("a reset restarts the interval: %+v", got)
	}
}
