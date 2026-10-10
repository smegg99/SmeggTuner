package dsp

import "testing"

// A steady reed locks once it has stayed the selected note for four native events, a second and a
// half after it is first measured. A synthetic source has no clock position, so the time is counted in analyses.
func TestSteadyReedLocksWithinOneAndAHalfSeconds(t *testing.T) {
	ms := runSynth(t, EngineConfig{}, 8, 440.5)

	analysis, measured, locked := 0, -1, -1
	for _, m := range ms {
		if m.Frame == FrameNone {
			continue
		}
		if measured < 0 && m.Frame == FrameMeasured {
			measured = analysis
		}
		if locked < 0 && m.Lock.Visible {
			locked = analysis
		}
		analysis++
	}
	if measured < 0 || locked < 0 {
		t.Fatalf("measured at analysis %d, locked at %d", measured, locked)
	}
	if gap := float64(locked-measured) / float64(detections[DetectionBalanced].perSecond); gap > 1.5 {
		t.Fatalf("locked %.2f s after the first measurement, want at most 1.5 s", gap)
	}
}

// The engine measures eight times a second, four times Dirk's native analysis rate.
func TestAnalysesRunFasterThanTheDetector(t *testing.T) {
	ms := runSynth(t, EngineConfig{}, 6, 440.5)

	analyses := 0
	for _, m := range ms {
		if m.Frame != FrameNone {
			analyses++
		}
	}
	// Six seconds at eight analyses a second, less the first one the ring cannot fill.
	if analyses < 40 {
		t.Fatalf("%d analyses in six seconds, want eight a second", analyses)
	}
}
