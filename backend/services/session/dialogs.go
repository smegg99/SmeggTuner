// Native file pickers. Dialog words come from the caller (locale); the path is validated by the
// matching import/set call, not here.
package session

import (
	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/backend/common/i18n"

	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/internal/documents"
)

const (
	sessionGlob          = "*.stsf"
	instrumentFilterGlob = "*.stif"
)

// OpenFileDialog shows the native session picker and returns the chosen path, or "" if cancelled.
func (s *Service) OpenFileDialog(title, filterName string) (string, error) {
	app := application.Get()
	if app == nil {
		return "", nil // no main thread to dispatch onto (unit tests, headless tooling)
	}
	path, err := app.Dialog.OpenFile().
		SetTitle(title).
		CanChooseFiles(true).
		CanChooseDirectories(false).
		AddFilter(filterName, sessionGlob).
		PromptForSingleSelection()
	if err != nil {
		logger.Warn(i18n.KeyLogSessionLoadFailed, logger.Err(err))
		return "", ErrLoadFailed
	}
	return path, nil
}

// OpenInstrumentDialog is OpenFileDialog for instrument files.
func (s *Service) OpenInstrumentDialog(title, filterName string) (string, error) {
	app := application.Get()
	if app == nil {
		return "", nil
	}

	path, err := app.Dialog.OpenFile().
		SetTitle(title).
		CanChooseFiles(true).
		CanChooseDirectories(false).
		AddFilter(filterName, instrumentFilterGlob).
		PromptForSingleSelection()
	if err != nil {
		logger.Warn(i18n.KeyLogSessionLoadFailed, logger.Err(err))
		return "", ErrLoadFailed
	}
	return path, nil
}

// SaveFileDialog asks where to write an export, "" if cancelled. kind picks the extension.
func (s *Service) SaveFileDialog(kind, suggested, title, filterName string) (string, error) {
	app := application.Get()
	if app == nil {
		return "", nil
	}

	glob := sessionGlob
	if kind == "instrument" {
		glob = instrumentFilterGlob
	}

	path, err := documents.SaveFile(documents.SaveOptions{
		Name: suggested, Title: title, Filter: filterName,
		Pattern: glob, MIME: "application/octet-stream",
	})
	if err != nil {
		logger.Warn(i18n.KeyLogSessionSaveFailed, logger.Err(err))
		return "", ErrSaveFailed
	}
	return path, nil
}

// OpenImageDialog picks a photograph; SetInstrumentImage does the decoding and capping.
func (s *Service) OpenImageDialog(title, filterName string) (string, error) {
	app := application.Get()
	if app == nil {
		return "", nil
	}

	path, err := app.Dialog.OpenFile().
		SetTitle(title).
		CanChooseFiles(true).
		CanChooseDirectories(false).
		AddFilter(filterName, "*.jpg;*.jpeg;*.JPG;*.JPEG;*.png;*.PNG").
		PromptForSingleSelection()
	if err != nil {
		logger.Warn(i18n.KeyLogSessionLoadFailed, logger.Err(err))
		return "", ErrImageUnreadable
	}
	return path, nil
}
