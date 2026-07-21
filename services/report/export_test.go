package report

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appconfig "smegg.me/smeggtuner/common/config"
	sessionsvc "smegg.me/smeggtuner/services/session"
)

func TestExportHTML(t *testing.T) {
	svc, sessions, _ := services(t)
	pass(t, sessions)

	out := export(t, svc, OptionsDTO{Format: FormatHTML})
	if !strings.HasSuffix(out.Path, ".html") {
		t.Fatalf("wrote %q, want an .html file", out.Path)
	}
	if !out.Opened {
		t.Error("the HTML report was not handed to the browser")
	}
	if base := filepath.Base(out.Path); !strings.HasPrefix(base, "morino-") {
		t.Errorf("default name %q does not name the instrument", base)
	}

	raw, err := os.ReadFile(out.Path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)

	// The merged pair's lobes are not reeds and must not reach the file.
	for _, lobe := range []string{"-33.7", "39.1"} {
		if strings.Contains(body, lobe) {
			t.Fatalf("the merged pair's lobe %s reached the report", lobe)
		}
	}
	if !strings.Contains(body, "Reeds not separated") {
		t.Error("the merged note does not say why it has no reeds")
	}
	if !strings.Contains(body, "440.0 Hz") {
		t.Error("the report does not quote the pass's reference")
	}
	// The letterhead is off unless asked for.
	if strings.Contains(body, `<div class="letterhead">`) {
		t.Error("the letterhead printed by default")
	}
}

func TestExportCSV(t *testing.T) {
	svc, sessions, _ := services(t)
	pass(t, sessions)

	out := export(t, svc, OptionsDTO{Format: FormatCSV})
	if !strings.HasSuffix(out.Path, ".csv") {
		t.Fatalf("wrote %q, want a .csv file", out.Path)
	}
	if out.Opened {
		t.Error("a CSV was handed to the browser; it is for a spreadsheet")
	}

	raw, err := os.ReadFile(out.Path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, "reference_a4_hz,440.00") {
		t.Error("the CSV does not carry the reference the numbers were measured from")
	}
	for _, lobe := range []string{"-33.7", "39.1"} {
		if strings.Contains(body, lobe) {
			t.Fatalf("the merged pair's lobe %s reached the spreadsheet", lobe)
		}
	}
}

// TestLetterhead: the letterhead is a config setting, not an export option.
func TestLetterhead(t *testing.T) {
	svc, sessions, _ := services(t)
	pass(t, sessions)

	appconfig.Global.Report = appconfig.Report{
		CompanyName:    "Smegg Accordion Service",
		CompanyAddress: "Warszawa",
		CompanyWebsite: "smegg.me",
	}
	t.Cleanup(func() { appconfig.Global.Report = appconfig.Report{} })

	opts := OptionsDTO{Format: FormatHTML, Date: "2026-07-12"}
	out := export(t, svc, opts)
	raw, err := os.ReadFile(out.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "Smegg Accordion Service") {
		t.Error("the letterhead was asked for and did not print")
	}

	opts.Format = FormatCSV
	out = export(t, svc, opts)
	raw, err = os.ReadFile(out.Path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "Smegg Accordion Service") {
		t.Error("a letterhead ended up in a CSV")
	}
}

func TestExportRefusals(t *testing.T) {
	svc, sessions, _ := services(t)

	if _, err := svc.Export(OptionsDTO{Format: FormatHTML}); !errors.Is(err, sessionsvc.ErrNoSession) {
		t.Errorf("Export with no session = %v, want ErrNoSession", err)
	}

	pass(t, sessions)

	if _, err := svc.Export(OptionsDTO{Format: "docx"}); !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("Export(docx) = %v, want ErrInvalidFormat", err)
	}
	if _, err := svc.Export(OptionsDTO{Format: ""}); !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("Export(no format) = %v, want ErrInvalidFormat", err)
	}
	if _, err := svc.Export(OptionsDTO{Format: FormatHTML, Date: "12/07/2026"}); !errors.Is(err, ErrInvalidDate) {
		t.Errorf("Export(bad date) = %v, want ErrInvalidDate", err)
	}

	logo := filepath.Join(t.TempDir(), "logo.png")
	if err := os.WriteFile(logo, []byte("not an image"), 0o600); err != nil {
		t.Fatal(err)
	}
	appconfig.Global.Report = appconfig.Report{CompanyName: "Smegg", LogoPath: logo}
	t.Cleanup(func() { appconfig.Global.Report = appconfig.Report{} })

	out, err := svc.Export(OptionsDTO{Format: FormatHTML})
	if err != nil {
		t.Errorf("Export with an unreadable logo = %v, want the card anyway", err)
	} else if raw, readErr := os.ReadFile(out.Path); readErr == nil && !strings.Contains(string(raw), "Smegg") {
		t.Error("the card went out without the letterhead it could still print")
	}

	if err := sessions.ClearTakes(); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Export(OptionsDTO{Format: FormatHTML}); !errors.Is(err, sessionsvc.ErrNoReadings) {
		t.Errorf("Export(empty pass) = %v, want ErrNoReadings", err)
	}
}
