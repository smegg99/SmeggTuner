//go:build !js

// backend/services/tuner/tone.go
package tuner

import (
	"math"
	"sync"
	"time"

	"smegg.me/smeggtuner/backend/common/i18n"

	"smegg.me/smeggtuner/backend/common/logger"
	coreaudio "smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// PlayTone sounds note while the note ruler is held (an extension of the reference tone button) at
// the engine's own pitch for it, A4 and transpose included, with no clock correction. Plays in every
// mode; Dirk's tones-only rule belongs to his Sound button. An uncalibrated sound card plays too.
// Sets toneUntil so observe() refuses to record while our own sine is in the room.
func (s *Service) PlayTone(note int) error {
	n := tuning.Note(note)
	if !n.Valid() {
		return rejected("tone_note", note, ErrInvalidNote)
	}
	cfg, _, _ := s.view()
	return s.sound(n, cfg, true)
}

// ToggleTone is the reference tone button: it sounds the manual note for the tone duration, or fades
// out the tone that is sounding. Needs a manual note, so never sounds in chords mode.
func (s *Service) ToggleTone() error {
	if s.tonePlaying() {
		return s.StopTone()
	}
	cfg, _, _ := s.view()
	if !cfg.ManualNote.Valid() {
		return rejected("tone", "toggle", ErrToneNeedsManual)
	}
	return s.sound(cfg.ManualNote, cfg, false)
}

// StopTone fades the tone out; must be idempotent (also called on pointer-cancel and window blur).
func (s *Service) StopTone() error {
	end := time.Now().Add(coreaudio.ToneFade).UnixNano()
	if until := s.toneUntil.Load(); until > end {
		s.toneUntil.CompareAndSwap(until, end)
	}

	s.toneMu.Lock()
	player := s.tone
	s.toneMu.Unlock()

	if player != nil {
		player.Stop()
	}
	return nil
}

// SetToneVolume sets the tone level 0..1 and returns the clamped value; independent
// of file playback, applied on the callback's next buffer.
func (s *Service) SetToneVolume(v float64) float64 {
	v = min(1, max(0, v))
	s.toneVolBits.Store(math.Float64bits(v))

	s.toneMu.Lock()
	player := s.tone
	s.toneMu.Unlock()
	if player != nil {
		player.SetVolume(v)
	}
	return v
}

// toneFreq is the band table's pitch for the key, so the tone sounds where the engine measures it.
func toneFreq(cfg dsp.EngineConfig, n tuning.Note) float64 {
	return meter.EqualTemperament(cfg.A4, cfg.Transpose)[int(n)-12]
}

// sound plays n; a press is the note strip's tone, which starts at once and is heard however short the
// click. Unlike the reference it plays on an uncalibrated sound card; the app asks for calibration instead.
func (s *Service) sound(n tuning.Note, cfg dsp.EngineConfig, press bool) error {
	player, err := s.tonePlayer()
	if err != nil {
		logger.Warn(i18n.KeyLogTunerTonePlayFailed, logger.Err(err))
		return ErrPlaybackFailed
	}
	// A lazily opened player starts at full, so assert the level.
	player.SetVolume(s.toneVolume())
	// Set the deadline before Play so there is no window where the tone sounds but observe still trusts the room.
	dur := s.toneDur()
	s.toneUntil.Store(time.Now().Add(dur).UnixNano())

	play := player.Play
	if press {
		play = player.Press
	}
	if err := play(toneFreq(cfg, n), dur); err != nil {
		s.toneUntil.Store(0)
		logger.Warn(i18n.KeyLogTunerTonePlayFailed, logger.Err(err))
		return ErrPlaybackFailed
	}
	logger.Debug(i18n.KeyLogTunerTonePlaying, logger.Int("note", int(n)))
	return nil
}

// toneVolume is the current tone level 0..1; starts full, not persisted.
func (s *Service) toneVolume() float64 {
	return math.Float64frombits(s.toneVolBits.Load())
}

// toneDevice is the reference tone's output; the web edition has none.
type toneDevice struct {
	toneMu   sync.Mutex
	tone     *coreaudio.TonePlayer
	openTone func() (*coreaudio.TonePlayer, error) // the output device; tests fail it so nothing sounds
}

func newToneDevice() toneDevice { return toneDevice{openTone: coreaudio.NewTonePlayer} }

// closeTone releases the playback device. Idempotent.
func (s *Service) closeTone() {
	s.toneMu.Lock()
	defer s.toneMu.Unlock()
	if s.tone != nil {
		s.tone.Close()
		s.tone = nil
	}
}

// tonePlayer opens the playback device on first use and keeps it open.
func (s *Service) tonePlayer() (*coreaudio.TonePlayer, error) {
	s.toneMu.Lock()
	defer s.toneMu.Unlock()
	if s.tone != nil {
		return s.tone, nil
	}
	p, err := s.openTone()
	if err != nil {
		return nil, err
	}
	s.tone = p
	return p, nil
}
