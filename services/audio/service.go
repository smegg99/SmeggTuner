// Package audio owns the tuner's input selection: a microphone or a WAV
// recording. It holds no DSP state; switching inputs swaps a small DTO.
package audio

import (
	"sync"

	coreaudio "smegg.me/smeggtuner/core/audio"
)

// DeviceDTO is one capture device offered to the UI. ID is an opaque key the UI passes back, never renders.
type DeviceDTO struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

type SourceKind string

const (
	SourceMic  SourceKind = "mic"
	SourceFile SourceKind = "file"
)

// SourceDTO is the current input selection. Kind decides which of DeviceID and Path matters.
type SourceDTO struct {
	Kind     SourceKind `json:"kind"`
	DeviceID string     `json:"deviceId"` // mic only; "" follows the system default
	Path     string     `json:"path"`     // file only
	Loop     bool       `json:"loop"`     // file only
	Name     string     `json:"name"`     // display label
}

// Service owns the input selection. Bound to the frontend by Wails; Build is called by services/tuner.
type Service struct {
	mu      sync.RWMutex
	current SourceDTO

	// decoded recording behind a file selection; nil on the mic
	file *coreaudio.FileSource

	// output device, opened once; nil when this machine has none, tried only once
	speaker      *coreaudio.Speaker
	speakerTried bool
}

// New starts on the system default microphone; the empty Name renders as the UI's "system default" label.
func New() *Service {
	return &Service{current: SourceDTO{Kind: SourceMic}}
}

// Current returns the active selection.
func (s *Service) Current() SourceDTO {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}
