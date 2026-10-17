package audio

import (
	"math"
	"sync"
	"testing"
	"time"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/core/audio/endpoint"
	"smegg.me/smeggtuner/backend/core/dsp"
)

// mixer is an in-memory OS mixer entry for one device.
type mixer struct {
	mu     sync.Mutex
	level  float64
	muted  bool
	closed bool
}

func (m *mixer) Level() (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.level, nil
}

func (m *mixer) SetLevel(v float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.level = v
	return nil
}

func (m *mixer) Muted() (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.muted, nil
}

func (m *mixer) SetMuted(v bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.muted = v
	return nil
}

func (m *mixer) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func steered(t *testing.T) (*Service, *mixer, *int) {
	t.Helper()
	s := testService(t, nil)
	m := &mixer{level: 0.9}
	opened := 0
	s.openEndpoint = func(id string) (endpoint.Endpoint, error) {
		if id != "aa" {
			t.Errorf("opened %q", id)
		}
		opened++
		return m, nil
	}
	return s, m, &opened
}

func waitLevel(t *testing.T, m *mixer, want float64) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		got, _ := m.Level()
		if math.Abs(got-want) < 1e-6 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("level %v, want %v", got, want)
		}
		time.Sleep(time.Millisecond)
	}
}

// interval is a measurement carrying level interval seq at rms percent of full scale.
func interval(seq uint64, rms float32) dsp.Measurement {
	return dsp.Measurement{Input: dsp.InputInterval{Seq: seq, RMS: rms / 100 * 32767, Peak: 100}}
}

func TestAutomaticVolumeSteersACaptureAndGivesTheLevelBack(t *testing.T) {
	s, m, _ := steered(t)
	s.captureStarted("aa")
	waitLevel(t, m, 0.5) // the default stored level
	for seq := uint64(1); seq <= 4; seq++ {
		s.OnMeasurement(interval(seq, 0.1))
		s.OnMeasurement(interval(seq, 0.1)) // every event repeats its interval; only a new Seq counts
	}
	s.OnMeasurement(interval(5, 2))
	waitLevel(t, m, 0.525)
	if st := s.AutoVolume("aa"); !st.On || !st.Active || !st.Supported {
		t.Fatalf("while steering: %+v", st)
	}
	s.captureStopped("aa")
	if got, _ := m.Level(); got != 0.9 || m.muted || !m.closed {
		t.Fatalf("after the capture: level %v muted %v closed %v, want the OS's 0.9", got, m.muted, m.closed)
	}
	if got := device("aa").Volume; math.Abs(got-0.525) > 1e-6 {
		t.Fatalf("stored level %v, want where the capture left it", got)
	}
	if st := s.AutoVolume("aa"); st.Active || math.Abs(st.Level-0.525) > 1e-6 {
		t.Fatalf("after the capture: %+v", st)
	}
}

func TestAutomaticVolumeOffLeavesTheOSAlone(t *testing.T) {
	s, m, opened := steered(t)
	d := appconfig.NewDevice()
	d.AutoVolume = false
	if err := appconfig.SetDevice("aa", d); err != nil {
		t.Fatal(err)
	}
	s.captureStarted("aa")
	s.OnMeasurement(interval(1, 2))
	s.captureStopped("aa")
	if *opened != 0 || m.level != 0.9 {
		t.Fatalf("opened %d, level %v: the OS mixer was touched", *opened, m.level)
	}
	if st := s.AutoVolume("aa"); st.On {
		t.Fatalf("reported on: %+v", st)
	}
}

func TestSwitchingAutomaticVolumeDuringACapture(t *testing.T) {
	s, m, _ := steered(t)
	s.captureStarted("aa")
	waitLevel(t, m, 0.5)
	if st, err := s.SetAutoVolume("", false); err != nil || st.On || st.Active || st.DeviceID != "aa" {
		t.Fatalf("switched off: %+v", st)
	}
	if got, _ := m.Level(); got != 0.9 {
		t.Fatalf("switching off left %v, want the OS's level back at once", got)
	}
	if _, err := s.SetAutoVolume("aa", true); err != nil {
		t.Fatal(err)
	}
	waitLevel(t, m, 0.5)
	s.captureStopped("aa")
	waitLevel(t, m, 0.9)
}

type restoringMixer struct {
	mixer
	restoring chan struct{}
	release   chan struct{}
	once      sync.Once
}

func (m *restoringMixer) SetLevel(v float64) error {
	if v == 0.9 {
		m.once.Do(func() { close(m.restoring); <-m.release })
	}
	return m.mixer.SetLevel(v)
}

func TestAutomaticVolumeRestartWaitsForPreviousRestore(t *testing.T) {
	s := testService(t, nil)
	m := &restoringMixer{mixer: mixer{level: 0.9}, restoring: make(chan struct{}), release: make(chan struct{})}
	opened := make(chan struct{}, 4)
	s.openEndpoint = func(string) (endpoint.Endpoint, error) { opened <- struct{}{}; return m, nil }
	s.captureStarted("aa")
	waitLevel(t, &m.mixer, 0.5)
	<-opened
	off := make(chan struct{})
	go func() {
		if _, err := s.SetAutoVolume("aa", false); err != nil {
			t.Error(err)
		}
		close(off)
	}()
	select {
	case <-m.restoring:
	case <-time.After(time.Second):
		t.Fatal("restore never started")
	}
	on := make(chan struct{})
	go func() {
		if _, err := s.SetAutoVolume("aa", true); err != nil {
			t.Error(err)
		}
		close(on)
	}()
	overlap := false
	select {
	case <-opened:
		overlap = true
	case <-time.After(50 * time.Millisecond):
	}
	close(m.release)
	<-off
	<-on
	waitLevel(t, &m.mixer, 0.5)
	s.captureStopped("aa")
	waitLevel(t, &m.mixer, 0.9)
	if overlap {
		t.Fatal("new endpoint began before prior OS level was restored")
	}
}
