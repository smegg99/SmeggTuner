// backend/core/audio/endpoint/endpoint.go

// Package endpoint reaches a capture device's input level and mute in the OS mixer: the PulseAudio
// protocol on Linux (PipeWire's pulse server included), WASAPI endpoint volume on Windows, nothing
// elsewhere.
package endpoint

import (
	"encoding/hex"
	"errors"
	"fmt"
)

// ErrUnsupported reports a build without an endpoint backend.
var ErrUnsupported = errors.New("endpoint volume is not supported on this system")

// Endpoint is one capture device in the OS mixer. Level is on the mixer's own scale, 0..1 where 1
// is the mixer's full level; PulseAudio can report and take more. Not safe for concurrent use.
type Endpoint interface {
	Level() (float64, error)
	SetLevel(float64) error
	Muted() (bool, error)
	SetMuted(bool) error
	Close() error
}

// Supported reports whether this build has a backend.
func Supported() bool { return supported }

// Open reaches the device a capture device id from core/audio names (hex of malgo's id bytes); ""
// is the system's default capture device.
func Open(deviceID string) (Endpoint, error) {
	raw, err := hex.DecodeString(deviceID)
	if err != nil {
		return nil, fmt.Errorf("bad device id: %w", err)
	}
	return open(raw)
}
