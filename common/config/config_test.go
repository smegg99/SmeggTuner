package config

import (
	"os"
	"path/filepath"
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
	if cfg.Tuner.Unit != "cent" {
		t.Fatalf("tuner defaults: %+v", cfg.Tuner)
	}
	if cfg.Audio.ClockPPM != 0 {
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
	if cfg.Report.CompanyName != "" || cfg.Report.CompanyAddress != "" || cfg.Report.CompanyWebsite != "" || cfg.Report.LogoPath != "" {
		t.Fatalf("report defaults: %+v", cfg.Report)
	}
	// Must mirror core/dsp.DefaultEngineConfig; the services boundary maps them onto it.
	if cfg.Engine.FineWindowMs != 3000 || cfg.Engine.LockHoldMs != 1250 || cfg.Engine.LockEpsilonHz != 0.1 {
		t.Fatalf("engine defaults: %+v", cfg.Engine)
	}
	if cfg.Tuner.ToneDurationMs != 10000 || cfg.Tuner.CalibrationCaptureMs != 2000 {
		t.Fatalf("tuner timing defaults: tone=%d calib=%d", cfg.Tuner.ToneDurationMs, cfg.Tuner.CalibrationCaptureMs)
	}
}

// A pre-engine config must still load: CUE fills the missing section from schema defaults.
func TestOlderConfigWithoutEngineSectionLoads(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	old := `{"preferences":{"theme":"dark"},"tuner":{"a4":442}}`
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", path)

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize on a pre-engine config: %v", err)
	}
	cfg := Get()
	if cfg.Tuner.A4 != 442 {
		t.Errorf("the user's a4 was lost: %v", cfg.Tuner.A4)
	}
	if cfg.Engine.FineWindowMs != 3000 || cfg.Engine.LockHoldMs != 1250 {
		t.Errorf("engine section did not fill from defaults: %+v", cfg.Engine)
	}
	if cfg.Tuner.CalibrationCaptureMs != 2000 {
		t.Errorf("new tuner timing did not fill: %d", cfg.Tuner.CalibrationCaptureMs)
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
