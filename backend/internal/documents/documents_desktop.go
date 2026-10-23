//go:build !android

// backend/internal/documents/documents_desktop.go
package documents

import "github.com/wailsapp/wails/v3/pkg/application"

// SaveFile returns a destination chosen in the platform save dialog.
func SaveFile(opts SaveOptions) (string, error) {
	app := application.Get()
	if app == nil {
		return "", nil
	}
	dialog := app.Dialog.SaveFile()
	dialog.SetOptions(&application.SaveFileDialogOptions{
		Title:                opts.Title,
		Filename:             opts.Name,
		CanCreateDirectories: true,
		Filters: []application.FileFilter{
			{DisplayName: opts.Filter, Pattern: opts.Pattern},
		},
	})
	return dialog.PromptForSingleSelection()
}

// Write renders directly to the selected local destination.
func Write(path string, render func(string) error) error { return render(path) }

// Open hands a local document to its default viewer.
func Open(path, mime string) error {
	app := application.Get()
	if app == nil {
		return nil
	}
	return app.Browser.OpenFile(path)
}
