// backend/services/config/config_test.go
package config

import (
	"path/filepath"
	"testing"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/i18n"
)

func TestPreferencePatchPreservesCalibrationAndTunerSettings(t *testing.T) {
	t.Setenv("CONFIG_PATH", filepath.Join(t.TempDir(), "config.json"))
	if _, err := appconfig.Initialize(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { i18n.SetLocale("en") })
	service := &Service{}
	old := service.GetConfig()
	device := appconfig.NewDevice()
	device.Calibration, device.CalibrationPercent = "calibrated", 0.125
	if err := appconfig.SetDevice("mic", device); err != nil {
		t.Fatal(err)
	}
	if err := appconfig.Update(func(c *appconfig.Config) { c.Tuner.A4 = 441 }); err != nil {
		t.Fatal(err)
	}
	if old.Tuner.A4 != 440 || len(old.Audio.Devices) != 0 {
		t.Fatal("initial frontend snapshot changed")
	}
	result, err := service.PatchConfig(map[string]any{"preferences": map[string]any{"theme": "dark", "language": "pl"}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Tuner.A4 != 441 || result.Audio.Device("mic") != device || result.Preferences.Theme != "dark" || i18n.Current() != "pl" {
		t.Fatalf("preference edit lost independent state: %+v", result)
	}
	if _, err := appconfig.Initialize(); err != nil {
		t.Fatal(err)
	}
	if current := service.GetConfig(); current.Tuner.A4 != 441 || current.Audio.Device("mic") != device || current.Preferences.Theme != "dark" {
		t.Fatalf("saved state differs after restart: %+v", current)
	}
	if _, err := service.PatchConfig(map[string]any{"tuner": map[string]any{"a4": -1}}); err == nil {
		t.Fatal("invalid configuration patch accepted")
	}
	if service.GetConfig().Tuner.A4 != 441 {
		t.Fatal("invalid patch replaced the accepted tuner setting")
	}
}
