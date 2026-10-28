package dsp

import (
	"context"
	"math"
	"time"

	"smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
)

// heartbeatsPerSecond paces the meter and waveform between the two analyses a second.
const heartbeatsPerSecond = 20

// Run blocks until ctx is cancelled or the source closes.
func (e *Engine) Run(ctx context.Context, src audio.Source) error {
	blocks, err := src.Start(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = src.Stop() }()
	_, timeline := src.(audio.Transport)
	rate := src.Info().SampleRate
	if !e.cfg.fits(rate) {
		return ErrSampleRate
	}
	r := newRunner(e.cfg, rate, timeline, e.seed)
	defer func() { e.learned = r.analyzer.Learned() }()
	for {
		// Controls go first, so one sent before a block is in force for it.
		select {
		case c := <-e.ctl:
			if c(r); r.fault != nil {
				return r.fault
			}
			continue
		default:
		}
		select {
		case <-ctx.Done():
			return nil
		case c := <-e.ctl:
			if c(r); r.fault != nil {
				return r.fault
			}
		case b, ok := <-blocks:
			if !ok {
				if es, isErr := src.(audio.ErrorSource); isErr {
					if err := es.Err(); err != nil {
						e.emit(r.heartbeat(e, StateDeviceLost))
						return err
					}
				}
				return nil
			}
			e.block(r, b)
		}
	}
}

// runner is one run's capture side: ring, cadence and the analyzer, all owned by Run's goroutine.
type runner struct {
	cfg      EngineConfig
	rate     int
	timeline bool
	analyzer *meter.Analyzer
	ring     *meter.SampleRing
	cadence  meter.CaptureCadence
	perSec   int    // the cadence's passes a second
	every    int    // passes a native event
	passes   int    // cadence passes since the analyzer started; every every-th is a native event (Analyze), the others Remeasure
	ids      []int  // the meter slot each measured slot reads
	pairs    []bool // which neighbouring measured slots beat
	panels   bool   // whether pair readouts show at all
	input    inputState
	fault    error // a control the run cannot carry out; Run returns it
	samples  []int16
	sinceHop int
	hopSum   float64
	hopCount int
	lastEnd  int
	haveLast bool
	srcAt    float64
}

// newRunner learns the background when cfg asks, unless seed is one learned on this input and coarse spectrum.
func newRunner(cfg EngineConfig, rate int, timeline bool, seed *meter.Learned) *runner {
	r := &runner{rate: rate, timeline: timeline, ring: meter.NewSampleRing(meter.RingCapacity), input: newInputState(rate)}
	r.configure(cfg)
	if cfg.LearnOnStart && !r.analyzer.Adopt(seed) {
		r.analyzer.Learn()
	}
	return r
}

// configure starts an analyzer for a new configuration; captured audio, a take request and, with an unchanged
// coarse spectrum, the learned background stay. Otherwise an active learning pass restarts.
func (r *runner) configure(cfg EngineConfig) {
	prev := r.analyzer
	r.cfg, r.ids, r.pairs, r.panels = cfg, cfg.meterSlots(), cfg.pairs(), cfg.pairPanels()
	if prev == nil {
		r.analyzer = meter.NewAnalyzer(cfg.analyzer(r.rate))
	} else {
		r.analyzer = prev.Rebuild(cfg.analyzer(r.rate))
	}
	r.passes = 0 // a new analyzer's first pass selects a note before anything remeasures it
	if d := cfg.detection(); d.perSecond != r.perSec {
		r.cadence, r.perSec, r.every = meter.NewCaptureCadence(r.rate, d.perSecond), d.perSecond, d.perSecond/meter.UpdatesPerSecond
	}
	if prev == nil {
		return
	}
	r.analyzer.TakeRequest = prev.TakeRequest
	if !r.analyzer.Carry(prev) && prev.Learning() {
		r.analyzer.Learn()
	}
}

func (e *Engine) block(r *runner, b audio.Block) {
	if len(b.Samples) == 0 {
		return
	}
	gap := false
	if r.timeline && b.SampleRate > 0 {
		start := int(math.Round(b.At.Seconds() * float64(r.rate)))
		gap = r.haveLast && abs(start-r.lastEnd) > r.rate/10
		r.lastEnd, r.haveLast = start+len(b.Samples), true
		r.srcAt = (b.At + time.Duration(float64(len(b.Samples))/float64(b.SampleRate)*float64(time.Second))).Seconds()
	}
	// A seek is a capture that begins there; the analyzer's deferred reset only restarts the windows.
	if gap {
		r.analyzer.Restart()
		r.passes = 0
	}
	if gap || r.analyzer.ResetPending {
		r.ring.Reset()
		r.cadence.ResetAccumulator()
		r.input.reset()
		r.analyzer.ResetPending = false
	}
	for _, x := range b.Samples {
		r.hopSum += float64(x) * float64(x)
	}
	r.hopCount += len(b.Samples)
	r.samples = meter.QuantizeFloat32(r.samples, b.Samples)
	r.ring.Append(r.samples)
	r.input.add(r.samples, b.Lost, r.rate)
	e.decimateWave(b.Samples)
	r.sinceHop += len(b.Samples)
	if r.cadence.EventPass(len(r.samples), true) {
		r.analyzer.Recording = e.recording
		native := r.passes%r.every == 0
		if r.analyzer.Gone(r.ring) {
			r.analyzer.LetGo(r.ring)
			native, r.passes = true, 0
		}
		var f meter.Frame
		if native {
			f = r.analyzer.Analyze(r.ring)
		} else {
			f = r.analyzer.Remeasure(r.ring)
		}
		r.passes++
		r.cadence.HandlerCompleted()
		e.emit(r.measurement(e, f))
		r.endHop()
		return
	}
	if r.sinceHop >= r.rate/heartbeatsPerSecond {
		e.emit(r.heartbeat(e, r.state()))
		r.endHop()
	}
}

func (r *runner) endHop() { r.sinceHop, r.hopSum, r.hopCount = 0, 0, 0 }

func (r *runner) level() float32 {
	if r.hopCount == 0 {
		return 0
	}
	return float32(math.Sqrt(r.hopSum / float64(r.hopCount)))
}

// state ranks learning first, then a flat-topped level interval, then silence.
func (r *runner) state() EngineState {
	switch {
	case r.analyzer.Learning():
		return StateInitializing
	case r.input.clipped:
		return StateTooLoud
	case r.level() < 1e-3:
		return StateTooQuiet
	}
	return StateRunning
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
