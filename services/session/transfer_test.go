package session

import (
	"path/filepath"
	"testing"

	coresession "smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/tuning"
)

// An exported-then-imported session is a second session, so an import cannot overwrite the original.
func TestAnImportedSessionDoesNotOverwriteTheOneItCameFrom(t *testing.T) {
	s := service(t)
	orig := create(t, s, "Hohner Morino", 3, 442)

	if err := s.UpsertTake(take(69, 442, -3, 0, 3)); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "out.stsf")
	if err := s.ExportSession(orig.ID, path); err != nil {
		t.Fatalf("export: %v", err)
	}

	got, err := s.ImportSession(path)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if got.ID == orig.ID {
		t.Fatal("the import landed on top of the session it came from")
	}
	if got.Name != orig.Name || got.A4 != 442 || got.Readings != 1 {
		t.Fatalf("the session did not survive the trip: %+v", got)
	}

	list, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("the bench holds %d sessions, want both", len(list))
	}
}

// Exporting the active session flushes first, so it cannot race the saver and drop the newest readings.
func TestExportingTheOpenSessionWaitsForItsOwnTakes(t *testing.T) {
	s := service(t)
	dto := create(t, s, "Bench", 3, 442)

	for _, n := range []tuning.Note{60, 62, 64} {
		if err := s.UpsertTake(take(n, 442, -3, 0, 3)); err != nil {
			t.Fatal(err)
		}
	}

	path := filepath.Join(t.TempDir(), "out.stsf")
	if err := s.ExportSession(dto.ID, path); err != nil {
		t.Fatalf("export: %v", err)
	}

	out, err := coresession.ReadSessionFile(path)
	if err != nil {
		t.Fatalf("what was exported will not read back: %v", err)
	}
	if len(out.Takes) != 3 {
		t.Fatalf("the export is missing readings: %+v", out.Takes)
	}
}
