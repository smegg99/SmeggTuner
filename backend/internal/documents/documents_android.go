//go:build android

// backend/internal/documents/documents_android.go
package documents

import (
	"fmt"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/backend/internal/androidhost"
)

// SaveFile returns the document URI granted by Android's storage access framework.
func SaveFile(opts SaveOptions) (string, error) {
	app := application.Get()
	if app == nil {
		return "", nil
	}
	return androidhost.SaveDocument(app.Context(), opts.Name, opts.Title, opts.MIME)
}

// Write stages a complete file before opening the selected content URI for truncation.
func Write(path string, render func(string) error) error {
	if !IsURI(path) {
		return render(path)
	}
	app := application.Get()
	if app == nil {
		return fmt.Errorf("document write requested without an application")
	}
	file, err := os.CreateTemp(os.TempDir(), "smeggtuner-export-*")
	if err != nil {
		return err
	}
	staged := file.Name()
	defer os.Remove(staged)
	if err := file.Close(); err != nil {
		return err
	}
	if err := render(staged); err != nil {
		return err
	}
	return androidhost.WriteDocument(app.Context(), path, staged)
}

// Open grants the chosen Android viewer temporary access to the document.
func Open(path, mime string) error {
	app := application.Get()
	if app == nil {
		return nil
	}
	return androidhost.OpenDocument(app.Context(), path, mime)
}
