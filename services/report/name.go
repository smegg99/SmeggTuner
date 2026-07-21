package report

import (
	"path/filepath"
	"strings"

	corereport "smegg.me/smeggtuner/core/report"
)

func defaultName(rep *corereport.Report, format string) string {
	parts := []string{slug(rep.Identity.Session)}
	parts = append(parts, rep.Session.At.Format(fileDatestamp))

	name := strings.Trim(strings.Join(parts, "-"), "-")
	if name == "" {
		name = "tuning-report"
	}
	return name + extension(format)
}

// slug reduces a name to portable, filesystem-safe characters.
func slug(s string) string {
	var b strings.Builder
	var dash bool
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		default:
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
				dash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func extension(format string) string {
	switch format {
	case FormatCSV:
		return ".csv"
	case FormatPDF:
		return ".pdf"
	default:
		return ".html"
	}
}

// withExtension adds the format's suffix if missing, so the browser renders rather than downloads.
func withExtension(path, format string) string {
	if strings.EqualFold(filepath.Ext(path), extension(format)) {
		return path
	}
	return path + extension(format)
}
