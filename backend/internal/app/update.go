// backend/internal/app/update.go
package app

import (
	"github.com/smegg99/s99wails/v2/shell/update"

	"smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/core/datastore"
	"smegg.me/smeggtuner/backend/internal/releasenotes"
)

var updateSvc = update.New(updateOptions())

func updateOptions() update.Options {
	return update.Options{
		Owner:       "smegg99",
		Repo:        "SmeggTuner",
		Version:     version,
		DownloadURL: "https://smeggtuner.com/#downloads",
		Changelog:   releasenotes.Changelog,
		Store:       datastore.UpdateStore{},
		Enabled:     func() bool { return config.Get().Preferences.CheckUpdates },
		Existing:    datastore.HasHistory,
		OnError: func(err error) {
			logger.Warn(i18n.KeyLogUpdateCheckFailed, logger.Err(err))
		},
	}
}
