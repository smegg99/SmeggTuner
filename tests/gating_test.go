package tests

import (
	"context"
	"math"
	"reflect"
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// partial is one steady sine of a scripted phase.
type partial struct{ freq, amp float64 }

// phase is a stretch of scripted audio; phases play back to back with continuous oscillators.
type phase struct {
	seconds  float64
	partials []partial
}

// scriptSource plays its phases as one continuous capture, as a microphone hears a room and then a reed.
type scriptSource struct{ phases []phase }

func (s scriptSource) Info() audio.SourceInfo {
	return audio.SourceInfo{Name: "script", SampleRate: 48000}
}
func (s scriptSource) Stop() error { return nil }

func (s scriptSource) Start(ctx context.Context) (<-chan audio.Block, error) {
	ch := make(chan audio.Block)
	go func() {
		defer close(ch)
		phases := map[float64]float64{}
		for _, p := range s.phases {
			total := int(p.seconds * 48000)
			for at := 0; at < total; at += 4096 {
				b := make([]float32, min(4096, total-at))
				for i := range b {
					var v float64
					for _, q := range p.partials {
						v += q.amp * math.Sin(phases[q.freq])
						phases[q.freq] += 2 * math.Pi * q.freq / 48000
					}
					b[i] = float32(v)
				}
				select {
				case ch <- audio.Block{Samples: b, SampleRate: 48000}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ch, nil
}

// The strayLine mechanism: unlearned, a weak room line under a weak 16' fundamental lets only the 3rd partial win.
func TestLearningTheRoomKeepsAStrayLineFromNamingTheNote(t *testing.T) {
	t.Parallel()
	want := tuning.Note(54) // F#3
	f := want.Freq(442)
	room := partial{120, 0.003}
	reed := []partial{room, {f, 0.05}, {2 * f, 0.04}, {3 * f, 0.15}}
	unlearned := runSource(t, dsp.EngineConfig{A4: 442}, scriptSource{[]phase{{8, reed}}}).all
	if !reads(unlearned, want+19) {
		t.Errorf("unlearned reads %s, want the 3rd partial %s", summary(unlearned), name(want+19))
	}
	learned := runSource(t, dsp.EngineConfig{A4: 442, LearnOnStart: true}, scriptSource{[]phase{{5, []partial{room}}, {8, reed}}}).all
	m := firstTake(t, learned)
	if m.Note != want || math.Abs(m.Reeds[0].Freq-f) > 0.05 {
		t.Fatalf("after learning the room took %s %+v, want %s at %.3f Hz", name(m.Note), m.Reeds[0], name(want), f)
	}
}

// StopAfterLock: the first take is the reading, every later analysis is held on its panels.
func TestStopAfterLockHoldsTheFirstTake(t *testing.T) {
	t.Parallel()
	spec := audio.SynthSpec{Duration: 10 * time.Second, Reeds: []audio.ReedSpec{{Freq: 440.5, Amp: 0.3}}}
	free := runSynth(t, spec, dsp.EngineConfig{A4: 440})
	held := runSynth(t, spec, dsp.EngineConfig{A4: 440, StopAfterLock: true})
	at, ref := takeAt(held), takeAt(free)
	if len(at) != 1 || len(ref) == 0 {
		t.Fatalf("gated run filed %d takes, free-running %d: want exactly one gated", len(at), len(ref))
	}
	take := held[at[0]]
	if at[0] != ref[0] || !reflect.DeepEqual(take.Reeds, free[ref[0]].Reeds) {
		t.Fatalf("gating changed the take: event %d %+v, free-running event %d %+v", at[0], take.Reeds, ref[0], free[ref[0]].Reeds)
	}
	kinds := map[dsp.FrameKind]int{}
	for _, m := range held[at[0]+1:] {
		if m.Frame == dsp.FrameNone {
			continue
		}
		kinds[m.Frame]++
		if !reflect.DeepEqual(m.Panels, take.Panels) {
			t.Fatalf("a held analysis changed the panels: %+v, took %+v", m.Panels, take.Panels)
		}
	}
	if kinds[dsp.FrameHeld] == 0 || kinds[dsp.FrameMeasured] != 0 {
		t.Fatalf("after the take: %v, want only held analyses", kinds)
	}
	if len(measured(free)) <= len(measured(held)) {
		t.Fatalf("free-running measured %d analyses, gated %d: the gate never closed", len(measured(free)), len(measured(held)))
	}
}

// A manual note is measured even when a louder note would win the automatic scan.
func TestAManualNoteOverridesTheScan(t *testing.T) {
	t.Parallel()
	spec := audio.SynthSpec{Duration: 8 * time.Second, Reeds: []audio.ReedSpec{{Freq: 440, Amp: 0.01}, {Freq: tuning.Note(76).Freq(440), Amp: 0.4}}}
	if all := runSynth(t, spec, dsp.EngineConfig{A4: 440}); !reads(all, 76) {
		t.Fatalf("the automatic scan reads %s, want the louder E5", summary(all))
	}
	m := firstTake(t, runSynth(t, spec, dsp.EngineConfig{A4: 440, ManualNote: tuning.NoteA4}))
	if m.Note != tuning.NoteA4 || math.Abs(m.Reeds[0].Freq-440) > 0.05 {
		t.Fatalf("manual A4 took %s %+v", name(m.Note), m.Reeds[0])
	}
}
