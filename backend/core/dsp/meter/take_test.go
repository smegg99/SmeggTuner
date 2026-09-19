package meter

import "testing"

func takeRing(rate int, freqs ...float64) *SampleRing {
	ring := NewSampleRing(RingCapacity)
	analyzerTone(ring, rate, 4, freqs...)
	return ring
}

// One take per lock: the lock rise writes it, the zero tuple restarts history, and the same note is inhibited.
func TestTakeOnLockThenInhibitedUntilTheLabelChanges(t *testing.T) {
	a := NewAnalyzer(analyzerConfig(48000, nil))
	a.Recording = true
	ring := takeRing(48000, 440.5)
	taken := -1
	for i := range 30 {
		f := a.Analyze(ring)
		if !f.Taken {
			continue
		}
		if taken >= 0 {
			t.Fatalf("second take at analysis %d after %d", i, taken)
		}
		taken = i
		if !f.Accepted || !a.History.Inhibited || !a.Display.Recorded || a.History.Countdown != 4 || a.History.Values != [3][8]float64{} {
			t.Fatalf("after the take: %+v recorded %v", a.History, a.Display.Recorded)
		}
	}
	if taken < 0 {
		t.Fatal("a steady note never took")
	}
	if a.History.Locked || a.Display.VisibleLock || !a.Display.Recorded {
		t.Fatalf("the countdown drops the lock, the recorded mark stays: %+v", a.History)
	}
	other := takeRing(48000, 523.5)
	for range 30 {
		if a.Analyze(other).Taken {
			return
		}
	}
	t.Fatal("a new note must clear the inhibit and take again")
}

func TestTakeRequestTakesWithoutALockAndOnlyATakeClearsIt(t *testing.T) {
	a := NewAnalyzer(analyzerConfig(48000, nil))
	a.TakeRequest = true
	ring := takeRing(48000, 440.5)
	for range 3 {
		if a.Analyze(ring).Taken {
			t.Fatal("no take while not recording")
		}
	}
	if !a.TakeRequest {
		t.Fatal("the request waits for a take")
	}
	a.Recording = true
	f := a.Analyze(ring)
	if !f.Taken || a.TakeRequest || a.History.Locked || f.Status != FrameMeasured {
		t.Fatalf("take now: %+v request %v lock %v", f.Status, a.TakeRequest, a.History.Locked)
	}
}

func TestTakeRefusesAPresentSlotBeyondTheDisplayRange(t *testing.T) {
	a := NewAnalyzer(analyzerConfig(48000, nil))
	a.Recording, a.TakeRequest = true, true
	if a.take([]float64{12, 58.5, Missing}) || !a.TakeRequest {
		t.Fatal("58.5 units refuses the take")
	}
	if !a.take([]float64{-58, Missing, 58}) || a.TakeRequest {
		t.Fatal("+-58 and the missing sentinel pass")
	}
	if a.take([]float64{0}) {
		t.Fatal("inhibited after a take")
	}
}

func TestRecordingTurnsTheManualBypassOff(t *testing.T) {
	quiet := NewSampleRing(RingCapacity)
	quiet.Append(make([]int16, 200000))
	for _, recording := range []bool{false, true} {
		a := NewAnalyzer(analyzerConfig(48000, func(c *AnalyzerConfig) { c.Octave, c.PitchClass, c.Continuous = 4, 9, true }))
		a.Recording = recording
		var f Frame
		for range 3 {
			f = a.Analyze(quiet)
		}
		if got := f.Note == 57; got == recording {
			t.Fatalf("recording %v: frame %+v", recording, f)
		}
	}
}

func TestBassTakesAndChordsNever(t *testing.T) {
	bass := NewAnalyzer(analyzerConfig(48000, func(c *AnalyzerConfig) { c.Bands = []Band{{0, 1}, {12, 1}} }))
	bass.Recording, bass.TakeRequest = true, true
	ring := takeRing(48000, 110.1, 220.3)
	f := bass.Analyze(ring)
	for range 3 {
		if f.Taken {
			break
		}
		f = bass.Analyze(ring)
	}
	if !f.Taken || !bass.Display.Recorded || bass.History.Locked {
		t.Fatalf("bass take now: %+v", f)
	}
	chords := NewAnalyzer(analyzerConfig(44100, func(c *AnalyzerConfig) {
		c.Chords, c.Reeds, c.ChordLabel = true, []int{1, 2, 3}, func(ChordIdentity) string { return "C" }
	}))
	chords.Recording, chords.TakeRequest = true, true
	ring = takeRing(44100, 261.63, 329.63, 392.0)
	for range 12 {
		if chords.Analyze(ring).Taken {
			t.Fatal("chords never record")
		}
	}
	if !chords.TakeRequest {
		t.Fatal("chords leave the request pending")
	}
}

func TestRegisterExtensionTakesOverAllSlots(t *testing.T) {
	a := NewAnalyzer(analyzerConfig(48000, func(c *AnalyzerConfig) { c.Bands = []Band{{0, 1}, {12, 2}} }))
	a.Recording = true
	ring := takeRing(48000, 110.1, 219.5, 221.5)
	takes := 0
	for range 30 {
		if a.Analyze(ring).Taken {
			takes++
		}
	}
	if takes != 1 || !a.History.Inhibited {
		t.Fatalf("takes %d inhibited %v", takes, a.History.Inhibited)
	}
}
