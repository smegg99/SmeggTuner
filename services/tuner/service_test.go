package tuner

import (
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	audiosvc "smegg.me/smeggtuner/services/audio"
)

func TestStartStopOnFileSource(t *testing.T) {
	s := fixtureService(t)

	var ticks atomic.Int64
	s.setEmitHookForTest(func() { ticks.Add(1) })

	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Stop() })
	if !s.IsRunning() {
		t.Fatal("not running after Start")
	}
	// setters must not deadlock while running
	if err := s.SetA4(442); err != nil {
		t.Fatal(err)
	}
	if err := s.Freeze(true); err != nil {
		t.Fatal(err)
	}
	if err := s.Freeze(false); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(6 * time.Second)
	for ticks.Load() < 3 && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}
	if ticks.Load() < 3 {
		t.Fatalf("only %d measurements in 6s", ticks.Load())
	}

	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}
	if s.IsRunning() {
		t.Fatal("still running after Stop")
	}
	// setters must not deadlock while stopped (they must never reach a non-draining Run loop)
	if err := s.SetA4(440); err != nil {
		t.Fatal(err)
	}
	if err := s.Freeze(true); err != nil {
		t.Fatal(err)
	}
}

func TestSettersValidate(t *testing.T) {
	s := New(audiosvc.New(), nil, nil)
	if err := s.SetA4(429); err == nil {
		t.Fatal("A4 429 must be rejected")
	}
	if err := s.SetA4(451); err == nil {
		t.Fatal("A4 451 must be rejected")
	}
	if err := s.SetManualNote(15); err == nil {
		t.Fatal("note below the tracked range must be rejected")
	}
	if err := s.SetManualNote(0); err != nil {
		t.Fatalf("note 0 (auto) must be accepted: %v", err)
	}
	if err := s.SetTranspose(99); err == nil {
		t.Fatal("transpose 99 must be rejected")
	}
}

// A rejected setter reports an i18n-keyed ServiceError, the one error shape the frontend understands.
func TestSetterErrorsAreKeyed(t *testing.T) {
	s := New(audiosvc.New(), nil, nil)
	var se *ServiceError
	if err := s.SetA4(400); !errors.As(err, &se) || se.Key != "tuner.error.invalidA4" {
		t.Fatalf("err = %v, want a keyed ServiceError", err)
	}
}

// Restart swaps input while running: it must tear the old engine down and come back running on a fresh source.
func TestRestartKeepsRunning(t *testing.T) {
	s := fixtureService(t)
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Stop() })

	var ticks atomic.Int64
	s.setEmitHookForTest(func() { ticks.Add(1) })

	if err := s.Restart(); err != nil {
		t.Fatal(err)
	}
	if !s.IsRunning() {
		t.Fatal("not running after Restart")
	}

	deadline := time.Now().Add(6 * time.Second)
	for ticks.Load() < 3 && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}
	if ticks.Load() < 3 {
		t.Fatalf("only %d measurements in 6s after Restart", ticks.Load())
	}
	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}
}

// Stop on a stopped service and Start on a running one are no-ops: the UI's buttons may be clicked twice.
func TestStopAndStartAreIdempotent(t *testing.T) {
	s := fixtureService(t)
	if err := s.Stop(); err != nil {
		t.Fatalf("Stop while stopped: %v", err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Stop() })
	if err := s.Start(); err != nil {
		t.Fatalf("Start while running: %v", err)
	}
	if !s.IsRunning() {
		t.Fatal("not running after the second Start")
	}
	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := s.Stop(); err != nil {
		t.Fatalf("second Stop: %v", err)
	}
}

// A source that cannot be built fails at Start with the audio service's key, and leaves the service stopped.
func TestStartReportsUnbuildableSource(t *testing.T) {
	as := audiosvc.New()
	// A file valid at selection time can vanish before Start.
	dir := t.TempDir()
	path := filepath.Join(dir, "gone.wav")
	src := filepath.Join("..", "..", "tests", "fixtures", "a-8.wav")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Skipf("fixture missing: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := as.SelectFile(path, false); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	s := New(as, nil, nil)
	if err := s.Start(); err == nil {
		t.Fatal("Start on a vanished file must fail")
	}
	if s.IsRunning() {
		t.Fatal("failed Start must leave the service stopped")
	}
}
