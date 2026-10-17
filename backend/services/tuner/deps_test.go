// backend/services/tuner/deps_test.go
package tuner

import (
	"testing"
	"time"

	coreaudio "smegg.me/smeggtuner/backend/core/audio"
	audiosvc "smegg.me/smeggtuner/backend/services/audio"
	sessionsvc "smegg.me/smeggtuner/backend/services/session"
)

var (
	_ Sessions = (*sessionsvc.Service)(nil)
	_ Input    = audiosvc.TunerInput{}
)

type fakeInput struct{}

func (fakeInput) Build() (coreaudio.Source, bool, error) { return nil, false, nil }
func (fakeInput) SourceName() string                     { return "fake" }
func (fakeInput) CorrectionPercent() float64             { return 0 }

func TestNewWithNilSessionsReadsAsNoSession(t *testing.T) {
	var typedNil *sessionsvc.Service
	cases := map[string]Sessions{"nil interface": nil, "typed nil": typedNil}
	for name, sessions := range cases {
		t.Run(name, func(t *testing.T) {
			s := New(fakeInput{}, sessions, nil)
			if got := s.Settings().SessionReeds; got != 0 {
				t.Fatalf("SessionReeds = %d, want 0 with no session service", got)
			}
			if err := s.SetA4(440); err != nil {
				t.Fatalf("SetA4(440) = %v, want nil with no session service", err)
			}
		})
	}
}

// rateInput is an input whose source runs at a rate the engine refuses.
type rateInput struct{ rate int }

func (r rateInput) Build() (coreaudio.Source, bool, error) {
	return coreaudio.NewSynthSource(coreaudio.SynthSpec{SampleRate: r.rate, Duration: time.Second}), true, nil
}
func (rateInput) SourceName() string         { return "rate" }
func (rateInput) CorrectionPercent() float64 { return 0 }

func TestUnsupportedRateSurfacesError(t *testing.T) {
	s := New(rateInput{rate: 1000}, nil, nil)
	log := captureEvents(t, s)
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	got := waitFor(5*time.Second, func() bool {
		for _, e := range log.all() {
			if e.name == EventState && e.state.Error == ErrSampleRate.Key {
				return true
			}
		}
		return false
	})
	if !got {
		t.Fatalf("no state event with %s; events: %+v", ErrSampleRate.Key, log.all())
	}
}
