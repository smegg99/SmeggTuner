package report

import (
	"regexp"
	"strings"
	"testing"

	"smegg.me/smeggtuner/common/i18n"
)

// The locale is process-global, so every test here restores it.

// rawKey catches a message ID that reached the page instead of its text.
var rawKey = regexp.MustCompile(`report\.[a-z]+\.[a-zA-Z]+`)

func TestSheetIsEnglishByDefault(t *testing.T) {
	defer i18n.SetLocale("en")
	i18n.SetLocale("en")

	out := text(t, sheet(t, musette(t)))
	for _, want := range []string{"Tuning table", "Tolerance", "Note", "Curr"} {
		if !strings.Contains(out, want) {
			t.Errorf("the sheet does not say %q", want)
		}
	}
	if strings.Contains(out, "435,0") {
		t.Error("an English sheet is using a decimal comma")
	}
}

func TestSheetFollowsTheLanguage(t *testing.T) {
	defer i18n.SetLocale("en")

	i18n.SetLocale("pl")
	out := text(t, sheet(t, musette(t)))

	// The pass reference is 435.0 Hz; in Polish it is 435,0.
	if !strings.Contains(out, "435,0 Hz") {
		t.Error("a Polish sheet still writes its numbers with a decimal point")
	}
	if strings.Contains(out, "435.0 Hz") {
		t.Error("a Polish sheet is writing 435.0 somewhere")
	}
}

// German is not shipped, so it exercises the English-per-key fallback.
func TestUntranslatedLanguageFallsBackToEnglish(t *testing.T) {
	defer i18n.SetLocale("en")

	i18n.SetLocale("de")
	out := text(t, sheet(t, musette(t)))

	if got := rawKey.FindString(out); got != "" {
		t.Errorf("the sheet printed the message id %q instead of a string", got)
	}
	if !strings.Contains(out, "Tuning table") {
		t.Error("an untranslated language did not fall back to English")
	}
}

// The Polish sheet must be fully Polish; each leak below comes from a different source.
func TestPolishSheetIsPolish(t *testing.T) {
	defer i18n.SetLocale("en")

	i18n.SetLocale("pl")
	out := text(t, sheet(t, musette(t)))

	for _, want := range []string{"Tabela strojenia", "Tolerancja", "Dźwięk", "Jest"} {
		if !strings.Contains(out, want) {
			t.Errorf("the Polish sheet does not say %q", want)
		}
	}
	for _, leak := range []string{
		"Tuning report", "Tuning table", "Reed 1", "Beat 1-2", "cents",
		"pass \"", "reference A4", "Generated",
	} {
		if strings.Contains(out, leak) {
			t.Errorf("the Polish sheet prints the English %q", leak)
		}
	}
	if got := rawKey.FindString(out); got != "" {
		t.Errorf("the sheet printed the message id %q instead of a string", got)
	}
}

// The lang attribute drives screen readers and hyphenation, whatever the words.
func TestSheetDeclaresItsLanguage(t *testing.T) {
	defer i18n.SetLocale("en")

	i18n.SetLocale("pl")
	if got := render(t, sheet(t, musette(t))); !strings.Contains(got, `<html lang="pl">`) {
		t.Error("the Polish sheet does not declare lang=pl")
	}

	i18n.SetLocale("en")
	if got := render(t, sheet(t, musette(t))); !strings.Contains(got, `<html lang="en">`) {
		t.Error("the English sheet does not declare lang=en")
	}
}

// Polish plurals: 12 vs 22 both end in 2, but 22 is "few" and 12 is "many".
func TestPolishPluralBands(t *testing.T) {
	defer i18n.SetLocale("en")
	i18n.SetLocale("pl")

	for _, c := range []struct {
		count int
		want  string
	}{
		{1, "1 dźwięk"},
		{2, "2 dźwięki"},
		{4, "4 dźwięki"},
		{5, "5 dźwięków"},
		{12, "12 dźwięków"},
		{22, "22 dźwięki"},
		{25, "25 dźwięków"},
	} {
		if got := i18n.Tn("report.summary.notes", c.count, nil); got != c.want {
			t.Errorf("Tn(notes, %d) = %q, want %q", c.count, got, c.want)
		}
	}
}

// Asserts the counts go through Tn, not hard-coded English fragments.
func TestCountsAreRealPlurals(t *testing.T) {
	defer i18n.SetLocale("en")
	i18n.SetLocale("en")

	one := i18n.Tn("report.summary.notes", 1, nil)
	many := i18n.Tn("report.summary.notes", 4, nil)
	if one != "1 note" || many != "4 notes" {
		t.Errorf("English counts read %q and %q", one, many)
	}

	i18n.SetLocale("pl")
	if got := i18n.Tn("report.summary.notes", 22, nil); !strings.HasPrefix(got, "22 ") {
		t.Errorf("Polish count lost its number: %q", got)
	}
}

func TestCurrentReportsTheLanguage(t *testing.T) {
	defer i18n.SetLocale("en")

	i18n.SetLocale("pl-PL")
	if got := i18n.Current(); got != "pl-PL" {
		t.Errorf("Current() = %q after SetLocale(pl-PL)", got)
	}
	// A region suffix still triggers the decimal comma.
	if got := decimal("130.8"); got != "130,8" {
		t.Errorf("decimal(130.8) under pl-PL = %q", got)
	}

	i18n.SetLocale("")
	if got := i18n.Current(); got != "en" {
		t.Errorf("Current() = %q after SetLocale(\"\"), want the default", got)
	}
}
