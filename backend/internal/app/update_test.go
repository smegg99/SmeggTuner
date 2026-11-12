// backend/internal/app/update_test.go
package app

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/smegg99/s99wails/v2/shell/update"
	"github.com/smegg99/s99wails/v2/shell/windowstate"

	"smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/core/datastore"
	"smegg.me/smeggtuner/backend/core/datastore/datastoretest"
)

func TestUpdateOptions(t *testing.T) {
	opts := updateOptions()
	if opts.Owner != "smegg99" || opts.Repo != "SmeggTuner" || opts.Version != version {
		t.Fatalf("release source = %s/%s version %q", opts.Owner, opts.Repo, opts.Version)
	}
	if opts.DownloadURL != "https://smeggtuner.com/#downloads" {
		t.Fatalf("DownloadURL = %q", opts.DownloadURL)
	}
	if !strings.Contains(opts.Changelog, "<!-- lang:en -->") || !strings.Contains(opts.Changelog, "<!-- lang:pl -->") {
		t.Fatal("bundled changelog is missing language sections")
	}
	if opts.Store == nil || opts.Enabled == nil || opts.Existing == nil || opts.OnError == nil {
		t.Fatal("update callbacks or store are missing")
	}
}

func TestUpdateOptionsReadCurrentPreference(t *testing.T) {
	t.Setenv("CONFIG_PATH", filepath.Join(t.TempDir(), "config.json"))
	if _, err := config.Initialize(); err != nil {
		t.Fatal(err)
	}
	opts := updateOptions()
	if !opts.Enabled() {
		t.Fatal("default update preference is disabled")
	}
	prefs := config.Get().Preferences
	prefs.CheckUpdates = false
	if err := config.SetPreferences(prefs); err != nil {
		t.Fatal(err)
	}
	if opts.Enabled() {
		t.Fatal("update callback kept the old preference")
	}
}

func TestUpdateStateRecordsFreshInstallBeforeWindowHistory(t *testing.T) {
	datastoretest.Init(t)
	svc := update.New(updateOptions())
	state := svc.State("en")
	if state.Notes != nil {
		t.Fatalf("fresh state = %+v", state)
	}
	if err := (datastore.WindowStore{}).Save(windowstate.State{Width: 1100, Height: 720}); err != nil {
		t.Fatal(err)
	}
	state = svc.State("pl")
	if state.Notes != nil {
		t.Fatalf("first window created upgrade notes: %+v", state)
	}
	memory, ok, err := (datastore.UpdateStore{}).Load()
	if err != nil || !ok || memory.Notes != version {
		t.Fatalf("fresh version not recorded: %+v, %v, %v", memory, ok, err)
	}
}

func TestUpdateStateShowsExistingInstallNotesOffline(t *testing.T) {
	datastoretest.Init(t)
	if err := (datastore.WindowStore{}).Save(windowstate.State{Width: 1100, Height: 720}); err != nil {
		t.Fatal(err)
	}
	opts := updateOptions()
	opts.Enabled = func() bool { return false }
	svc := update.New(opts)
	state := svc.State("pl")
	if state.Notes == nil || state.Notes.Lang != "pl" || state.Notes.Version != version {
		t.Fatalf("existing state = %+v", state)
	}
	if err := svc.NotesShown(); err != nil {
		t.Fatal(err)
	}
	state = update.New(opts).State("en")
	if state.Notes != nil {
		t.Fatalf("notes repeated after restart: %+v", state)
	}
}
