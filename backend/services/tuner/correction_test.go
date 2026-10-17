package tuner

import (
	"testing"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	coreaudio "smegg.me/smeggtuner/backend/core/audio"
)

// storeEveryMic stores d under every key the default mic can resolve to, so the test does not
// depend on which capture devices this machine has.
func storeEveryMic(t *testing.T, calibration string, percent float64) {
	t.Helper()
	d := appconfig.NewDevice()
	d.Calibration, d.CalibrationPercent = calibration, percent
	keys := []string{""}
	if devices, err := coreaudio.Devices(); err == nil {
		for _, dev := range devices {
			keys = append(keys, dev.ID)
		}
	}
	for _, k := range keys {
		if err := appconfig.SetDevice(k, d); err != nil {
			t.Fatal(err)
		}
	}
}

// A recording has no sound card: no correction, and nothing for the tone to wait for.
func TestAFileNeedsNoCorrection(t *testing.T) {
	initConfig(t)
	s := fixtureService(t)
	if c := s.correction(); c != 0 || s.snapshot().Calibration != 0 {
		t.Fatalf("file correction %v, engine %v", c, s.snapshot().Calibration)
	}
}

// An uncalibrated mic measures uncorrected and still sounds the reference tone.
func TestAnUncalibratedMicMeasuresUncorrectedAndPlays(t *testing.T) {
	s := silentService(t)
	if c := s.correction(); c != 0 || s.snapshot().Calibration != 0 {
		t.Fatalf("uncalibrated mic: %v, engine %v", c, s.snapshot().Calibration)
	}
	mustOK(t, s.SetManualNote(69))
	wantErr(t, s.ToggleTone(), ErrPlaybackFailed)
	wantErr(t, s.PlayTone(69), ErrPlaybackFailed)
}

// The engine runs on the mic's own stored correction; an unreliable card is uncorrected but allowed.
func TestTheMicsStoredCorrectionReachesTheEngine(t *testing.T) {
	s := freshService(t)
	storeEveryMic(t, "calibrated", 0.012)
	if c := s.correction(); c != 0.012 {
		t.Fatalf("calibrated mic: %v", c)
	}
	if got := s.snapshot().Calibration; got != 0.012 {
		t.Fatalf("engine calibration %v, want the device's 0.012", got)
	}
	storeEveryMic(t, "unreliable", 0)
	if c := s.correction(); c != 0 {
		t.Fatalf("unreliable mic: %v", c)
	}
}

// A calibration that ends while the tuner holds its config moves the stored config at once.
func TestAdoptCorrectionTakesTheNewValue(t *testing.T) {
	s := freshService(t)
	_ = s.snapshot()
	storeEveryMic(t, "calibrated", -0.25)
	s.AdoptCorrection()
	s.mu.Lock()
	got := s.cfg.Calibration
	s.mu.Unlock()
	if got != -0.25 {
		t.Fatalf("adopted %v, want -0.25", got)
	}
}
