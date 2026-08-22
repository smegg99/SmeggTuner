//go:build !js

package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestInitializeWritesAndLoadsDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	t.Setenv("CONFIG_PATH", path)

	resolved, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if resolved != path {
		t.Fatalf("resolved path = %q, want %q", resolved, path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("default config not written: %v", err)
	}

	cfg := Get()
	if cfg.Preferences.Theme != "auto" {
		t.Errorf("default theme = %q, want auto", cfg.Preferences.Theme)
	}
	if cfg.Preferences.AccentMode != "auto" {
		t.Errorf("default accent mode = %q, want auto", cfg.Preferences.AccentMode)
	}
	if cfg.Preferences.AccentColor != "#2563eb" {
		t.Errorf("default accent = %q, want #2563eb", cfg.Preferences.AccentColor)
	}
	if !cfg.Preferences.CloseToTray {
		t.Error("default close_to_tray should be true")
	}
	if cfg.Logger.Level != "INFO" {
		t.Errorf("default logger level = %q, want INFO", cfg.Logger.Level)
	}
}

func TestTunerDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	t.Setenv("CONFIG_PATH", path)

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	cfg := Get()
	if cfg.Tuner.A4 != 440.0 {
		t.Fatalf("a4 default = %v", cfg.Tuner.A4)
	}
	if cfg.Tuner.DeviationUnit != "reference" || cfg.Tuner.BeatUnit != "hz" {
		t.Fatalf("unit defaults: %+v", cfg.Tuner)
	}
	if len(cfg.Audio.Devices) != 0 || cfg.Audio.Device("any") != NewDevice() {
		t.Fatalf("audio defaults: %+v", cfg.Audio)
	}
	if cfg.Audio.DeviceID != "" || cfg.Audio.HumFilter50 || cfg.Audio.HumFilter60 {
		t.Fatalf("audio defaults: %+v", cfg.Audio)
	}
	if cfg.Tuner.ScaleNaming != "cdefgab" || cfg.Tuner.ErrorReference != "scale" {
		t.Fatalf("tuner defaults: %+v", cfg.Tuner)
	}
	if cfg.Tuner.StopAfterLock || cfg.Tuner.ContinuousUpdateManual {
		t.Fatalf("tuner defaults: %+v", cfg.Tuner)
	}
	// Must match core/target defaults, or a no-config-file tuner would judge differently.
	if cfg.Tuner.Tolerance != 1.0 || cfg.Tuner.BeatTolerance != 3.0 {
		t.Fatalf("tolerance defaults: %+v", cfg.Tuner)
	}
	if cfg.Report.CompanyName != "" || cfg.Report.CompanyAddress != "" || cfg.Report.CompanyWebsite != "" {
		t.Fatalf("report defaults: %+v", cfg.Report)
	}
	if cfg.Tuner.ToneDurationMs != 43000 || cfg.Tuner.CalibrationCaptureMs != 2000 {
		t.Fatalf("tuner timing defaults: tone=%d calib=%d", cfg.Tuner.ToneDurationMs, cfg.Tuner.CalibrationCaptureMs)
	}
	if cfg.Tuner.Mode != "tones" || !slices.Equal(cfg.Tuner.ReedBoxes, []int{1}) || !slices.Equal(cfg.Tuner.OctaveOffsets, []int{12}) {
		t.Fatalf("panel defaults: mode %q reeds %v octaves %v", cfg.Tuner.Mode, cfg.Tuner.ReedBoxes, cfg.Tuner.OctaveOffsets)
	}
	if cfg.Tuner.Transpose != 0 || cfg.Tuner.LockA4 {
		t.Fatalf("transpose %d / lock_a4 %v, want 0 / false", cfg.Tuner.Transpose, cfg.Tuner.LockA4)
	}
}

// The reed and octave boxes can never be empty, and transpose stays inside one octave.
func TestPanelRanges(t *testing.T) {
	for _, tc := range []struct {
		tuner string
		ok    bool
	}{
		{`"reed_boxes":[2,3]`, true}, {`"reed_boxes":[]`, false}, {`"reed_boxes":[4]`, false},
		{`"octave_offsets":[24]`, true}, {`"octave_offsets":[12,24]`, true}, {`"octave_offsets":[]`, false}, {`"octave_offsets":[0]`, false},
		{`"transpose":11`, true}, {`"transpose":12`, false}, {`"transpose":-1`, false},
		{`"mode":"chords"`, true}, {`"mode":"bass"`, false},
	} {
		path := filepath.Join(t.TempDir(), "config.json")
		if err := os.WriteFile(path, []byte(`{"tuner":{`+tc.tuner+`}}`), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Setenv("CONFIG_PATH", path)
		if _, err := Initialize(); (err == nil) != tc.ok {
			t.Errorf("%s: err = %v, want ok=%v", tc.tuner, err, tc.ok)
		}
	}
}

// A box list written by a setter replaces the stored one; it is not merged element by element.
func TestPatchReplacesBoxLists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"tuner":{"reed_boxes":[1,2,3]}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", path)
	if _, err := Initialize(); err != nil {
		t.Fatal(err)
	}
	c := *Get()
	c.Tuner.ReedBoxes = []int{2}
	if err := SetConfig(c); err != nil {
		t.Fatal(err)
	}
	if got := Get().Tuner.ReedBoxes; !slices.Equal(got, []int{2}) {
		t.Fatalf("reed boxes = %v, want [2]", got)
	}
}

