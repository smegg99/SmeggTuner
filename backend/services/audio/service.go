// Package audio owns the tuner's input: a microphone or a WAV recording, the sound-card clock
// calibration of each microphone and the automatic input volume. It holds no DSP state.
package audio

import (
	"sync"

	coreaudio "smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/audio/endpoint"
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

// Service owns the input selection, the sound-card calibration and the automatic volume. Bound to
// the frontend by Wails; Build and OnMeasurement are called by services/tuner.
type Service struct {
	mu      sync.RWMutex
	current SourceDTO
	// key is the stored key of the selected mic, resolved once per selection; keyed says it is.
	key   string
	keyed bool

	// decoded recording behind a file selection; nil on the mic
	file *coreaudio.FileSource

	// output device, opened once; nil when this machine has none, tried only once
	speaker      *coreaudio.Speaker
	speakerTried bool

	// calOpMu serializes starting sessions; calMu guards the running one and is never held while
	// waiting for it to end.
	calOpMu   sync.Mutex
	calMu     sync.Mutex
	cal       *calibration
	calErrors map[string]string // last persistence failure per device; guarded by calMu

	// volMu guards the capture the automatic volume follows and its steering.
	volOpMu   sync.Mutex
	volMu     sync.Mutex
	capKey    string
	capturing bool
	steer     *steering
	lastSeq   uint64

	// The OS seams, replaced by tests.
	listDevices  func() ([]coreaudio.DeviceInfo, error)
	openCapture  func(deviceID string) (coreaudio.Source, error)
	openEndpoint func(deviceID string) (endpoint.Endpoint, error)
}

// New starts on the system default microphone; the empty Name renders as the UI's "system default" label.
func New() *Service {
	return &Service{
		current:      SourceDTO{Kind: SourceMic},
		listDevices:  coreaudio.Devices,
		openCapture:  func(id string) (coreaudio.Source, error) { return coreaudio.NewMicSource(id) },
		openEndpoint: endpoint.Open,
	}
}

// Current returns the active selection.
func (s *Service) Current() SourceDTO {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// ServiceShutdown is the Wails lifecycle hook: it ends a running calibration and releases the output device.
func (s *Service) ServiceShutdown() error {
	s.CancelCalibration()
	return s.Close()
}
