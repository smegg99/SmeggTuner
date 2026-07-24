package dsp

import (
	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/tuning"
)

const msgTrackedNoteChanged logger.MessageID = "tracked note changed"

// Engine runs the full measurement pipeline over an audio.Source and emits Measurement events from
// its run loop goroutine. All mutable state lives inside Run; Update and Freeze hand work to it via
// channels, so no field needs locking.
type Engine struct {
	cfg       EngineConfig
	emit      func(Measurement)
	mutCh     chan func(*EngineConfig)
	frozen    bool
	frzCh     chan bool
	recalibCh chan struct{}
	latched   Measurement
	// whether the level held still across the last fine window; run loop only.
	steady bool
	// the most recent block, thinned for the input strip; an array so decimation costs no allocation. Run loop only.
	wave [WaveformPoints]float32
	// the height Measurement.Spectrum is drawn against, and the note it was measured on. Run loop only.
	specPeak float64
	specFc   float64
	// which key the tracked note resolved to in compound mode, cached while the note holds and
	// re-resolved once a second (compAge counts fine hops). Run loop only.
	compFor  tuning.Note
	compBase tuning.Note
	compAge  int
	// per-band residual-angle trackers for the compound verdicts, reset with the note; sized for
	// the tallest register there is, a six-voice bass machine. Run loop only.
	compTracks [6]bandTrack
}

// NewEngine builds an engine that emits once per coarse hop. While a note sounds the fine stage runs
// on the same cadence and fills in Note, Reeds, Beats and Spectrum. Reedless measurements still flow
// in silence and between notes, so consumers must not assume every measurement carries a fine result.
// emit is called from Run's goroutine and must not block it.
func NewEngine(cfg EngineConfig, emit func(Measurement)) *Engine {
	cfg.fill()
	return &Engine{
		cfg:       cfg,
		emit:      emit,
		mutCh:     make(chan func(*EngineConfig), 8),
		frzCh:     make(chan bool, 8),
		recalibCh: make(chan struct{}, 8),
	}
}

// Update mutates the configuration, serialized against the run loop. The send is non-blocking: only
// Run drains the channel, so a config change made while the engine is stopped is dropped rather than
// parking the caller. The service re-applies its stored config on the next Start.
func (e *Engine) Update(mut func(*EngineConfig)) {
	select {
	case e.mutCh <- mut:
	default:
	}
}

// Freeze latches the last emitted Measurement; processing continues. Non-blocking like Update, but a
// dropped Freeze has no re-apply path, so only call it while Run is active and re-assert it after a restart.
func (e *Engine) Freeze(on bool) {
	select {
	case e.frzCh <- on:
	default:
	}
}

// Recalibrate restarts the noise-floor measurement, so the engine re-enters its quiet warm-up window
// (StateInitializing) as if the run had just begun. Non-blocking like Freeze; a no-op on a source
// whose calibration length is zero (a file). Called when recording arms, to measure the room afresh.
func (e *Engine) Recalibrate() {
	select {
	case e.recalibCh <- struct{}{}:
	default:
	}
}
