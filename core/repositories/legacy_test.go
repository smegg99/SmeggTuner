package repositories_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/datastore/datastoretest"
	"smegg.me/smeggtuner/core/repositories"
	"smegg.me/smeggtuner/core/session"
)

// legacyDir builds a pre-database data directory, including one junk file that must be skipped rather than fatal.
func legacyDir(t *testing.T) (dir string, sessionID, instrumentID string) {
	t.Helper()
	dir = t.TempDir()

	s := session.New("Jan K.", session.Instrument{ReedCount: 3, Banks: []session.Bank{session.BankM1, session.BankM2, session.BankM3}}, 442)
	then := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
	s.Created, s.Updated = then, then
	writeLegacyJSON(t, filepath.Join(dir, "sessions", s.ID+session.LegacyFileExt), map[string]any{
		"v": session.Version, "id": s.ID, "name": s.Name, "instrument": s.Instrument,
		"a4": s.A4, "created": s.Created, "updated": s.Updated,
	})
	writeLegacyJSON(t, filepath.Join(dir, "sessions", "broken"+session.LegacyFileExt), map[string]any{"v": 99})

	tpl := session.Template{ID: session.NewID(), Name: "Morino", Instrument: session.Instrument{ReedCount: 3}}
	writeLegacyJSON(t, filepath.Join(dir, "instruments", tpl.ID+".json"), map[string]any{
		"v": session.TemplateVersion, "id": tpl.ID, "name": tpl.Name, "instrument": tpl.Instrument,
	})
	if err := os.WriteFile(filepath.Join(dir, "instruments", tpl.ID+".jpg"), photo(t, 300, 200), 0o644); err != nil {
		t.Fatal(err)
	}

	return dir, s.ID, tpl.ID
}

func writeLegacyJSON(t *testing.T, path string, v any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLegacyDataComesForwardOnce(t *testing.T) {
	datastoretest.Init(t)
	dir, sessionID, instrumentID := legacyDir(t)

	stats, err := repositories.ImportLegacy(dir)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Sessions != 1 || stats.Instruments != 1 || stats.Skipped != 1 {
		t.Fatalf("stats = %+v, want 1 session, 1 instrument, 1 skipped", stats)
	}

	got, err := repositories.GetSessionRepository().Get(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Updated.Year() != 2024 {
		t.Fatalf("the import touched the history: updated %v", got.Updated)
	}

	tpl, err := repositories.GetInstrumentRepository().Get(instrumentID)
	if err != nil {
		t.Fatal(err)
	}
	if !tpl.HasImage {
		t.Fatal("the photograph did not come forward")
	}
	if jpg, _, err := repositories.GetInstrumentRepository().Image(instrumentID); err != nil || len(jpg) == 0 {
		t.Fatalf("the photograph cannot be read back: %v", err)
	}

	// The legacy files stay in place.
	if _, err := os.Stat(filepath.Join(dir, "sessions", sessionID+session.LegacyFileExt)); err != nil {
		t.Fatalf("the import moved the legacy files: %v", err)
	}
}

// A second run into a non-empty datastore imports nothing; re-import would resurrect deleted rows.
func TestLegacyImportRunsOnlyIntoAnEmptyDatastore(t *testing.T) {
	datastoretest.Init(t)
	dir, sessionID, _ := legacyDir(t)

	if _, err := repositories.ImportLegacy(dir); err != nil {
		t.Fatal(err)
	}
	if err := repositories.GetSessionRepository().Delete(sessionID); err != nil {
		t.Fatal(err)
	}

	stats, err := repositories.ImportLegacy(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !stats.Empty() {
		t.Fatalf("a second import brought data back: %+v", stats)
	}
	if _, err := repositories.GetSessionRepository().Get(sessionID); err == nil {
		t.Fatal("a deleted session was resurrected by the second import")
	}
}
