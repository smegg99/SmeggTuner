package app

import (
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
)

type seen struct{ n int }

func (s *seen) OnMeasurement(dsp.Measurement) { s.n++ }

func TestObserversReachEveryFollower(t *testing.T) {
	a, b := &seen{}, &seen{}
	observers{a, b}.OnMeasurement(dsp.Measurement{})
	if a.n != 1 || b.n != 1 {
		t.Fatalf("followers saw %d and %d measurements, want 1 each", a.n, b.n)
	}
}
