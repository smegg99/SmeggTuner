package tuner

import (
	"context"

	"smegg.me/smeggtuner/backend/common/i18n"

	"smegg.me/smeggtuner/backend/common/logger"
	coreaudio "smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
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
		logger.Warn(i18n.KeyLogTunerStartFailed, logger.Err(err))
		return err
	}

	p := s.session()
	ctx, cancel := context.WithCancel(context.Background())
	r := &run{
		cancel:       cancel,
		ctx:          ctx,
		observations: make(chan dsp.Measurement, emitQueue),
		events:       make(chan dsp.Measurement, emitQueue),
		done:         make(chan struct{}),
		isMic:        isMic,
		source:       s.audio.SourceName(),
	}
	s.shown.Store(autoNote)

	// The config, the run and the re-asserted controls are one step under the locks, so a setter lands either before (in cfg) or after (as a control).
	s.persistMu.Lock()
	s.mu.Lock()
	cfg := s.refreshLocked(p)
	// A microphone learns the room's background first, once: later runs on it start from what it learned. A
	// file starts mid-note and learns nothing.
	cfg.LearnOnStart = isMic
	r.engine = dsp.NewEngine(cfg, r.offer)
	if isMic {
		r.engine.Seed(s.learned[r.source])
	}
	s.run = r
	// Freeze and recording are not part of EngineConfig, so Run cannot restore them: re-assert them.
	if s.frozen {
		r.engine.Freeze(true)
	}
	if s.recording {
		r.engine.SetRecording(true)
	}
	s.mu.Unlock()
	s.persistMu.Unlock()

	// UI delivery is lossy and independent of the ordered recording/input-interval observer.
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		for m := range r.events {
			emitEvent(EventMeasurement, decorate(m, s.adopt()))
			s.noteShown(m)
			if hook := s.hook(); hook != nil {
				hook()
			}
		}
	}()

	observed := make(chan struct{})
	go func() {
		defer close(observed)
		r.observe(s.observe)
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

	// Announce the run before the goroutine can end it, or a device that dies as it opens races its own "running" event.
	logger.Info(i18n.KeyLogTunerStarted,
		logger.Str("source", r.source), logger.Bool("mic", isMic),
		logger.Any("a4", cfg.A4), logger.Int("reeds", cfg.ReedCount))
	emitEvent(EventState, StateDTO{Running: true, Source: r.source})

	go func() {
		runErr := r.engine.Run(ctx, src)
		if l := r.engine.Learned(); isMic && l != nil {
			s.mu.Lock()
			if s.learned == nil {
				s.learned = map[string]*meter.Learned{}
			}
			s.learned[r.source] = l
			s.mu.Unlock()
		}
		// Run has returned, so nothing can call offer any more: closing the queue is safe.
		close(r.events)
		close(r.observations)
		<-observed
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
		logger.Debug(i18n.KeyLogTunerEventsDropped, logger.Int64("dropped", dropped))
	}

	state := StateDTO{Running: false, Source: r.source}
	if runErr != nil {
		state.Error = r.errorKey(runErr)
		logger.Warn(i18n.KeyLogTunerRunFailed, logger.Err(runErr), logger.Str("key", state.Error))
	} else {
		logger.Info(i18n.KeyLogTunerStopped)
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
