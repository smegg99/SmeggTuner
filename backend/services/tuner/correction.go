package tuner

// correction is the current input's clock correction in percent; with no input there is no device to correct.
func (s *Service) correction() float64 {
	if s.audio == nil {
		return 0
	}
	return s.audio.CorrectionPercent()
}

// AdoptCorrection hands a running engine the current input's clock correction; internal/app calls it
// whenever a sound-card calibration starts or ends.
func (s *Service) AdoptCorrection() {
	p := s.session()
	pct := s.correction()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg.Calibration == pct {
		return
	}
	s.cfg.Calibration = pct
	if r := s.liveLocked(); r != nil {
		r.reconfigure(impose(s.cfg, s.panel, p))
	}
}
