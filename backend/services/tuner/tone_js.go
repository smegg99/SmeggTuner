//go:build js

// backend/services/tuner/tone_js.go
package tuner

// toneDevice is empty in the web edition: there is no reference tone.
type toneDevice struct{}

func newToneDevice() toneDevice { return toneDevice{} }

func (s *Service) closeTone() {}
