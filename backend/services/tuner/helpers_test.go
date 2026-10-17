package tuner

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	goaudio "github.com/go-audio/audio"
	"github.com/go-audio/wav"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
	audiosvc "smegg.me/smeggtuner/backend/services/audio"
)

const fixtureName = "a-8.wav"

// fixtureService returns a tuner already pointed at the a-8 fixture, so no test needs a microphone.
func fixtureService(t *testing.T) *Service {
	t.Helper()
	path := filepath.Join("..", "..", "..", "tests", "fixtures", fixtureName)
	if _, err := os.Stat(path); err != nil {
		t.Skipf("fixture missing: %v", err)
	}
	as := audiosvc.New()
	if err := as.SelectFile(path, true); err != nil {
		t.Fatal(err)
	}
	return New(audiosvc.NewTunerInput(as), nil, nil)
}

// capturedEvent is one event as the frontend would receive it.
type capturedEvent struct {
	name  string
	state StateDTO
	meas  MeasurementDTO
}

// eventLog records everything the service pushes at the frontend by replacing emitEvent, so the captured stream is exactly the UI's.
type eventLog struct {
	mu     sync.Mutex
	events []capturedEvent
}

func (l *eventLog) emit(name string, data any) {
	e := capturedEvent{name: name}
	switch v := data.(type) {
	case StateDTO:
		e.state = v
	case MeasurementDTO:
		e.meas = v
	}
	l.mu.Lock()
	l.events = append(l.events, e)
	l.mu.Unlock()
}

func (l *eventLog) all() []capturedEvent {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]capturedEvent(nil), l.events...)
}

// cadences counts the two measurement cadences: heartbeats between analyses, and the analyses themselves.
func (l *eventLog) cadences() (heartbeats, analyses int) {
	for _, e := range l.all() {
		if e.name != EventMeasurement {
			continue
		}
		if e.meas.Frame == dsp.FrameNone {
			heartbeats++
		} else {
			analyses++
		}
	}
	return heartbeats, analyses
}

// measured is an analysis of note with every slot sounding at the given reference-unit deviations and the visible lock on.
func measured(note tuning.Note, devs ...float64) dsp.Measurement {
	ref := note.Freq(440)
	m := dsp.Measurement{State: dsp.StateRunning, Frame: dsp.FrameMeasured, Note: note, ScalePitch: ref,
		Lock: dsp.Lock{Visible: true, Internal: true, Acquired: true}}
	for _, d := range devs {
		m.Reeds = append(m.Reeds, dsp.ReedMeasure{Present: true, Freq: meter.FrequencyHz(ref, d), Dev: d})
	}
	for i := 1; i < len(m.Reeds); i++ {
		m.Beats = append(m.Beats, dsp.BeatMeasure{Pair: fmt.Sprintf("%d-%d", i, i+1), Present: true,
			Hz: meter.HarmonicBeat(m.Reeds[i-1].Freq, m.Reeds[i].Freq, ref, true)})
	}
	return m
}

// captureEvents swaps the emit seam for the duration of a test.
func captureEvents(t *testing.T, s *Service) *eventLog {
	t.Helper()
	log := &eventLog{}
	prev := emitEvent
	emitEvent = log.emit
	t.Cleanup(func() {
		// Stop first: restoring the seam under a live run would race the run's emit goroutine.
		_ = s.Stop()
		emitEvent = prev
	})
	return log
}

func waitFor(d time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return cond()
}

// spinUntil busy-waits for cond; the ordering test's window is microseconds wide, so a sleeping poll would miss it.
func spinUntil(d time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		runtime.Gosched()
	}
	return cond()
}

// shortWav writes a 0.6 s tone; file playback is real-time paced, so repeated-EOF tests need a short recording.
func shortWav(t *testing.T) string {
	t.Helper()
	const (
		rate  = 48000
		secs  = 0.6
		freq  = 440.0
		level = 0.3
	)
	path := filepath.Join(t.TempDir(), "short.wav")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	enc := wav.NewEncoder(f, rate, 16, 1, 1)
	buf := &goaudio.IntBuffer{
		Format:         &goaudio.Format{NumChannels: 1, SampleRate: rate},
		SourceBitDepth: 16,
		Data:           make([]int, int(rate*secs)),
	}
	for i := range buf.Data {
		buf.Data[i] = int(level * math.MaxInt16 * math.Sin(2*math.Pi*freq*float64(i)/rate))
	}
	if err := enc.Write(buf); err != nil {
		t.Fatal(err)
	}
	if err := enc.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

// initConfig points common/config at a throwaway config file.
func initConfig(t *testing.T) {
	t.Helper()
	t.Setenv("CONFIG_PATH", t.TempDir())
	if _, err := appconfig.Initialize(); err != nil {
		t.Fatal(err)
	}
}
