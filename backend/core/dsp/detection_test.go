package dsp

import (
	"context"
	"math"
	"sync/atomic"
	"testing"

	"smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// segment is freq Hz at amp for secs; freq 0 is silence.
type segment struct{ freq, amp, secs float64 }

// segmentSource plays segments back to back with a faint noise floor and counts the samples it has handed over.
type segmentSource struct {
	segs []segment
	sent atomic.Int64
}

func (s *segmentSource) Info() audio.SourceInfo {
	return audio.SourceInfo{Name: "segments", SampleRate: 48000}
}
func (s *segmentSource) Stop() error { return nil }
func (s *segmentSource) Start(ctx context.Context) (<-chan audio.Block, error) {
	ch := make(chan audio.Block)
	go func() {
		defer close(ch)
		n, seed := 0, uint32(1)
		for _, sg := range s.segs {
			total := int(sg.secs * 48000)
			for done := 0; done < total; done += 480 {
				b := make([]float32, min(480, total-done))
				for i := range b {
					seed = seed*1664525 + 1013904223
					noise := (float64(seed)/math.MaxUint32 - 0.5) * 0.001
					b[i] = float32(sg.amp*math.Sin(2*math.Pi*sg.freq*float64(n+i)/48000) + noise)
				}
				select {
				case ch <- audio.Block{Samples: b, SampleRate: 48000}:
				case <-ctx.Done():
					return
				}
				n += len(b)
				s.sent.Store(int64(n))
			}
		}
	}()
	return ch, nil
}

// timed is one analysis and when, in seconds of audio, it was published.
type timed struct {
	at float64
	m  Measurement
}

func runSegments(t *testing.T, cfg EngineConfig, segs ...segment) []timed {
	t.Helper()
	src := &segmentSource{segs: segs}
	var out []timed
	e := NewEngine(cfg, func(m Measurement) {
		if m.Frame != FrameNone {
			out = append(out, timed{float64(src.sent.Load()) / 48000, m})
		}
	})
	if err := e.Run(context.Background(), src); err != nil {
		t.Fatal(err)
	}
	return out
}

// firstReading is when note is first measured at or after from, -1 if never.
func firstReading(ms []timed, note tuning.Note, from float64) float64 {
	for _, x := range ms {
		if x.at >= from && x.m.Frame == FrameMeasured && x.m.Note == note {
			return x.at
		}
	}
	return -1
}

const (
	a4, gs4, c5 = tuning.Note(69), tuning.Note(68), tuning.Note(72)
)

func TestALockedNoteHandsOverToTheNextOneWithinHalfASecond(t *testing.T) {
	for _, tc := range []struct {
		name string
		next float64
		want tuning.Note
	}{{"a minor third up", 523.25, c5}, {"a semitone down", 415.3, gs4}} {
		t.Run(tc.name, func(t *testing.T) {
			ms := runSegments(t, EngineConfig{}, segment{440, 0.2, 4}, segment{tc.next, 0.2, 4})
			if at := firstReading(ms, tc.want, 4); at < 0 || at-4 > 0.6 {
				t.Fatalf("%s read %.2f s after it began, want within 0.6 s", tc.want.Name(tuning.NamingCDEFGAB), at-4)
			}
		})
	}
}

func TestSteadyKeepsTheReferenceHandOver(t *testing.T) {
	ms := runSegments(t, EngineConfig{Detection: DetectionSteady}, segment{440, 0.2, 4}, segment{523.25, 0.2, 5})
	if at := firstReading(ms, c5, 4); at-4 < 3 {
		t.Fatalf("C5 read %.2f s after it began; the reference holds the lock and drains the window for about 4 s", at-4)
	}
}

func TestABellowsReversalKeepsTheLock(t *testing.T) {
	ms := runSegments(t, EngineConfig{}, segment{440, 0.2, 4}, segment{0, 0, 0.6}, segment{440, 0.2, 3})
	for _, x := range ms {
		if x.at > 2.5 && x.m.Frame != FrameMeasured {
			t.Fatalf("at %.2f s the lock let go across a 0.6 s gap (%v)", x.at, x.m.Frame)
		}
	}
}

func TestANotePlayedWhileLearningIsStillHeard(t *testing.T) {
	ms := runSegments(t, EngineConfig{LearnOnStart: true}, segment{0, 0, 2}, segment{440, 0.2, 6})
	if at := firstReading(ms, a4, 0); at < 0 {
		t.Fatal("an A4 that began during learning was never read")
	}
}

func TestANoteGoingOnMuchQuieterIsStillRead(t *testing.T) {
	ms := runSegments(t, EngineConfig{LearnOnStart: true}, segment{0, 0, 5}, segment{440, 0.2, 4}, segment{440, 0.004, 5})
	for _, x := range ms {
		if x.at > 7 && (x.m.Frame != FrameMeasured || x.m.Note != a4) {
			t.Fatalf("at %.2f s an A4 fifty times quieter read %v note %d", x.at, x.m.Frame, x.m.Note)
		}
	}
}

func TestAStoppedNoteClearsWithinASecondAndAHalf(t *testing.T) {
	ms := runSegments(t, EngineConfig{LearnOnStart: true}, segment{0, 0, 5}, segment{440, 0.2, 3}, segment{0, 0, 4})
	for _, x := range ms {
		if x.at > 9.5 && x.m.Frame == FrameMeasured {
			t.Fatalf("at %.2f s, 1.5 s after the note stopped, it still reads note %d", x.at, x.m.Note)
		}
	}
}

func TestPresetsSetThePace(t *testing.T) {
	for d, want := range map[Detection]int{DetectionSteady: 2, DetectionBalanced: 8, DetectionFast: 8} {
		ms := runSegments(t, EngineConfig{Detection: d}, segment{440, 0.2, 4})
		if n := len(ms); n < want*4-2 || n > want*4 {
			t.Errorf("%s: %d analyses in 4 s, want %d a second", d, n, want)
		}
	}
	if c := DefaultEngineConfig(); c.Detection != DetectionBalanced {
		t.Errorf("default detection %q, want balanced", c.Detection)
	}
}
