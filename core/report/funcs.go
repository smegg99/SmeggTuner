package report

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"smegg.me/smeggtuner/common/i18n"
)

// Localization is read at call time so the template parses once; CSV is deliberately not localized (a decimal comma would break its columns).

// decimal swaps the decimal point for a comma in pl locale, rather than reformatting.
func decimal(s string) string {
	if strings.HasPrefix(i18n.Current(), "pl") {
		return strings.ReplaceAll(s, ".", ",")
	}
	return s
}

var funcs = template.FuncMap{
	// cents keeps its sign even for zero.
	"cents": func(v float64) string { return decimal(fmt.Sprintf("%+.1f", v)) },
	"hz":    func(v float64) string { return decimal(fmt.Sprintf("%+.2f", v)) },
	"freq":  func(v float64) string { return decimal(fmt.Sprintf("%.1f", v)) },
	"tol":   func(v float64) string { return decimal(fmt.Sprintf("%.1f", v)) },
	"date":  func(t time.Time) string { return t.Format("2006-01-02") },
	"stamp": func(t time.Time) string { return t.Format("2006-01-02 15:04") },

	// lang sets the document language for hyphenation and screen readers.
	"lang": func() string { return i18n.Current() },

	// T/Tn/Tf translate; a key missing from the active language falls back to English.
	"T": i18n.T,
	"Tn": func(key string, count int, pairs ...any) string {
		return i18n.Tn(key, count, dataOf(pairs))
	},
	"Tf": func(key string, pairs ...any) string {
		return i18n.Tf(key, dataOf(pairs))
	},
	// add and sub nudge already-computed graph labels; the template does no other math.
	"add": func(a, b float64) float64 { return a + b },
	"sub": func(a, b float64) float64 { return a - b },
	// logo passes through only a data:image URI (already checked by LoadLogo); anything else is dropped.
	"logo": func(s string) template.URL {
		if !strings.HasPrefix(s, "data:image/") {
			return ""
		}
		return template.URL(s)
	},
	"home":      func() template.URL { return template.URL(homeURL) },
	"homeLabel": func() string { return strings.TrimPrefix(homeURL, "https://") },

	// site builds an https href from a typed address (missing scheme means https, non-http(s) is not linked); it returns template.URL (unescaped), so the check must happen here.
	"site": func(s string) template.URL {
		s = strings.TrimSpace(s)
		if s == "" {
			return ""
		}
		if i := strings.Index(s, "://"); i >= 0 {
			switch strings.ToLower(s[:i]) {
			case "http", "https":
				return template.URL(s)
			default:
				return ""
			}
		}
		return template.URL("https://" + s)
	},
	"cell": func(cells []ReedCell, i int) ReedCell {
		if i < 0 || i >= len(cells) {
			return ReedCell{}
		}
		return cells[i]
	},
	"beat": func(cells []BeatCell, i int) BeatCell {
		if i < 0 || i >= len(cells) {
			return BeatCell{}
		}
		return cells[i]
	},
}

// dataOf turns "Name", value pairs into a map for go-i18n; an odd trailing arg is dropped, not panicked on.
func dataOf(pairs []any) map[string]any {
	if len(pairs) < 2 {
		return nil
	}
	out := make(map[string]any, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			continue
		}
		out[key] = pairs[i+1]
	}
	return out
}
