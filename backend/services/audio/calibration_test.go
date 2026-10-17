package audio

import (
	"context"
	"errors"
	"math"
	"sync"
	"testing"
	"time"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	coreaudio "smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
)

// clockSource is a capture whose blocks carry a synthetic device clock: a 10 ms packet of a device
// running offset percent fast, stamped with simulated host time and handed over as fast as read.
type clockSource struct {
	offset  float64
	seconds float64
	lost    func(i int) bool
}

func (c *clockSource) Info() coreaudio.SourceInfo {
	return coreaudio.SourceInfo{Name: "clock", SampleRate: 48000, Realtime: true}
}
func (c *clockSource) Stop() error { return nil }

func (c *clockSource) Start(ctx context.Context) (<-chan coreaudio.Block, error) {
	ch := make(chan coreaudio.Block)
	t0 := time.Now()
	go func() {
		defer close(ch)
		for i := 0; float64(i+1)*0.01 <= c.seconds; i++ {
			t := float64(i+1) * 0.01
			b := coreaudio.Block{SampleRate: 48000, Time: t0.Add(time.Duration(t * float64(time.Second))),
				Frames: int64(math.Floor(48000 * (1 + c.offset/100) * t))}
			if c.lost != nil && c.lost(i) {
				b.Lost = 480
			}
			select {
			case ch <- b:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

// testService is a service with a loaded config in a temp dir, one default device "aa" and the
// capture replaced by src.
func testService(t *testing.T, src coreaudio.Source) *Service {
	t.Helper()
	t.Setenv("CONFIG_PATH", t.TempDir())
	if _, err := appconfig.Initialize(); err != nil {
		t.Fatal(err)
	}
	s := New()
	s.listDevices = func() ([]coreaudio.DeviceInfo, error) {
		return []coreaudio.DeviceInfo{{ID: "bb", Name: "B"}, {ID: "aa", Name: "A", Default: true}}, nil
	}
	s.openCapture = func(string) (coreaudio.Source, error) { return src, nil }
	return s
}

// listen records every calibration event and hands back the one that ends the session.
func listen(t *testing.T) (all func() []CalibrationDTO, last <-chan CalibrationDTO) {
	t.Helper()
	var mu sync.Mutex
	var seen []CalibrationDTO
	done := make(chan CalibrationDTO, 1)
	prev := emitEvent
	emitEvent = func(name string, data any) {
		if st, ok := data.(CalibrationDTO); ok && name == EventCalibration {
			mu.Lock()
			seen = append(seen, st)
			mu.Unlock()
			if !st.Running {
				done <- st
			}
		}
	}
	t.Cleanup(func() { emitEvent = prev })
	return func() []CalibrationDTO {
		mu.Lock()
		defer mu.Unlock()
		return append([]CalibrationDTO(nil), seen...)
	}, done
}

func waitEnd(t *testing.T, last <-chan CalibrationDTO) CalibrationDTO {
	t.Helper()
	select {
	case st := <-last:
		return st
	case <-time.After(10 * time.Second):
		t.Fatal("the session never ended")
		return CalibrationDTO{}
	}
}

func TestACalibrationStoresTheDevicesCorrection(t *testing.T) {
	s := testService(t, &clockSource{offset: 0.0123, seconds: 60})
	all, last := listen(t)
	if err := s.StartCalibration(""); err != nil {
		t.Fatal(err)
	}
	end := waitEnd(t, last)
	if end.DeviceID != "aa" || end.Status != CalibrationCalibrated || end.Percent != 0.012 || end.Progress != 100 {
		t.Fatalf("outcome %+v, want aa calibrated at 0.012", end)
	}
	var preparing, calibrating bool
	for _, st := range all() {
		preparing = preparing || st.Status == CalibrationPreparing
		calibrating = calibrating || (st.Status == CalibrationCalibrating && st.Elapsed > 2 && st.Running)
	}
	if !preparing || !calibrating || end.Elapsed < 22 || end.Elapsed > 24 {
		t.Fatalf("states seen: preparing %v calibrating %v, ended at %d s", preparing, calibrating, end.Elapsed)
	}
	if got := s.CalibrationState("aa"); got.Status != CalibrationCalibrated || got.Percent != 0.012 || got.Running {
		t.Fatalf("stored %+v", got)
	}
	if c := s.Correction(); c.Missing || c.Percent != 0.012 {
		t.Fatalf("the default mic's correction %+v, want the stored 0.012", c)
	}
}

func TestLostAudioMakesTheCalibrationUnreliable(t *testing.T) {
	s := testService(t, &clockSource{offset: 0.01, seconds: 60, lost: func(i int) bool { return i == 999 }})
	_, last := listen(t)
	if err := s.StartCalibration("aa"); err != nil {
		t.Fatal(err)
	}
	if end := waitEnd(t, last); end.Status != CalibrationUnreliable || end.Percent != 0 {
		t.Fatalf("outcome %+v, want unreliable with 0", end)
	}
	if c := s.Correction(); c.Missing || c.Percent != 0 {
		t.Fatalf("an unreliable card runs uncorrected but counts as measured: %+v", c)
	}
}

// Starting forgets the old correction, so a session that never finishes leaves the device uncalibrated.
func TestAnAbandonedCalibrationLeavesTheDeviceUncalibrated(t *testing.T) {
	s := testService(t, &clockSource{offset: 0.01, seconds: 5})
	d := appconfig.NewDevice()
	d.Calibration, d.CalibrationPercent = "calibrated", 0.2
	if err := appconfig.SetDevice("aa", d); err != nil {
		t.Fatal(err)
	}
	_, last := listen(t)
	if err := s.StartCalibration("aa"); err != nil {
		t.Fatal(err)
	}
	if end := waitEnd(t, last); end.Status != CalibrationUncalibrated || end.Running {
		t.Fatalf("a capture that ran out: %+v", end)
	}
	if c := s.Correction(); !c.Missing {
		t.Fatalf("correction %+v, want missing", c)
	}
	if p := device("aa").CalibrationPercent; p != meter.CalibrationMissing {
		t.Fatalf("stored %v, want the missing marker", p)
	}
}

func TestCancelEndsTheSession(t *testing.T) {
	s := testService(t, nil)
	s.openCapture = func(string) (coreaudio.Source, error) {
		return &slowSource{clockSource{offset: 0.01, seconds: 1e6}}, nil
	}
	_, last := listen(t)
	if err := s.StartCalibration("aa"); err != nil {
		t.Fatal(err)
	}
	if st := s.CalibrationState("aa"); !st.Running {
		t.Fatalf("state while running %+v", st)
	}
	s.CancelCalibration()
	if end := waitEnd(t, last); end.Status != CalibrationUncalibrated || end.Running {
		t.Fatalf("cancelled: %+v", end)
	}
	if st := s.CalibrationState("aa"); st.Running {
		t.Fatal("still running after cancel")
	}
}

func TestACalibrationNeedsTheDevice(t *testing.T) {
	s := testService(t, nil)
	s.openCapture = func(string) (coreaudio.Source, error) { return nil, errors.New("unplugged") }
	if err := s.StartCalibration("aa"); !errors.Is(err, ErrDeviceGone) {
		t.Fatalf("err = %v", err)
	}
}

// slowSource paces its packets so a session stays running while the test looks at it.
type slowSource struct{ clockSource }

func (c *slowSource) Start(ctx context.Context) (<-chan coreaudio.Block, error) {
	in, _ := c.clockSource.Start(ctx)
	out := make(chan coreaudio.Block)
	go func() {
		defer close(out)
		for b := range in {
			select {
			case out <- b:
			case <-ctx.Done():
				return
			}
			time.Sleep(time.Millisecond)
		}
	}()
	return out, nil
}

func TestAFileNeedsNoCorrection(t *testing.T) {
	s := testService(t, nil)
	if c := s.Correction(); !c.Missing {
		t.Fatalf("an uncalibrated default mic: %+v", c)
	}
	s.mu.Lock()
	s.current, s.keyed = SourceDTO{Kind: SourceFile, Path: "x.wav"}, false
	s.mu.Unlock()
	if c := s.Correction(); c.Missing || c.Percent != 0 {
		t.Fatalf("a file: %+v", c)
	}
}