// A partial config loads: CUE fills every missing key from schema defaults.
func TestPartialConfigFillsFromDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	partial := `{"preferences":{"theme":"dark"},"tuner":{"a4":442}}`
	if err := os.WriteFile(path, []byte(partial), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", path)

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize on a partial config: %v", err)
	}
	cfg := Get()
	if cfg.Tuner.A4 != 442 {
		t.Errorf("the user's a4 was lost: %v", cfg.Tuner.A4)
	}
	if cfg.Tuner.DeviationUnit != "reference" || cfg.Audio.Device("x") != NewDevice() {
		t.Errorf("missing keys did not fill from defaults: %+v %+v", cfg.Tuner, cfg.Audio)
	}
	if cfg.Tuner.CalibrationCaptureMs != 2000 {
		t.Errorf("tuner timing did not fill: %d", cfg.Tuner.CalibrationCaptureMs)
	}
}

func TestDeviceRanges(t *testing.T) {
	for _, tc := range []struct {
		device string
		ok     bool
	}{
		{`{"calibration_percent":-500}`, true}, {`{"calibration_percent":499.999}`, true},
		{`{"calibration_percent":500.001}`, false}, {`{"calibration_percent":-600}`, false},
		{`{"calibration_percent":-999999}`, true}, {`{"calibration_percent":-999998}`, false},
		{`{"calibration":"unreliable"}`, true}, {`{"calibration":"preparing"}`, false},
		{`{"volume":1}`, true}, {`{"volume":1.01}`, false},
	} {
		path := filepath.Join(t.TempDir(), "config.json")
		if err := os.WriteFile(path, []byte(`{"audio":{"devices":{"abc":`+tc.device+`}}}`), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Setenv("CONFIG_PATH", path)
		if _, err := Initialize(); (err == nil) != tc.ok {
			t.Errorf("device %s: err = %v, want ok=%v", tc.device, err, tc.ok)
		}
	}
}

// A listed device fills its missing keys from the schema, and those defaults are NewDevice's.
func TestAPartialDeviceTakesNewDevicesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"audio":{"devices":{"abc":{"auto_volume":false}}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", path)
	if _, err := Initialize(); err != nil {
		t.Fatal(err)
	}
	d := Get().Audio.Device("abc")
	if d.AutoVolume {
		t.Fatal("the stored auto_volume was lost")
	}
	d.AutoVolume = true
	if d != NewDevice() {
		t.Fatalf("schema defaults %+v differ from NewDevice %+v", d, NewDevice())
	}
}

func TestSetDeviceKeepsTheOtherDevices(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("CONFIG_PATH", path)
	if _, err := Initialize(); err != nil {
		t.Fatal(err)
	}
	a := NewDevice()
	a.Calibration, a.CalibrationPercent = "calibrated", 0.012
	b := NewDevice()
	b.Volume = 0.3
	if err := SetDevice("a", a); err != nil {
		t.Fatal(err)
	}
	if err := SetDevice("b", b); err != nil {
		t.Fatal(err)
	}
	loader = nil
	Global = Config{}
	if _, err := Initialize(); err != nil {
		t.Fatal(err)
	}
	if got := Get().Audio; got.Device("a") != a || got.Device("b") != b {
		t.Fatalf("devices after reload: %+v", got.Devices)
	}
}

func TestSetPreferencesPersists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	t.Setenv("CONFIG_PATH", path)

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	prefs := Get().Preferences
	prefs.Theme = "dark"
	prefs.AccentColor = "#ff8800"
	prefs.Language = "pl"
	if err := SetPreferences(prefs); err != nil {
		t.Fatalf("SetPreferences: %v", err)
	}

	if got := Get().Preferences.Theme; got != "dark" {
		t.Errorf("in-memory theme = %q, want dark", got)
	}
	if got := Get().Preferences.AccentColor; got != "#ff8800" {
		t.Errorf("in-memory accent = %q, want #ff8800", got)
	}

	loader = nil
	Global = Config{}
	if _, err := Initialize(); err != nil {
		t.Fatalf("re-Initialize: %v", err)
	}
	if got := Get().Preferences.Theme; got != "dark" {
		t.Errorf("persisted theme = %q, want dark", got)
	}
	if got := Get().Preferences.Language; got != "pl" {
		t.Errorf("persisted language = %q, want pl", got)
	}
}

func TestAConfigOutsideTheSchemaStopsStartup(t *testing.T) {
	for name, content := range map[string]string{
		"dropped theme": `{"preferences":{"theme":"darkHighContrast"}}`,
		"unknown key":   `{"engine":{"lock_hold_ms":1250}}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CONFIG_PATH", path)
			if _, err := Initialize(); err == nil {
				t.Fatalf("Initialize accepted %s", content)
			}
		})
	}
}
