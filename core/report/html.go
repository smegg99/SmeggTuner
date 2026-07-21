package report

import (
	"embed"
	"encoding/base64"
	"errors"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
)

//go:embed templates/*.gohtml
var templates embed.FS

var (
	// ErrLogoUnreadable reports a logo file that would not open.
	ErrLogoUnreadable = errors.New("report: logo could not be read")
	// ErrLogoNotImage reports a logo file that is not an image a browser will draw.
	ErrLogoNotImage = errors.New("report: logo is not an image")
)

// maxLogoBytes caps the logo so a mailed report stays small enough to send.
const maxLogoBytes = 1 << 20

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

// LoadLogo reads an image off disk and encodes it as a data URI for the letterhead.
func LoadLogo(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", nil
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "", ErrLogoUnreadable
	}
	if info.Size() > maxLogoBytes {
		return "", ErrLogoNotImage
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", ErrLogoUnreadable
	}
	mime := http.DetectContentType(raw)
	if !strings.HasPrefix(mime, "image/") {
		return "", ErrLogoNotImage
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(raw), nil
}
