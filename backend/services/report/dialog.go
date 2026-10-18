// Desktop seams (save dialog, browser); tests swap these vars, and both no-op with no running app.
package report

import (
	"path/filepath"

	"smegg.me/smeggtuner/backend/common/i18n"

	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/internal/documents"
)

var (
	saveDialog    = realSaveDialog
	openInBrowser = realOpenInBrowser
)

func realSaveDialog(name, format string) (string, error) {
	app := application.Get()
	if app == nil {
		return "", nil
	}
	filter, glob := htmlFilter, htmlGlob
	switch format {
	case FormatCSV:
		filter, glob = csvFilter, csvGlob
	case FormatPDF:
		filter, glob = pdfFilter, pdfGlob
	}
	path, err := documents.SaveFile(documents.SaveOptions{
		Name: name, Title: saveTitle, Filter: filter, Pattern: glob, MIME: mimeType(format),
	})
	if err != nil {
		logger.Warn(i18n.KeyLogReportFailed, logger.Err(err))
		return "", ErrWriteFailed
	}
	return path, nil
}

func realOpenInBrowser(path string) (bool, error) {
	app := application.Get()
	if app == nil {
		return false, nil
	}
	abs, err := filepath.Abs(path)
	if documents.IsURI(path) {
		abs, err = path, nil
	}
	if err != nil {
		return false, err
	}
	if err := documents.Open(abs, "text/html"); err != nil {
		return false, err
	}
	return true, nil
}

func mimeType(format string) string {
	switch format {
	case FormatCSV:
		return "text/csv"
	case FormatPDF:
		return "application/pdf"
	default:
		return "text/html"
	}
}
