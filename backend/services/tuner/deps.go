// backend/services/tuner/deps.go
package tuner

import (
	"reflect"

	coreaudio "smegg.me/smeggtuner/backend/core/audio"
	coresession "smegg.me/smeggtuner/backend/core/session"
)

// Sessions is the session service as the tuner reads it.
type Sessions interface {
	Goal() coresession.Goal
	SetA4(hz float64) error
	SetTranspose(semitones int) error
}

// Input is the audio input as the tuner reads it.
type Input interface {
	Build() (coreaudio.Source, bool, error)
	SourceName() string
	CorrectionPercent() float64
}

// nilIfTypedNil turns a nil pointer inside an interface into a nil interface, so "no dependency" has one spelling.
func nilIfTypedNil[T any](v T) T {
	if rv := reflect.ValueOf(&v).Elem(); rv.Kind() == reflect.Interface && !rv.IsNil() {
		if inner := rv.Elem(); inner.Kind() == reflect.Pointer && inner.IsNil() {
			var zero T
			return zero
		}
	}
	return v
}
