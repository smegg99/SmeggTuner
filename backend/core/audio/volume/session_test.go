// backend/core/audio/volume/session_test.go
package volume

import (
	"errors"
	"testing"
)

// mixer is an in-memory endpoint that records every level it was set to.
type mixer struct {
	level float64
	muted bool
	sets  []float64
	fail  error
}

func (m *mixer) Level() (float64, error) { return m.level, m.fail }
func (m *mixer) Muted() (bool, error)    { return m.muted, m.fail }
func (m *mixer) SetMuted(v bool) error {
	m.muted = v
	return m.fail
}
func (m *mixer) SetLevel(v float64) error {
	if m.fail != nil {
		return m.fail
	}
	m.level = v
	m.sets = append(m.sets, v)
	return nil
}

func TestASessionGivesTheOSItsLevelBack(t *testing.T) {
	m := &mixer{level: 1.3, muted: true}
	s, err := Begin(m, 0.42)
	if err != nil {
		t.Fatal(err)
	}
	if m.level != 0.42 {
		t.Fatalf("begin left the level at %v, want the stored 0.42", m.level)
	}
	m.muted = false // a capture runs unmuted here; the session must not touch mute until the end
	for range 4 {
		if err := s.Interval(Interval{RMS: pct(0.1), Peak: 100}); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Interval(Interval{RMS: pct(2), Peak: 2000}); err != nil {
		t.Fatal(err)
	}
	left, err := s.Left()
	if err != nil {
		t.Fatal(err)
	}
	if left < 0.444 || left > 0.446 {
		t.Fatalf("left at %v, want the raised 0.445", left)
	}
	if err := s.Restore(); err != nil {
		t.Fatal(err)
	}
	if m.level != 1.3 || !m.muted {
		t.Fatalf("restored level %v muted %v, want the OS's 1.3 and muted", m.level, m.muted)
	}
}

func TestAMutedCaptureIsNotSteered(t *testing.T) {
	m := &mixer{level: 0.5}
	s, err := Begin(m, 0.05)
	if err != nil {
		t.Fatal(err)
	}
	m.muted = true
	if err := s.Interval(Interval{RMS: pct(2), Peak: 2000}); err != nil {
		t.Fatal(err)
	}
	if len(m.sets) != 1 {
		t.Fatalf("a muted endpoint was set: %v", m.sets)
	}
}

func TestBeginReportsAnEndpointThatFails(t *testing.T) {
	broken := errors.New("gone")
	if _, err := Begin(&mixer{fail: broken}, 0.5); !errors.Is(err, broken) {
		t.Fatalf("err = %v", err)
	}
}
