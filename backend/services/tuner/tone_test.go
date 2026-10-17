package tuner

import (
	"errors"
	"math"
	"testing"
	"time"

	coreaudio "smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// The tone sounds where the engine measures: A4 and the transpose move it, the clock correction does not.
func TestToneFrequency(t *testing.T) {
	for _, tc := range []struct {
		a4   float64
		t    int
		note tuning.Note
		want float64
	}{
		{440, 0, 69, 440}, {442, 0, 69, 442}, {440, 3, 69, 440 * math.Exp2(3.0/12)}, {440, 0, 57, 220},
	} {
		got := toneFreq(dsp.EngineConfig{A4: tc.a4, Transpose: tc.t, Calibration: 0.5}, tc.note)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("A4 %v T %d note %d = %v Hz, want %v", tc.a4, tc.t, tc.note, got, tc.want)
		}
	}
}

// silentService is a fresh service whose output device never opens: a tone that gets past every gate
// fails with ErrPlaybackFailed instead of sounding on the machine running the tests.
func silentService(t *testing.T) *Service {
	t.Helper()
	s := freshService(t)
	s.openTone = func() (*coreaudio.TonePlayer, error) { return nil, errors.New("no output in tests") }
	return s
}

// The tone button needs a manual note; neither the mode nor an uncalibrated card refuses a tone.
func TestToneGates(t *testing.T) {
	s := silentService(t)
	wantErr(t, s.ToggleTone(), ErrToneNeedsManual)
	mustOK(t, s.SetMode("octaves"))
	wantErr(t, s.PlayTone(69), ErrPlaybackFailed)
	mustOK(t, s.SetManualNote(69))
	wantErr(t, s.ToggleTone(), ErrPlaybackFailed)
	mustOK(t, s.SetMode("chords"))
	wantErr(t, s.PlayTone(69), ErrPlaybackFailed)
	wantErr(t, s.ToggleTone(), ErrToneNeedsManual)
	wantErr(t, s.PlayTone(5), ErrInvalidNote)
}

// A second press while the tone sounds fades it out; the room counts as ours until the fade ends.
func TestToneToggleStops(t *testing.T) {
	s := freshService(t)
	s.toneUntil.Store(time.Now().Add(time.Minute).UnixNano())
	mustOK(t, s.ToggleTone())
	until := time.Unix(0, s.toneUntil.Load())
	if left := time.Until(until); left <= 0 || left > 600*time.Millisecond {
		t.Fatalf("tone guard ends in %v, want the 0.5 s fade", left)
	}
	if !s.tonePlaying() {
		t.Fatal("the fade is still our tone")
	}
}
