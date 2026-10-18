package report

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	sessionsvc "smegg.me/smeggtuner/backend/services/session"
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

	// The reed that did not sound prints absent, never as its -99 sentinel.
	if strings.Contains(body, "-99") {
		t.Fatal("the silent reed's sentinel reached the report as a number")
	}
	if !strings.Contains(body, `class="absent"`) {
		t.Error("the silent reed does not print as absent")
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
	if strings.Contains(body, "-99") {
		t.Fatal("the silent reed's sentinel reached the spreadsheet as a number")
	}
	if !strings.Contains(body, "tolerance_ref,") {
		t.Error("the CSV does not say its deviations are reference units by default")
	}
}

// The sheet prints in the unit the tuner screen is set to; storage stays reference units.
func TestExportFollowsTheDeviationUnit(t *testing.T) {
	svc, sessions, _ := services(t)
	pass(t, sessions)

	appconfig.Global.Tuner.DeviationUnit = "cents"
	t.Cleanup(func() { appconfig.Global.Tuner.DeviationUnit = "" })

	out := export(t, svc, OptionsDTO{Format: FormatCSV})
	raw, err := os.ReadFile(out.Path)
	if err != nil {
		t.Fatal(err)
	}
	if body := string(raw); !strings.Contains(body, "tolerance_cents,") || strings.Contains(body, "tolerance_ref,") {
		t.Fatalf("a cents setting must print a cents sheet:\n%s", body)
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

	// The letterhead is words, and words alone: there is no image for a technician to put on the
	// card, so nothing here can fail to load.
	appconfig.Global.Report = appconfig.Report{CompanyName: "Smegg"}
	t.Cleanup(func() { appconfig.Global.Report = appconfig.Report{} })

	out, err := svc.Export(OptionsDTO{Format: FormatHTML})
	if err != nil {
		t.Errorf("Export with a letterhead = %v, want the card", err)
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
