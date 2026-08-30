package report

import (
	"embed"
	"html/template"
	"io"
)

//go:embed templates/*.gohtml
var templates embed.FS

// The mascot, as one ink path. It is the site's own art (SmeggTunerSite web/app/assets/img),
// recoloured to currentColor: the source fill is #0f1218, a blue-black, and this sheet is
// grayscale by rule.
//
//go:embed assets/mascot.svg
var mascotSVG string

const homeURL = "https://smeggtuner.com"

var tpl = template.Must(
	template.New("report").Funcs(funcs).ParseFS(templates, "templates/*.gohtml"),
)

// HTML writes the self-contained sheet, the same one services/report renders to PDF.
func HTML(w io.Writer, r *Report) error {
	if r == nil {
		return ErrNoSession
	}
	return tpl.ExecuteTemplate(w, "report.gohtml", r)
}
