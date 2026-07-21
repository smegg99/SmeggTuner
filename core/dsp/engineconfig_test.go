package dsp

import (
	"testing"
	"time"
)

// The engine's defaults are what the detector is calibrated around and the golden fixtures measure
// against, so a change here is not allowed to happen by accident.
func TestDefaultEngineConfig(t *testing.T) {
	c := DefaultEngineConfig()

	cases := []struct {
		name string
		got  any
		want any
	}{
		{"A4", c.A4, 440.0},
		{"ReedCount", c.ReedCount, 1},
		{"FineWindow", c.FineWindow, 3 * time.Second},
		{"LockHold", c.LockHold, 1250 * time.Millisecond},
		{"LockEpsilonHz", c.LockEpsilonHz, 0.1},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("default %s = %v, want %v", tc.name, tc.got, tc.want)
		}
	}
}

// fill defaults only the zero-valued knobs: a caller that set a knob keeps it.
func TestFillKeepsSetKnobs(t *testing.T) {
	c := EngineConfig{
		FineWindow:    2 * time.Second,
		LockHold:      800 * time.Millisecond,
		LockEpsilonHz: 0.25,
		ReedCount:     3,
	}
	c.fill()

	if c.FineWindow != 2*time.Second {
		t.Errorf("FineWindow = %v, want it kept at 2s", c.FineWindow)
	}
	if c.LockHold != 800*time.Millisecond {
		t.Errorf("LockHold = %v, want it kept at 800ms", c.LockHold)
	}
	if c.LockEpsilonHz != 0.25 {
		t.Errorf("LockEpsilonHz = %v, want it kept at 0.25", c.LockEpsilonHz)
	}
	if c.ReedCount != 3 {
		t.Errorf("ReedCount = %v, want it kept at 3", c.ReedCount)
	}
}
