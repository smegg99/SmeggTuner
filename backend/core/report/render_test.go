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

// A silent slot prints a dash, never a zero, and the sheet still prints every reading around it.
func TestHTMLPrintsASilentSlotAsAbsent(t *testing.T) {
	r := sheet(t, musette(t))
	out := text(t, r)

	gap := row(t, r, 64)
	for _, c := range gap.Reeds {
		if c.Present && !strings.Contains(out, fmt.Sprintf("%+.1f", c.Curr)) {
			t.Errorf("reed %d of the gapped row (%+.1f) is not on the sheet", c.Reed, c.Curr)
		}
	}
	if !strings.Contains(out, `class="sep absent">-<`) {
		t.Error("the silent slot does not print as absent")
	}
	beat := row(t, r, 60).Beats[0]
	if !strings.Contains(out, fmt.Sprintf("%+.2f", beat.CurrHz)) {
		t.Errorf("the beat rate (%+.2f Hz) is not on the sheet", beat.CurrHz)
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

	// Case-insensitive: the sentence may lead the phrase or follow one, and which it does is
	// copy, not behaviour.
	if !strings.Contains(strings.ToLower(out), "no goal curve") {
		t.Error("the sheet does not say that there is no goal curve")
	}
	for _, col := range []string{">Curr<", ">Err<"} {
		if !strings.Contains(out, col) {
			t.Errorf("the %s column was dropped because there is no curve", col)
		}
	}
}
