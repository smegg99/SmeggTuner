// Desktop seams (save dialog, browser); tests swap these vars, and both no-op with no running app.
package report

import (
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/common/logger"
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
	// The save dialog has no SetTitle in the builder, only in its options struct.
	dialog := app.Dialog.SaveFile()
	dialog.SetOptions(&application.SaveFileDialogOptions{
		Title:                saveTitle,
		Filename:             name,
		CanCreateDirectories: true,
		Filters:              []application.FileFilter{{DisplayName: filter, Pattern: glob}},
	})
	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		logger.Warn(logger.MsgReportFailed, logger.Err(err))
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
	if err != nil {
		return false, err
	}
	if err := app.Browser.OpenURL("file://" + abs); err != nil {
		return false, err
	}
	return true, nil
}
