// Package tuner owns the measurement engine's lifecycle: it runs a dsp.Engine over a
// services/audio source and forwards each Measurement to the frontend joined to the
// active session's goal. No session open is a first-class state, not a fallback.
package tuner

import (
	"math"
	"sync"
	"sync/atomic"

	coreaudio "smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/dsp"
	audiosvc "smegg.me/smeggtuner/services/audio"
	sessionsvc "smegg.me/smeggtuner/services/session"
)

// Service owns the engine lifecycle. Bound to the frontend by Wails.
type Service struct {
	audio *audiosvc.Service
	// sessions is the active session, read on every measurement; nil reads as no session open.
	sessions *sessionsvc.Service
	record   Recorder

	// opMu serializes Start/Stop/Restart; never held while reading a lock-free field, never taken by a setter.
	opMu sync.Mutex

	// mu guards the stored config and the run; never held while calling the session service (emit reads session first).
	mu sync.Mutex
	// cfg is the base config; what the engine runs on is this imposed upon by the active session.
	cfg      dsp.EngineConfig
	imposed  imposed // what the session last made the engine adopt
	frozen   bool    // user intent, re-asserted after every Start
	run      *run
	emitHook func()

	// persistMu serializes the config file's read-modify-write; Wails dispatches setters concurrently.
	persistMu sync.Mutex

	// persistFailed: a config write failed, so the in-memory config is the truth for our fields until it succeeds.
	persistFailed atomic.Bool

	toneMu sync.Mutex
	tone   *coreaudio.TonePlayer

	// toneVolBits is the tone level 0..1 as Float64bits; not persisted, resets to full each launch.
	toneVolBits atomic.Uint64

	// toneUntil is the playing tone's deadline (UnixNano); 0 or past is silence. Atomic: read on the measurement path.
	toneUntil atomic.Int64
}

// New builds the service; nothing is opened here, and sessions and record may be nil.
func New(audio *audiosvc.Service, sessions *sessionsvc.Service, record Recorder) *Service {
	s := &Service{
		audio:    audio,
		sessions: sessions,
		record:   record,
		cfg: dsp.EngineConfig{
			A4:        defaultA4,
			ReedCount: minReeds,
			Highpass:  lowCutAlways,
		},
	}
	s.toneVolBits.Store(math.Float64bits(1))
	return s
}

// Start builds a source and runs the engine in its own goroutine; a no-op if already
// running. Only source construction is reported synchronously; later failure surfaces via EventState.
func (s *Service) Start() error {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	return s.start()
}

// Stop cancels the engine's context and waits for Run to return. Stopping a stopped engine is a no-op.
func (s *Service) Stop() error {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	s.stop()
	return nil
}

// Restart stops and starts the engine on a fresh source (a device or file swap). Freeze
// survives: the engine resets freeze each Run, so the service re-asserts the user's choice.
func (s *Service) Restart() error {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	s.stop()
	return s.start()
}

// IsRunning reports whether an engine is running. A run that has ended and is still retiring is not running.
func (s *Service) IsRunning() bool {
	return s.current() != nil
}

// current returns the live run, or nil when stopped or retiring; callers needing a retiring run read s.run directly.
func (s *Service) current() *run {
	s.mu.Lock()
	r := s.run
	s.mu.Unlock()
	if r == nil || r.dead.Load() {
		return nil
	}
	return r
}

// ServiceShutdown is the Wails lifecycle hook: it stops the engine and releases the tone device. Idempotent.
func (s *Service) ServiceShutdown() error {
	_ = s.Stop()
	s.toneMu.Lock()
	if s.tone != nil {
		s.tone.Close()
		s.tone = nil
	}
	s.toneMu.Unlock()
	return nil
}

// setEmitHookForTest is a test seam run per measurement, so it may be set before or
// after Start. Unexported so Wails does not bind it as an uncallable stub.
func (s *Service) setEmitHookForTest(f func()) {
	s.mu.Lock()
	s.emitHook = f
	s.mu.Unlock()
}
