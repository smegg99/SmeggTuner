package report

import (
	"bytes"
	"fmt"
	stdhtml "html"
	"strings"
	"testing"
)

// render is the sheet as written to disk; html/template escapes a leading plus, so number assertions use text().
func render(t *testing.T, r *Report) string {
	t.Helper()
	var buf bytes.Buffer
	if err := HTML(&buf, r); err != nil {
		t.Fatalf("HTML: %v", err)
	}
	return buf.String()
}

// text is the sheet unescaped, as a browser would draw it.
func text(t *testing.T, r *Report) string {
	t.Helper()
	return stdhtml.UnescapeString(render(t, r))
}

func lobes() []string {
	return []string{
		fmt.Sprintf("%+.1f", mergedLobeLow),
		fmt.Sprintf("%+.1f", mergedLobeHigh),
		fmt.Sprintf("%.2f", mergedLobeLow),
		fmt.Sprintf("%.2f", mergedLobeHigh),
	}
}

// A merged pair's lobes must never print, and the row must say why its reeds are missing.
func TestHTMLNeverPrintsMergedReeds(t *testing.T) {
	out := text(t, sheet(t, musette(t)))

	for _, lobe := range lobes() {
		if strings.Contains(out, lobe) {
			t.Fatalf("the merged pair's lobe %s reached the printed sheet: it is not a reed, and a "+
				"technician reading it files the wrong one", lobe)
		}
	}
	if !strings.Contains(out, "Reeds not separated") {
		t.Error("the merged row does not say why it has no reeds")
	}
	if !strings.Contains(out, "M - reeds merged") {
		t.Error("the sheet has a merged row and no legend explaining it")
	}

	merged := row(t, sheet(t, musette(t)), 64)
	beat := merged.Beats[0]
	if !beat.Present {
		t.Fatal("fixture: the merged note has no beat")
	}
	if !strings.Contains(out, fmt.Sprintf("%+.1f", beat.Curr)) {
		t.Errorf("the merged note's beat (%+.1f cents) is not on the sheet", beat.Curr)
	}
	if !strings.Contains(out, fmt.Sprintf("%+.2f", beat.CurrHz)) {
		t.Errorf("the merged note's beat rate (%+.2f Hz) is not on the sheet", beat.CurrHz)
	}
}

// Reeds recovered from a measured beat print like any other, and say so.
func TestHTMLPrintsDerivedReeds(t *testing.T) {
	r := sheet(t, musette(t))
	out := text(t, r)

	d := row(t, r, 65)
	for _, c := range d.Reeds {
		if !strings.Contains(out, fmt.Sprintf("%+.1f", c.Curr)) {
			t.Errorf("reed %d of the derived row (%+.1f) is not on the sheet", c.Reed, c.Curr)
		}
	}
	if !strings.Contains(out, "D - reeds from the beat") {
		t.Error("the derived row has no legend")
	}
}

// The pass's own reference, not the session's current one; they differ here by 7 Hz.
func TestHTMLQuotesThePassReference(t *testing.T) {
	out := text(t, sheet(t, musette(t)))

	if !strings.Contains(out, "435.0 Hz") {
		t.Error("the sheet does not quote the session's reference")
	}
	if strings.Contains(out, "442.0 Hz") {
		t.Error("the sheet quotes the session's current reference, which this pass was not measured against")
	}
}

// No goal curve is ordinary: the columns stay the same as with one.
func TestHTMLWithoutACurve(t *testing.T) {
	s := musette(t)
	s.Curve = nil
	out := text(t, sheet(t, s))

	if !strings.Contains(out, "no goal curve") {
		t.Error("the sheet does not say that there is no goal curve")
	}
	for _, col := range []string{">Curr<", ">Err<"} {
		if !strings.Contains(out, col) {
			t.Errorf("the %s column was dropped because there is no curve", col)
		}
	}
	if strings.Contains(out, "Reeds not separated") != true {
		t.Error("the merged row lost its explanation when the curve went away")
	}
}
