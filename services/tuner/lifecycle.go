package tuner

import (
	"context"

	"smegg.me/smeggtuner/common/logger"
	coreaudio "smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/dsp"
)

// start runs with opMu held.
func (s *Service) start() error {
	s.mu.Lock()
	prev := s.run
	s.mu.Unlock()
	if prev != nil {
		if !prev.dead.Load() {
			return nil // already running; a double-clicked button is a no-op
		}
		// The previous run ended on its own but hasn't emitted its terminal event yet; wait, or this Start's Running:true would be overtaken by that stale Running:false. opMu is held throughout.
		<-prev.done
	}

	src, isMic, err := s.audio.Build()
	if err != nil {
		logger.Warn(logger.MsgTunerStartFailed, logger.Err(err))
		return err
	}

	cfg := s.refresh()
	cfg.CalibSecs = fileCalibSecs
	if isMic {
		cfg.CalibSecs = micCalibSecs
	}

	s.mu.Lock()
	frozen := s.frozen
	s.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	r := &run{
		cancel: cancel,
		events: make(chan dsp.Measurement, emitQueue),
		done:   make(chan struct{}),
		isMic:  isMic,
		source: s.audio.Current().Name,
	}
	r.engine = dsp.NewEngine(cfg, r.offer)

	// The engine emits from its DSP goroutine and must never be blocked by it, so the Wails call happens here, one goroutine removed; the hook is read per measurement so a test that sets it after Start still sees this run's stream.
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		var hold holder
		for m := range r.events {
			g, rl := s.adopt()
			emitEvent(EventMeasurement, decorate(hold.filter(m, rl), g))
			// The recorder reads the engine's stream, not the screen's.
			s.observe(m)
			if hook := s.hook(); hook != nil {
				hook()
			}
		}
	}()

	// The needle is driven from here, not the measurement stream, because a paused file emits nothing and a seeked one must still move; it is waited for like the emitter, so it cannot outlive the terminal EventState.
	playhead := make(chan struct{})
	if t, ok := src.(coreaudio.Transport); ok {
		go func() {
			defer close(playhead)
			followPlayhead(ctx, t)
		}()
	} else {
		close(playhead) // a microphone has no playhead
	}

	s.mu.Lock()
	s.run = r
	s.mu.Unlock()

	// Freeze is not part of EngineConfig, so Run cannot restore it: re-assert it.
	if frozen {
		r.engine.Freeze(true)
	}

	// Announce the run before the goroutine can end it, or a device that dies as it opens races its own "running" event.
	logger.Info(logger.MsgTunerStarted,
		logger.Str("source", r.source), logger.Bool("mic", isMic),
		logger.Any("a4", cfg.A4), logger.Int("reeds", cfg.ReedCount))
	emitEvent(EventState, StateDTO{Running: true, Source: r.source})

	go func() {
		runErr := r.engine.Run(ctx, src)
		// Run has returned, so nothing can call offer any more: closing the queue is safe.
		close(r.events)
		<-drained
		cancel()
		<-playhead // the needle stops before the engine announces that it has stopped
		s.finish(r, runErr)
	}()
	return nil
}

// stop runs with opMu held; it never holds mu while waiting, so the engine's goroutine can take mu to retire itself. A nil s.run means finish already closed done.
func (s *Service) stop() {
	s.mu.Lock()
	r := s.run
	s.mu.Unlock()
	if r == nil {
		return
	}
	r.cancel()
	<-r.done
}

// finish retires a run; the order is the point: dead first so IsRunning stops treating it as live, s.run cleared last so stop can still find it and wait, and between those two this goroutine is the only emitter.
func (s *Service) finish(r *run, runErr error) {
	r.dead.Store(true)

	if dropped := r.drops.Load(); dropped > 0 {
		logger.Debug(logger.MsgTunerEventsDropped, logger.Int64("dropped", dropped))
	}

	state := StateDTO{Running: false, Source: r.source}
	if runErr != nil {
		state.Error = r.errorKey(runErr)
		logger.Warn(logger.MsgTunerRunFailed, logger.Err(runErr), logger.Str("key", state.Error))
	} else {
		logger.Info(logger.MsgTunerStopped)
	}
	emitEvent(EventState, state)

	close(r.done)

	s.mu.Lock()
	if s.run == r {
		s.run = nil
	}
	s.mu.Unlock()
}

// hook returns the test seam's callback, read fresh per measurement so a hook installed after Start still sees the current run's stream.
func (s *Service) hook() func() {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.emitHook
}
