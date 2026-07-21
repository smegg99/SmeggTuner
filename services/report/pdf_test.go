package report

import (
	"bytes"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestExportPDF starts a real headless browser (skipped without one) to catch failures nothing smaller can.
func TestExportPDF(t *testing.T) {
	if !haveBrowser() {
		t.Skip("no headless browser on this machine")
	}
	svc, sessions, _ := services(t)
	pass(t, sessions)

	out := export(t, svc, OptionsDTO{Format: FormatPDF})
	if !strings.HasSuffix(out.Path, ".pdf") {
		t.Fatalf("wrote %q, want a .pdf file", out.Path)
	}
	if out.Opened {
		t.Error("the PDF was handed to a browser; it is a saved file")
	}

	raw, err := os.ReadFile(out.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(raw, []byte("%PDF-")) {
		t.Fatalf("the file is not a PDF: starts %.20q", raw)
	}

	// One page, A4: a regression once doubled margins onto a second sheet.
	if n := bytes.Count(raw, []byte("/Type /Page\n")); n != 1 {
		t.Errorf("the card printed on %d pages, want 1", n)
	}
	if w, h := pageSizeMM(t, raw); !closeTo(w, 210, 1) || !closeTo(h, 297, 1) {
		t.Errorf("the page is %.1f x %.1f mm, want A4 (210 x 297)", w, h)
	}
}

// A five-reed card is landscape; printed portrait it loses the right-hand reeds off the paper.
func TestExportPDFTurnsThePage(t *testing.T) {
	if !haveBrowser() {
		t.Skip("no headless browser on this machine")
	}
	svc, sessions, _ := services(t)
	passWithReeds(t, sessions, 5)

	out := export(t, svc, OptionsDTO{Format: FormatPDF})
	raw, err := os.ReadFile(out.Path)
	if err != nil {
		t.Fatal(err)
	}
	w, h := pageSizeMM(t, raw)
	if w < h {
		t.Errorf("a five-reed card printed %.1f x %.1f mm, which is upright", w, h)
	}
}

func haveBrowser() bool {
	for _, name := range []string{"google-chrome-stable", "google-chrome", "chromium", "chromium-browser"} {
		if _, err := exec.LookPath(name); err == nil {
			return true
		}
	}
	return false
}

// pageSizeMM reads the first MediaBox (in points, 1/72") and returns millimetres.
func pageSizeMM(t *testing.T, raw []byte) (float64, float64) {
	t.Helper()
	m := regexp.MustCompile(`/MediaBox \[([\d.]+) ([\d.]+) ([\d.]+) ([\d.]+)\]`).FindSubmatch(raw)
	if m == nil {
		t.Fatal("the PDF has no MediaBox, so it declares no page size")
	}
	num := func(b []byte) float64 {
		v, err := strconv.ParseFloat(string(b), 64)
		if err != nil {
			t.Fatalf("MediaBox holds %q, which is not a number", b)
		}
		return v
	}
	const mmPerPoint = 25.4 / 72
	return (num(m[3]) - num(m[1])) * mmPerPoint, (num(m[4]) - num(m[2])) * mmPerPoint
}

func closeTo(got, want, slack float64) bool { return math.Abs(got-want) <= slack }
