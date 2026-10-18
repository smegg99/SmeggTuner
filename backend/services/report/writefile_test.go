package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportCancelled(t *testing.T) {
	svc, sessions, dir := services(t)
	pass(t, sessions)

	saveDialog = func(string, string) (string, error) { return "", nil }

	out := export(t, svc, OptionsDTO{Format: FormatHTML})
	if out.Path != "" || out.Opened {
		t.Fatalf("a cancelled export produced %+v", out)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("a cancelled export wrote %d files", len(entries))
	}
}

func TestExportAddsTheExtension(t *testing.T) {
	svc, sessions, dir := services(t)
	pass(t, sessions)

	saveDialog = func(string, string) (string, error) { return filepath.Join(dir, "bench"), nil }

	out := export(t, svc, OptionsDTO{Format: FormatHTML})
	if filepath.Base(out.Path) != "bench.html" {
		t.Errorf("wrote %q, want bench.html", filepath.Base(out.Path))
	}
	if _, err := os.Stat(out.Path); err != nil {
		t.Errorf("the report is not where it says it is: %v", err)
	}
}

func TestExportDoesNotDestroyAnOlderReport(t *testing.T) {
	svc, sessions, dir := services(t)
	pass(t, sessions)

	path := filepath.Join(dir, "report.html")
	if err := os.WriteFile(path, []byte("yesterday's report"), 0o600); err != nil {
		t.Fatal(err)
	}
	saveDialog = func(string, string) (string, error) { return path, nil }

	out := export(t, svc, OptionsDTO{Format: FormatHTML})
	raw, err := os.ReadFile(out.Path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "yesterday") {
		t.Fatal("the new report did not replace the old one")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("the export left %d files in the directory, want 1", len(entries))
	}
}

func TestDefaultNameIsReadableAYearLater(t *testing.T) {
	if got := slug("Hohner Morino - Jan K."); got != "hohner-morino-jan-k" {
		t.Errorf("slug = %q", got)
	}
	if got := withExtension("/tmp/x.csv", FormatCSV); got != "/tmp/x.csv" {
		t.Errorf("withExtension doubled an extension: %q", got)
	}
}
