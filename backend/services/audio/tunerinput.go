// backend/services/audio/tunerinput.go
package audio

import coreaudio "smegg.me/smeggtuner/backend/core/audio"

// TunerInput is this service as services/tuner reads it; a plain value, not a bound service.
type TunerInput struct{ svc *Service }

// NewTunerInput wraps the service for the tuner.
func NewTunerInput(svc *Service) TunerInput { return TunerInput{svc: svc} }

func (t TunerInput) Build() (coreaudio.Source, bool, error) { return t.svc.Build() }
func (t TunerInput) SourceName() string                     { return t.svc.Current().Name }
func (t TunerInput) CorrectionPercent() float64             { return t.svc.Correction().Percent }
