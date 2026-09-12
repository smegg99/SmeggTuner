package meter

import "testing"

func TestStreamCadenceThresholdBoundaryAndOddRate(t *testing.T) {
	c := NewCaptureCadence(44101, UpdatesPerSecond)
	if c.Threshold != 22050 || c.EventPass(22049, true) || !c.EventPass(1, true) || c.Accumulated != 0 {
		t.Fatalf("cadence %+v", c)
	}
}

func TestStreamCadenceBurstSubtractsOneThreshold(t *testing.T) {
	c := NewCaptureCadence(48000, UpdatesPerSecond)
	if !c.EventPass(72000, true) || c.Accumulated != 48000 {
		t.Fatalf("burst %+v", c)
	}
	if c.EventPass(0, true) || c.Accumulated != 24000 {
		t.Fatalf("pending pass %+v", c)
	}
	c.HandlerCompleted()
	if !c.EventPass(0, true) || c.Accumulated != 0 {
		t.Fatalf("after completion %+v", c)
	}
}

func TestStreamCadencePendingDiscardsThreshold(t *testing.T) {
	c := NewCaptureCadence(48000, UpdatesPerSecond)
	if !c.EventPass(24000, true) || c.EventPass(24000, true) || c.Accumulated != 0 {
		t.Fatalf("cadence %+v", c)
	}
	c.HandlerCompleted()
	if c.EventPass(0, true) {
		t.Fatal("completion alone must not post")
	}
}

func TestStreamCadenceMissingWindowRetainsFrames(t *testing.T) {
	c := NewCaptureCadence(48000, UpdatesPerSecond)
	if c.EventPass(48000, false) || c.Accumulated != 48000 {
		t.Fatalf("no target %+v", c)
	}
	if !c.EventPass(0, true) || c.Accumulated != 24000 {
		t.Fatalf("target back %+v", c)
	}
}

func TestStreamCadenceResetAndCompletionAreDistinct(t *testing.T) {
	c := NewCaptureCadence(48000, UpdatesPerSecond)
	c.EventPass(48000, true)
	c.ResetAccumulator()
	if c.Accumulated != 0 || !c.Pending {
		t.Fatalf("reset %+v", c)
	}
	c.HandlerCompleted()
	if c.Pending {
		t.Fatal("completion must clear pending")
	}
}
