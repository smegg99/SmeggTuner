// backend/common/i18n/i18n.go
// Package i18n loads S99Locale-generated native messages through S99Wails.
package i18n

import (
	"embed"
	"fmt"
	"io/fs"
	"sync/atomic"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/smegg99/s99logger"
	"github.com/smegg99/s99wails/v2/locale"
)

//go:embed locales/*.json
var localeFS embed.FS

//go:generate go tool loc build -c ../../../i18n/config.cue

var locales = mustLoad()

// current mirrors the last SetLocale code, which the locale package does not itself retain.
var current atomic.Pointer[string]

// Localizer returns the active S99Wails localizer for generated accessors.
func Localizer() *goi18n.Localizer { return locales.Current() }

// Translator localizes log messages without changing stable event identifiers.
func Translator() s99logger.Translator { return locales.Translator() }

// mustLoad parses the embedded locale files; a failure is a build defect, so it panics.
func mustLoad() *locale.Locales {
	sub, err := fs.Sub(localeFS, "locales")
	if err != nil {
		panic(fmt.Sprintf("i18n: sub locales fs: %v", err))
	}
	l, err := locale.New(sub, DefaultLocale)
	if err != nil {
		panic(fmt.Sprintf("i18n: load locales: %v", err))
	}
	return l
}

// SetLocale switches the active locale; go-i18n matching handles region suffixes (pl-PL to pl) and falls back to English.
func SetLocale(code string) {
	if code == "" {
		code = DefaultLocale
	}
	current.Store(&code)
	locales.SetLocale(code)
}

// OnChange registers a callback invoked after every SetLocale.
func OnChange(cb func()) {
	locales.OnChange(cb)
}

// T returns the translated string for key, or the key itself if missing.
func T(key string) string {
	return locales.T(key)
}

// Tf returns the translated string for key with template data.
func Tf(key string, data map[string]any) string {
	return locales.Tf(key, data)
}

// Tn returns the plural form of key for count (read in the message as {{.PluralCount}}), using CLDR plural rules.
func Tn(key string, count int, data map[string]any) string {
	return locales.Tn(key, count, data)
}

// Current returns the locale code last passed to SetLocale: the active language, independent of per-key English fallback.
func Current() string {
	if code := current.Load(); code != nil {
		return *code
	}
	return DefaultLocale
}
