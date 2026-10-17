package audio

import (
	"context"
	"errors"
	"os"
	"testing"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	coreaudio "smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
)

// blockDeviceWrites replaces only this test's config file with a directory.
func blockDeviceWrites(t *testing.T) func() {
	t.Helper()
	path := appconfig.GetConfigPath()
	if err := os.Rename(path, path+".saved"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	restore := func() {
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(path+".saved", path); err != nil {
			t.Fatal(err)
		}
	}
	return restore
}

type idleCalibrationSource struct {
	ctx     context.Context
	stopped bool
}

func (s *idleCalibrationSource) Info() coreaudio.SourceInfo {
	return coreaudio.SourceInfo{SampleRate: 48000}
}
func (s *idleCalibrationSource) Start(ctx context.Context) (<-chan coreaudio.Block, error) {
	s.ctx = ctx
	return make(chan coreaudio.Block), nil
}
func (s *idleCalibrationSource) Stop() error { s.stopped = true; return nil }

func TestCalibrationResetFailureStopsSourceAndPreservesCorrection(t *testing.T) {
	src := &idleCalibrationSource{}
	s := testService(t, src)
	d := appconfig.NewDevice()
	d.Calibration, d.CalibrationPercent = string(CalibrationCalibrated), 0.2
	if err := appconfig.SetDevice("aa", d); err != nil {
		t.Fatal(err)
	}
	_, last := listen(t)
	restore := blockDeviceWrites(t)
	defer restore()
	if err := s.StartCalibration("aa"); !errors.Is(err, ErrCalibrationSave) {
		t.Fatalf("start = %v", err)
	}
	if !src.stopped || src.ctx == nil || src.ctx.Err() != context.Canceled {
		t.Fatal("capture left unmanaged")
	}
	st := waitEnd(t, last)
	if st.Running || st.Error != ErrCalibrationSave.Key || st.Percent != 0.2 {
		t.Fatalf("event %+v", st)
	}
	if got := s.CalibrationState("aa"); got != st {
		t.Fatalf("query %+v differs from event %+v", got, st)
	}
	if got := s.Correction(); got.Missing || got.Percent != 0.2 {
		t.Fatalf("old correction lost: %+v", got)
	}
}

func TestCalibrationFinishFailureDoesNotPublishAcceptedCorrection(t *testing.T) {
	s := testService(t, nil)
	_, last := listen(t)
	c := &calibration{key: "aa", state: CalibrationDTO{DeviceID: "aa", Running: true}}
	s.cal = c
	restore := blockDeviceWrites(t)
	s.finish(c, meter.ClockReport{Status: meter.ClockCalibrated, Value: 0.2}, c.state)
	st := waitEnd(t, last)
	if st.Running || st.Status != CalibrationUncalibrated || st.Error != ErrCalibrationSave.Key || st.Percent != 0 {
		t.Fatalf("failed write announced accepted calibration: %+v", st)
	}
	if got := s.CalibrationState("aa"); got != st {
		t.Fatalf("query %+v differs from event %+v", got, st)
	}
	if got := s.Correction(); !got.Missing {
		t.Fatalf("unsaved correction in use: %+v", got)
	}
	restore()
	// A later successful start clears the remembered error.
	s.openCapture = func(string) (coreaudio.Source, error) { return &idleCalibrationSource{}, nil }
	if err := s.StartCalibration("aa"); err != nil {
		t.Fatal(err)
	}
	if got := s.CalibrationState("aa"); got.Error != "" || !got.Running {
		t.Fatalf("retry %+v", got)
	}
	s.CancelCalibration()
	waitEnd(t, last)
}

func TestAutoVolumeWriteFailureDoesNotChangeSteering(t *testing.T) {
	s, m, _ := steered(t)
	s.captureStarted("aa")
	waitLevel(t, m, 0.5)
	restore := blockDeviceWrites(t)
	st, err := s.SetAutoVolume("aa", false)
	if !errors.Is(err, ErrAutoVolumeSave) || !st.On || !st.Active {
		t.Fatalf("failed switch: %+v, %v", st, err)
	}
	// Saving the last level also fails; restoring the OS level must still happen.
	s.captureStopped("aa")
	if got, _ := m.Level(); got != 0.9 || !m.closed {
		t.Fatalf("OS level not restored: %v closed=%v", got, m.closed)
	}
	if got := device("aa"); !got.AutoVolume || got.Volume != 0.5 {
		t.Fatalf("failed writes changed config: %+v", got)
	}
	restore()
}
