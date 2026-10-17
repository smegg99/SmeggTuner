package tuner

import (
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// SetManualNote pins the tracked note; 0 hands tracking back to the detector. Not persisted; refused
// in chords mode, which picks its own notes.
func (s *Service) SetManualNote(note int) error {
	n := tuning.Note(note)
	if note != autoNote && !n.Valid() {
		return rejected("manual_note", note, ErrInvalidNote)
	}
	_, err := s.manual("manual_note", func(tuning.Note) tuning.Note { return n })
	return err
}

// SetAuto hands note selection to the detector, or leaves it by pinning the note on screen (A4 when
// none is). Returns the pinned note, 0 in automatic.
func (s *Service) SetAuto(on bool) (int, error) {
	return s.manual("auto", func(tuning.Note) tuning.Note {
		if on {
			return autoNote
		}
		return tuning.ManualStart(tuning.Note(s.shown.Load()))
	})
}

// StepNote moves the manual note a semitone up (+1) or down (-1) within E0..C9; from automatic it pins A4.
func (s *Service) StepNote(delta int) (int, error) {
	if delta != 1 && delta != -1 {
		return 0, rejected("note_step", delta, ErrInvalidNote)
	}
	return s.manual("note_step", func(n tuning.Note) tuning.Note { return n.Step(delta) })
}

// manual moves the manual note through pick and returns where it landed.
func (s *Service) manual(setting string, pick func(tuning.Note) tuning.Note) (int, error) {
	var out tuning.Note
	err := s.set(setting, func(c *dsp.EngineConfig, pn *panel, open bool) error {
		next := pick(c.ManualNote)
		if next != autoNote && !open && pn.mode == ModeChords {
			return ErrChordsAuto
		}
		c.ManualNote, out = next, next
		return nil
	}, nil)
	return int(out), err
}

// noteShown remembers the key an analysis named, for SetAuto.
func (s *Service) noteShown(m dsp.Measurement) {
	if m.Frame != dsp.FrameNone && m.Note.Valid() {
		s.shown.Store(int64(m.Note))
	}
}
