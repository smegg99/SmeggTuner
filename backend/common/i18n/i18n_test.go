package i18n

import "testing"

// Locale files use flat dotted message IDs; go-i18n reserves keys inside nested maps.
func TestDottedKeysResolve(t *testing.T) {
	defer SetLocale("en")

	SetLocale("en")
	if got := T("tray.show"); got != "Show" {
		t.Fatalf("T(tray.show) = %q, want Show", got)
	}
	if got := T("app.name"); got != "SmeggTuner" {
		t.Fatalf("T(app.name) = %q", got)
	}
}

func TestLocaleSwitchAndMatching(t *testing.T) {
	defer SetLocale("en")

	// Region suffixes must match to the base language.
	SetLocale("pl-PL")
	if got := T("tray.show"); got == "Show" || got == "tray.show" {
		t.Fatalf("T(tray.show) after SetLocale(pl-PL) = %q, want Polish translation", got)
	}

	// Unknown locales fall back to the default language.
	SetLocale("xx")
	if got := T("tray.show"); got != "Show" {
		t.Fatalf("T(tray.show) after SetLocale(xx) = %q, want Show", got)
	}
}

func TestMissingKeyReturnsKey(t *testing.T) {
	if got := T("no.such.key"); got != "no.such.key" {
		t.Fatalf("T(no.such.key) = %q, want the key itself", got)
	}
}

func TestOnChangeFires(t *testing.T) {
	defer SetLocale("en")

	fired := 0
	OnChange(func() { fired++ })
	SetLocale("pl")
	if fired != 1 {
		t.Fatalf("OnChange fired %d times, want 1", fired)
	}
}
