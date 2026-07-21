package report

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"smegg.me/smeggtuner/core/session"
)

// Printed sheet: nothing may be fetched, nothing may run off the paper.
func TestHTMLIsSelfContainedAndPrintable(t *testing.T) {
	out := text(t, sheet(t, musette(t)))

	for _, forbidden := range []string{"<script", "<link", "@import", "url(http", "src=\"http", "//cdn"} {
		if strings.Contains(out, forbidden) {
			t.Errorf("the sheet reaches outside itself: %q", forbidden)
		}
	}
	// A link is not a fetch: remaining URLs must be an anchor href or the SVG namespace.
	for pos := 0; ; {
		rel := strings.Index(out[pos:], "http")
		if rel < 0 {
			break
		}
		at := pos + rel
		if strings.HasPrefix(out[at:], "http://www.w3.org/2000/svg") || strings.HasSuffix(out[:at], `<a href="`) {
			pos = at + len("http")
			continue
		}
		t.Errorf("the sheet holds a URL that is neither a link nor the SVG namespace: %.40q", out[at:])
		break
	}

	for _, want := range []string{
		"@media print",
		"@page",
		"size: A4 portrait", // three reeds: 16 columns, they fit upright
		"table-layout: fixed",
		"display: table-header-group", // the header repeats on every printed page
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the print stylesheet is missing %q", want)
		}
	}
}

// Grayscale only: every colour is a grey, and a hex is grey when its three channels are equal.
func TestTheSheetIsGrayscale(t *testing.T) {
	out := text(t, sheet(t, musette(t)))

	for _, syntax := range []string{"rgb(", "hsl(", "color-mix("} {
		if strings.Contains(out, syntax) {
			t.Errorf("the sheet uses %q: a colour the printer may not have", syntax)
		}
	}

	for _, m := range regexp.MustCompile(`#[0-9a-fA-F]{3,6}\b`).FindAllString(out, -1) {
		hex := m[1:]
		if len(hex) == 3 {
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		if len(hex) != 6 {
			continue
		}
		if !strings.EqualFold(hex[0:2], hex[2:4]) || !strings.EqualFold(hex[2:4], hex[4:6]) {
			t.Errorf("the sheet holds a colour that is not a grey: %s", m)
		}
	}
}

// The page turns before a column is dropped; past five reeds the table is grouped.
func TestHTMLTurnsThePageForWiderInstruments(t *testing.T) {
	five := render(t, sheet(t, bench(t, 5)))
	if !strings.Contains(five, "size: A4 landscape") {
		t.Error("a five-reed instrument (28 columns) does not turn the page")
	}
	if !strings.Contains(five, `<body class="wide landscape">`) {
		t.Error("the body does not carry the landscape layout")
	}

	eight := render(t, sheet(t, bench(t, 8)))
	if !strings.Contains(eight, `<body class="grouped">`) {
		t.Error("an eight-reed instrument does not fall back to grouped tables")
	}
	for reed := 1; reed <= 8; reed++ {
		if !strings.Contains(eight, fmt.Sprintf("<caption>Reed %d</caption>", reed)) {
			t.Errorf("the grouped sheet has no table for reed %d", reed)
		}
	}
	if !strings.Contains(eight, "<caption>Beat 7-8</caption>") {
		t.Error("the grouped sheet drops the last beat pair")
	}
}

// bench is a one-note session at the given reed count, for the column arithmetic.
func bench(t *testing.T, reedCount int) *session.Session {
	t.Helper()
	s := session.New("Bench", session.Instrument{ReedCount: reedCount}, passA4)
	cents := make([]float64, reedCount)
	for i := range cents {
		cents[i] = float64(i) * 4
	}
	s.UpsertTake(reeds(60, cents...))
	return s
}
