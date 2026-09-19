// backend/core/dsp/meter/plan_test.go
package meter

import (
	"reflect"
	"testing"
)

// A rebuilt analyzer takes over the old one's FFT buffers, which held other audio; its frames must be the
// ones a fresh analyzer gives, and a different window size must get its own tables.
func TestARebuiltAnalyzerMeasuresAsAFreshOne(t *testing.T) {
	cfg := analyzerConfig(48000, func(c *AnalyzerConfig) { c.Watch = 3 })
	old := NewAnalyzer(cfg)
	noise := NewSampleRing(RingCapacity)
	analyzerTone(noise, 48000, 4, 311.1)
	for range 4 {
		old.Analyze(noise)
	}

	rebuilt, fresh := old.Rebuild(cfg), NewAnalyzer(cfg)
	if rebuilt.coarsePlan.transform != old.coarsePlan.transform || rebuilt.watch.plans[0].transform != old.watch.plans[0].transform {
		t.Fatal("a rebuild with unchanged windows did not take over the FFT plans")
	}
	ring := NewSampleRing(RingCapacity)
	analyzerTone(ring, 48000, 4, 440.5)
	for i := range 6 {
		if got, want := rebuilt.Analyze(ring), fresh.Analyze(ring); !reflect.DeepEqual(got, want) {
			t.Fatalf("pass %d: rebuilt %+v, fresh %+v", i, got, want)
		}
	}

	other := rebuilt.Rebuild(analyzerConfig(44100, nil))
	if other.coarsePlan.transform == rebuilt.coarsePlan.transform && other.coarsePlan.window.FFTSize != rebuilt.coarsePlan.window.FFTSize {
		t.Fatal("a different window size reused a transform of the wrong size")
	}
}
