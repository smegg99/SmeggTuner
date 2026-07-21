package record

// Armed reports whether locks are being saved.
func (s *Service) Armed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.armed
}

// SetArmed turns recording on and off. It stamps the session the arm was made in
// so the next session:active does not disarm it.
func (s *Service) SetArmed(on bool) {
	s.mu.Lock()
	s.armed = on
	s.lastSession = s.sessionID()
	s.mu.Unlock()
	s.PublishState()
}

// sessionID is the open session's id, or "" when there is none. Callers hold s.mu.
// Lock order is this mutex then the session service's; services/session never calls
// back into here, so it cannot deadlock.
func (s *Service) sessionID() string {
	d, err := s.sessions.Data()
	if err != nil {
		return ""
	}
	return d.SessionID
}

func (s *Service) state() StateDTO {
	d, err := s.sessions.Data()
	if err != nil {
		return StateDTO{}
	}
	s.mu.Lock()
	armed := s.armed
	s.mu.Unlock()

	return StateDTO{
		SessionID: d.SessionID,
		Readings:  len(d.Takes),
		Armed:     armed,
	}
}

// PublishState ships only the state, not the table: the active session changes on
// every curve edit, and shipping the table would reship every reading per drag frame.
func (s *Service) PublishState() {
	emitEvent(EventState, s.state())
}

// SessionChanged re-arms (into warm-up) only on a change of session identity,
// because services/session emits session:active from every mutator. A change of
// session also ships the table; a mutation within one session ships only the state.
func (s *Service) SessionChanged() {
	s.mu.Lock()
	changed := false
	if id := s.sessionID(); id != s.lastSession {
		s.lastSession = id
		s.armed = false
		changed = true
	}
	s.mu.Unlock()

	if changed {
		if _, err := s.publishTable(); err == nil {
			return
		}
	}
	s.PublishState()
}

// publishTable pushes the table and the state beside it.
func (s *Service) publishTable() (*TableDTO, error) {
	emitEvent(EventState, s.state())

	d, err := s.sessions.Data()
	if err != nil {
		return nil, err
	}
	t := s.tableOf(d)
	emitEvent(EventTable, *t)
	return t, nil
}
