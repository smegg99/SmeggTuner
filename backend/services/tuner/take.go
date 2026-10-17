package tuner

// SetRecording says a record pass is armed in a session: the engine then writes takes and the
// manual-note bypass is off. Re-asserted after every Start; internal/app drives it from record's state.
func (s *Service) SetRecording(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recording = on
	if r := s.liveLocked(); r != nil {
		r.engine.SetRecording(on)
	}
}

// TakeNow asks for a take from the next analysis without waiting for a lock (the Enter key). A no-op
// unless recording, and while our own tone sounds.
func (s *Service) TakeNow() {
	if s.tonePlaying() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if r := s.liveLocked(); r != nil && s.recording {
		r.engine.TakeNow()
	}
}
