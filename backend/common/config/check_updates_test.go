// backend/common/config/check_updates_test.go
//go:build !js

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckUpdatesDefaultsAndPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("CONFIG_PATH", path)
	for _, contents := range []string{"{}", `{"preferences":{"check_updates":false}}`} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Initialize(); err != nil {
			t.Fatal(err)
		}
		want := contents == "{}"
		if got := Get().Preferences.CheckUpdates; got != want {
			t.Fatalf("CheckUpdates = %v, want %v", got, want)
		}
		prefs := Get().Preferences
		prefs.CheckUpdates = !want
		if err := SetPreferences(prefs); err != nil {
			t.Fatal(err)
		}
		if _, err := Initialize(); err != nil {
			t.Fatal(err)
		}
		if Get().Preferences.CheckUpdates != !want {
			t.Fatal("update preference was not persisted")
		}
	}
}
