package audio

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	coreaudio "smegg.me/smeggtuner/core/audio"
)

func TestSelectFileValidates(t *testing.T) {
	s := New()
	if err := s.SelectFile("/nonexistent/nope.wav", false); err == nil {
		t.Fatal("expected error for missing file")
	}
	if s.Current().Kind != SourceMic {
		t.Fatal("failed selection must not change the current source")
	}
}

func TestSelectFileReportsUnreadable(t *testing.T) {
	s := New()
	err := s.SelectFile("/nonexistent/nope.wav", false)
	if !errors.Is(err, ErrFileUnreadable) {
		t.Fatalf("err = %v, want ErrFileUnreadable", err)
	}
	var se *ServiceError
	if !errors.As(err, &se) || se.Key != "tuner.error.fileUnreadable" {
		t.Fatalf("err key = %+v, want an i18n-keyed ServiceError", se)
	}
}

func TestSelectFileRejectsNonWav(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notaudio.wav")
	if err := os.WriteFile(path, []byte("this is not a wav file"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	s := New()
	if err := s.SelectFile(path, false); !errors.Is(err, ErrFileUnreadable) {
		t.Fatalf("err = %v, want ErrFileUnreadable", err)
	}
	if s.Current().Kind != SourceMic {
		t.Fatal("failed selection must not change the current source")
	}
}

func TestSelectFileAcceptsWav(t *testing.T) {
	wav := filepath.Join("..", "..", "tests", "fixtures", "a-8.wav")
	if _, err := os.Stat(wav); err != nil {
		t.Skipf("fixture missing: %v", err)
	}
	s := New()
	if err := s.SelectFile(wav, true); err != nil {
		t.Fatalf("SelectFile: %v", err)
	}
	cur := s.Current()
	if cur.Kind != SourceFile || !cur.Loop || cur.Name != "a-8.wav" {
		t.Fatalf("current = %+v", cur)
	}
	src, isMic, err := s.Build()
	if err != nil || isMic {
		t.Fatalf("Build: src=%v isMic=%v err=%v", src != nil, isMic, err)
	}
	if src.Info().SampleRate != 48000 {
		t.Fatalf("rate = %d", src.Info().SampleRate)
	}
}

// A mic is single-Start: every Build hands back a fresh one.
func TestBuildReturnsAFreshMic(t *testing.T) {
	s := New()

	first, isMic, err := s.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !isMic {
		t.Fatal("a fresh service is not on the mic")
	}
	second, _, err := s.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if first == second {
		t.Fatal("Build returned the same mic twice; a mic is single-Start")
	}
}

// A file is not single-Start: Build hands back the same transport, playhead and selection intact.
func TestBuildKeepsTheFileTheUserHasBeenSteering(t *testing.T) {
	wav := filepath.Join("..", "..", "tests", "fixtures", "a-8.wav")
	if _, err := os.Stat(wav); err != nil {
		t.Skipf("fixture missing: %v", err)
	}
	s := New()
	if err := s.SelectFile(wav, false); err != nil {
		t.Fatalf("SelectFile: %v", err)
	}

	s.SetRange(0.5, 1.0)
	s.Seek(0.75)

	first, isMic, err := s.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if isMic {
		t.Fatal("a selected file built a mic")
	}
	second, _, err := s.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if first != second {
		t.Fatal("Build decoded the file again; it must hand back the transport the view is driving")
	}

	got := s.Transport()
	if got.From != 0.5 || got.To != 1.0 {
		t.Fatalf("selection %v..%v survived Build as %v..%v", 0.5, 1.0, got.From, got.To)
	}
	if got.Position < 0.5 || got.Position >= 1.0 {
		t.Fatalf("the playhead came out of Build at %v, outside the selection", got.Position)
	}
}

// The mic has no transport, so the file view does not appear.
func TestTheMicHasNoTransport(t *testing.T) {
	if got := New().Transport(); got.Available {
		t.Fatalf("the microphone reported a transport: %+v", got)
	}
}

func TestDefaultsToMic(t *testing.T) {
	s := New()
	cur := s.Current()
	if cur.Kind != SourceMic || cur.DeviceID != "" {
		t.Fatalf("default source = %+v, want system-default mic", cur)
	}
}

func TestSelectMicRejectsUnknownDevice(t *testing.T) {
	if devs, err := coreaudio.Devices(); err != nil || len(devs) == 0 {
		t.Skipf("no capture devices available: err=%v n=%d", err, len(devs))
	}
	s := New()
	err := s.SelectMic("deadbeef")
	if !errors.Is(err, ErrDeviceGone) {
		t.Fatalf("err = %v, want ErrDeviceGone", err)
	}
	if s.Current().DeviceID != "" {
		t.Fatalf("failed selection must not change the current source: %+v", s.Current())
	}
}

// OpenFileDialog without a running Wails app must degrade to "cancelled", never panic.
func TestOpenFileDialogWithoutApp(t *testing.T) {
	s := New()
	path, err := s.OpenFileDialog("Select a WAV recording", "WAV audio")
	if err != nil || path != "" {
		t.Fatalf("OpenFileDialog with no app: path=%q err=%v, want empty and nil", path, err)
	}
}
