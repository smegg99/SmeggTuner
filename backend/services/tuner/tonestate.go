// backend/services/tuner/tonestate.go
package tuner

import "time"

// tonePlaying reports whether our own reference tone is sounding, so observe() refuses to record it.
func (s *Service) tonePlaying() bool {
	until := s.toneUntil.Load()
	return until != 0 && time.Now().UnixNano() < until
}
