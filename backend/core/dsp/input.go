package dsp

import "smegg.me/smeggtuner/backend/core/dsp/meter"

// inputState is one run's level intervals, the clip latch behind StateTooLoud and the hold after lost audio.
type inputState struct {
	levels   meter.LevelMeter
	last     InputInterval
	clipped  bool // the last completed interval's latch; a ring reset clears it
	restart  bool
	lossLeft int // samples the unreliable flag still holds for
}

func newInputState(rate int) inputState {
	return inputState{levels: meter.NewLevelMeter(rate), restart: true}
}

// reset is the capture ring reset: the open interval and the clip latch start again.
func (s *inputState) reset() {
	s.levels.Reset()
	s.clipped, s.restart = false, true
}

// add accounts one converted block; lost frames before it hold the unreliable flag for 0.1 s of audio.
func (s *inputState) add(samples []int16, lost, rate int) {
	if lost > 0 {
		s.lossLeft = rate / 10
	} else {
		s.lossLeft = max(0, s.lossLeft-len(samples))
	}
	iv, done := s.levels.Add(samples)
	if !done {
		return
	}
	s.last = InputInterval{Seq: s.last.Seq + 1, RMS: iv.RMS, Peak: iv.Peak, Clipped: iv.Clipped, Restart: s.restart}
	s.clipped, s.restart = iv.Clipped, false
}

func (s *inputState) unreliable() bool { return s.lossLeft > 0 }
