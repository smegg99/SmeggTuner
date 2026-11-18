package i18n

import (
	"encoding/json"
	"strings"
	"testing"
)

// Locale files use flat dotted message IDs; go-i18n reserves keys inside nested maps.
func TestDottedKeysResolve(t *testing.T) {
	defer SetLocale("en")

	SetLocale("en")
	if got := T("common.tray.show"); got != "Show" {
		t.Fatalf("T(common.tray.show) = %q, want Show", got)
	}
	if got := T("common.app.name"); got != "SmeggTuner" {
		t.Fatalf("T(common.app.name) = %q", got)
	}
}

func TestLocaleSwitchAndMatching(t *testing.T) {
	defer SetLocale("en")

	// Region suffixes must match to the base language.
	SetLocale("pl-PL")
	if got := T("common.tray.show"); got == "Show" || got == "common.tray.show" {
		t.Fatalf("T(common.tray.show) after SetLocale(pl-PL) = %q, want Polish translation", got)
	}

	// Unknown locales fall back to the default language.
	SetLocale("xx")
	if got := T("common.tray.show"); got != "Show" {
		t.Fatalf("T(common.tray.show) after SetLocale(xx) = %q, want Show", got)
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

// Loading all catalogs into one bundle must preserve regional and script variants.
func TestNativeMessagesUseSelectedCatalog(t *testing.T) {
	defer SetLocale(DefaultLocale)
	for _, code := range Locales {
		t.Run(code, func(t *testing.T) {
			data, err := LocaleFiles.ReadFile("locales/" + code + ".json")
			if err != nil {
				t.Fatal(err)
			}
			var messages map[string]any
			if err := json.Unmarshal(data, &messages); err != nil {
				t.Fatal(err)
			}
			SetLocale(code)
			for key, entry := range messages {
				want, plain := entry.(string)
				if plain && !strings.Contains(want, "{{") {
					if got := T(key); got != want {
						t.Errorf("%s: got %q, want %q", key, got, want)
					}
				}
			}
		})
	}
}
