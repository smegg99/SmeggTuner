// Package session owns the on-disk store and the one tuning session open at a time. Curve math is
// core/target's, the model core/session's; takes are appended through here by services/record
// because the lock over them lives here. Having no active session is normal, not a degraded mode
// (see Goal); only recording needs one.
//
// Mutations happen on this service's goroutine, off the engine's emit path: each marks the session
// dirty and the saver writes it out. Save and Close flush and report the result themselves.
package session

import (
	"sync"

	coresession "smegg.me/smeggtuner/core/session"
)

// Service owns the store and the active session, bound to the frontend by Wails. Everything it
// hands out is a copy: a take may be appended from the engine's goroutine mid-marshal.
type Service struct {
	mu     sync.RWMutex
	active *coresession.Session
	// register is which switch is pulled; see bench.go.
	register string
	// bassSide says the bench faces the bass keyboard; bassRegister is the pulled bass switch, or
	// empty for the whole (or fixed) machine.
	bassSide     bool
	bassRegister string

	// saveMu serializes flushes, so an older snapshot can never land on disk on top of a newer one.
	saveMu   sync.Mutex
	dirty    chan struct{}
	quit     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

// New builds the service. It reads nothing, so it is safe before the config and datastore exist.
func New() *Service {
	s := &Service{
		dirty: make(chan struct{}, 1),
		quit:  make(chan struct{}),
		done:  make(chan struct{}),
	}
	go s.saver()
	return s
}

// ServiceShutdown is the Wails lifecycle hook: it flushes the active session and stops the saver. Idempotent.
func (s *Service) ServiceShutdown() error {
	err := s.flush()
	s.stopOnce.Do(func() {
		close(s.quit)
		<-s.done
	})
	return err
}
