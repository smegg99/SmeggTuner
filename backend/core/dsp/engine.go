package dsp

import (
	"sync"

	"smegg.me/smeggtuner/backend/core/dsp/meter"
)

// Engine runs the reference measurement core over an audio.Source and emits from Run's goroutine.
// All mutable state lives inside Run; the control methods hand work to it over one channel.
type Engine struct {
	cfg       EngineConfig
	emit      func(Measurement)
	ctl       chan func(*runner)
	ctlMu     sync.Mutex
	pending   []func(*runner)
	frozen    bool
	recording bool
	latched   Measurement
	wave      [WaveformPoints]float32
	seed      *meter.Learned // a previous run's background on this input, adopted instead of learning
	learned   *meter.Learned // what this run's analyzer had learned when Run returned
}

// NewEngine builds an engine; emit is called from Run's goroutine and must not block it.
func NewEngine(cfg EngineConfig, emit func(Measurement)) *Engine {
	cfg.fill()
	return &Engine{cfg: cfg, emit: emit, ctl: make(chan func(*runner), 1)}
}

// Seed hands Run a background a previous run learned on the same input; call it before Run.
func (e *Engine) Seed(l *meter.Learned) { e.seed = l }

// Learned is the background the last Run ended with, or nil; read it after Run returns.
func (e *Engine) Learned() *meter.Learned { return e.learned }

// send preserves every control; the channel only wakes Run to drain the ordered queue.
func (e *Engine) send(c func(*runner)) {
	e.ctlMu.Lock()
	e.pending = append(e.pending, c)
	e.ctlMu.Unlock()
	select {
	case e.ctl <- e.applyControls:
	default:
	}
}

func (e *Engine) applyControls(r *runner) {
	e.ctlMu.Lock()
	pending := e.pending
	e.pending = nil
	e.ctlMu.Unlock()
	for _, c := range pending {
		c(r)
		if r.fault != nil {
			return
		}
	}
}

// Update mutates the configuration between analyses; one the capture rate cannot carry ends Run with ErrSampleRate.
func (e *Engine) Update(mut func(*EngineConfig)) {
	e.send(func(r *runner) {
		mut(&e.cfg)
		e.cfg.fill()
		if !e.cfg.fits(r.rate) {
			r.fault = ErrSampleRate
			return
		}
		r.configure(e.cfg)
	})
}

// Freeze latches the last measured result; processing continues, takes stop.
func (e *Engine) Freeze(on bool) { e.send(func(*runner) { e.frozen = on }) }

// Recalibrate starts the eight-update background learning, as on a fresh run.
func (e *Engine) Recalibrate() { e.send(func(r *runner) { r.analyzer.Learn() }) }

// SetRecording turns takes on and the manual-note bypass (Continuous) off, as the reference record index does.
// Turning it off drops a pending TakeNow, so the next pass does not open with an unlocked take.
func (e *Engine) SetRecording(on bool) {
	e.send(func(r *runner) {
		e.recording = on
		r.analyzer.TakeRequest = r.analyzer.TakeRequest && on
	})
}

// TakeNow asks for a take without a lock; the next analysis that may take does, and only a take clears it.
func (e *Engine) TakeNow() { e.send(func(r *runner) { r.analyzer.TakeRequest = true }) }

func (e *Engine) publish(m Measurement) {
	if e.frozen && e.latched.Frame == FrameMeasured {
		held := e.latched
		held.State, held.Waveform, held.InputLevel = StateFrozen, m.Waveform, m.InputLevel
		held.Input, held.Unreliable = m.Input, m.Unreliable
		e.emit(held)
		return
	}
	if m.Frame == FrameMeasured {
		e.latched = m
	}
	e.emit(m)
}
