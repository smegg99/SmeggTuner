// backend/core/audio/endpoint/pulse_linux_test.go
//go:build linux && !android

package endpoint

import (
	"encoding/hex"
	"math"
	"os"
	"testing"

	"github.com/jfreymuth/pulse/proto"

	"smegg.me/smeggtuner/backend/core/audio/volume"
)

// privateServer points the backend at a task-owned pulse server: SMEGGTUNER_PULSE_SERVER is its
// server string and SMEGGTUNER_PULSE_SOURCE a stereo source on it the test may move. Without both
// the test skips; it never talks to a desktop session's server.
func privateServer(t *testing.T) string {
	t.Helper()
	server, source := os.Getenv("SMEGGTUNER_PULSE_SERVER"), os.Getenv("SMEGGTUNER_PULSE_SOURCE")
	if server == "" || source == "" {
		t.Skip("no private pulse server: set SMEGGTUNER_PULSE_SERVER and SMEGGTUNER_PULSE_SOURCE")
	}
	t.Setenv("PULSE_SERVER", server)
	return source
}

// malgoID is a source name as core/audio lists it: the hex of malgo's 256-byte id.
func malgoID(name string) string {
	raw := make([]byte, 256)
	copy(raw, name)
	return hex.EncodeToString(raw)
}

func openSource(t *testing.T, id string) *pulseSource {
	t.Helper()
	ep, err := Open(id)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ep.Close() })
	return ep.(*pulseSource)
}

func TestPulseLevelAndMute(t *testing.T) {
	ep := openSource(t, malgoID(privateServer(t)))
	for _, want := range []float64{0.37, 1, 0.1, 1.25} {
		if err := ep.SetLevel(want); err != nil {
			t.Fatal(err)
		}
		if got, err := ep.Level(); err != nil || math.Abs(got-want) > 1.0/float64(proto.VolumeNorm) {
			t.Fatalf("level %v, err %v, want %v", got, err, want)
		}
	}
	for _, want := range []bool{true, false} {
		if err := ep.SetMuted(want); err != nil {
			t.Fatal(err)
		}
		if got, err := ep.Muted(); err != nil || got != want {
			t.Fatalf("muted %v, err %v, want %v", got, err, want)
		}
	}
}

func TestPulseKeepsTheBalance(t *testing.T) {
	ep := openSource(t, hex.EncodeToString([]byte(privateServer(t))))
	info, err := ep.info()
	if err != nil {
		t.Fatal(err)
	}
	if len(info.ChannelVolumes) < 2 {
		t.Skipf("source has %d channels", len(info.ChannelVolumes))
	}
	vols := make(proto.ChannelVolumes, len(info.ChannelVolumes))
	for i := range vols {
		vols[i] = proto.VolumeNorm / 2
	}
	vols[1] = proto.VolumeNorm / 4
	if err := ep.client.Request(&proto.SetSourceVolume{SourceIndex: proto.Undefined, SourceName: ep.name, ChannelVolumes: vols}, nil); err != nil {
		t.Fatal(err)
	}
	if err := ep.SetLevel(0.8); err != nil {
		t.Fatal(err)
	}
	if info, err = ep.info(); err != nil {
		t.Fatal(err)
	}
	left, right := info.ChannelVolumes[0].Norm(), info.ChannelVolumes[1].Norm()
	if math.Abs(left-0.8) > 1e-4 || math.Abs(right-0.4) > 1e-4 {
		t.Fatalf("channels %v / %v, want 0.8 / 0.4", left, right)
	}
}

func TestPulseDefaultAndUnknownSources(t *testing.T) {
	privateServer(t)
	ep := openSource(t, "")
	if ep.name == "" {
		t.Fatal("the default source has no name")
	}
	if _, err := ep.Level(); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(malgoID("no-such-source")); err == nil {
		t.Fatal("an unknown source opened")
	}
}

func TestOpenRejectsABadID(t *testing.T) {
	if _, err := Open("zz"); err == nil {
		t.Fatal("a non-hex id opened")
	}
}

// The automatic volume's session against a real server: the OS level and mute come back at the end.
func TestPulseSessionGivesTheLevelBack(t *testing.T) {
	ep := openSource(t, malgoID(privateServer(t)))
	if err := ep.SetLevel(0.8); err != nil {
		t.Fatal(err)
	}
	if err := ep.SetMuted(false); err != nil {
		t.Fatal(err)
	}
	s, err := volume.Begin(ep, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := ep.Level(); math.Abs(got-0.5) > 1e-4 {
		t.Fatalf("begin left %v, want the stored level", got)
	}
	if err := s.Interval(volume.Interval{}); err != nil {
		t.Fatal(err)
	}
	if err := s.Restore(); err != nil {
		t.Fatal(err)
	}
	if got, _ := ep.Level(); math.Abs(got-0.8) > 1e-4 {
		t.Fatalf("restored %v, want the OS's 0.8", got)
	}
}
