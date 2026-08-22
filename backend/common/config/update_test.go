//go:build !js

// backend/common/config/update_test.go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func initializeUpdateTest(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("CONFIG_PATH", path)
	if _, err := Initialize(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestUpdatePreservesIndependentWritesAndReferences(t *testing.T) {
	path := initializeUpdateTest(t)
	var workers sync.WaitGroup
	errors := make(chan error, 24)
	for range 16 {
		workers.Go(func() {
			errors <- Update(func(c *Config) { c.Tuner.Tolerance++ })
		})
	}
	for i := range 8 {
		workers.Go(func() {
			d := NewDevice()
			d.Calibration, d.CalibrationPercent = "calibrated", 0.125
			errors <- SetDevice(string(rune('a'+i)), d)
		})
	}
	workers.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	if c := Get(); c.Tuner.Tolerance != 17 || len(c.Audio.Devices) != 8 {
		t.Fatalf("lost concurrent edits: tolerance=%v devices=%v", c.Tuner.Tolerance, c.Audio.Devices)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var stored Config
	if err := json.Unmarshal(raw, &stored); err != nil {
		t.Fatal(err)
	}
	if stored.Storage.DBPath != "@{datadir:smeggtuner.db}" {
		t.Fatalf("portable data reference was replaced: %q", stored.Storage.DBPath)
	}
	if _, err := Initialize(); err != nil {
		t.Fatal(err)
	}
	if c := Get(); c.Tuner.Tolerance != 17 || len(c.Audio.Devices) != 8 {
		t.Fatalf("edits did not survive reload: %+v", c)
	}
}

func TestConfigSnapshotDoesNotExposeMutableState(t *testing.T) {
	initializeUpdateTest(t)
	device := NewDevice()
	if err := SetDevice("input", device); err != nil {
		t.Fatal(err)
	}
	copy := Get()
	delete(copy.Audio.Devices, "input")
	copy.Tuner.ReedBoxes[0] = 3
	copy.Tuner.OctaveOffsets[0] = 24
	current := Get()
	if len(current.Audio.Devices) != 1 || current.Audio.Device("input") != device || current.Tuner.ReedBoxes[0] != 1 || current.Tuner.OctaveOffsets[0] != 12 {
		t.Fatalf("snapshot mutation escaped: %+v", current)
	}
}

func TestRejectedUpdateDoesNotChangeMemoryOrDisk(t *testing.T) {
	path := initializeUpdateTest(t)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := Update(func(c *Config) { c.Tuner.A4 = -1 }); err == nil {
		t.Fatal("invalid frequency accepted")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if Get().Tuner.A4 != 440 || string(before) != string(after) {
		t.Fatal("rejected edit changed configuration")
	}
}
