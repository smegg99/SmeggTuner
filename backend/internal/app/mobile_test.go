// backend/internal/app/mobile_test.go
package app

import (
	"errors"
	"reflect"
	"testing"
)

func TestMobileSuspendStopsAudioBeforeSaving(t *testing.T) {
	var calls []string
	failure := errors.New("stop tone")
	var reported []error
	lifecycle := mobileLifecycle{
		cancelCal: func() { calls = append(calls, "calibration") },
		stopTone: func() error {
			calls = append(calls, "tone")
			return failure
		},
		stop:    func() error { calls = append(calls, "capture"); return nil },
		save:    func() error { calls = append(calls, "save"); return nil },
		onError: func(err error) { reported = append(reported, err) },
	}
	lifecycle.suspend()
	if want := []string{"calibration", "tone", "capture", "save"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("suspend calls = %v, want %v", calls, want)
	}
	if len(reported) != 1 || !errors.Is(reported[0], failure) {
		t.Fatalf("reported = %v", reported)
	}
}
