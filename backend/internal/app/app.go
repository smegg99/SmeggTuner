// backend/internal/app/app.go
package app

import (
	"embed"
	"errors"

	"smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/core/datastore"
	audioservice "smegg.me/smeggtuner/backend/services/audio"
	recordservice "smegg.me/smeggtuner/backend/services/record"
	reportservice "smegg.me/smeggtuner/backend/services/report"
	sessionservice "smegg.me/smeggtuner/backend/services/session"
	tunerservice "smegg.me/smeggtuner/backend/services/tuner"
)

// Embedded beside main.go and handed in by it.
var assets embed.FS

// Services are safe before config.Initialize; cleanup also owns the session saver started by New.
var (
	audioSvc   = audioservice.New()
	sessionSvc = sessionservice.New()
	recordSvc  = recordservice.New(sessionSvc)
	reportSvc  = reportservice.New(sessionSvc)
	tunerSvc   = tunerservice.New(audioservice.NewTunerInput(audioSvc), sessionSvc, observers{recordSvc, audioSvc})
)

// Main wires and runs the desktop app, returning the process exit code.
func Main(embeddedAssets embed.FS) int {
	assets = embeddedAssets
	return run(runApplication, showStartupFailure)
}

// run starts the app with launch; a failure before launch is logged and shown with failed.
func run(launch func() error, failed func(startupFailure)) (exitCode int) {
	defer func() {
		if err := cleanup(); err != nil {
			exitCode = 1
		}
	}()

	logger.Initialize()
	logger.Info(i18n.KeyLogAppStarting)

	logger.Debug(i18n.KeyLogLoadingConfig)
	path, err := config.Initialize()
	if err != nil {
		logger.Error(i18n.KeyLogConfigLoadFailed, logger.Err(err))
		failed(startupFailure{path: absolute(path), err: err})
		return 1
	}
	logger.Debug(i18n.KeyLogConfigLoaded, logger.Str("config_path", path))

	cfg := config.Get()

	if cfg.Preferences.Language != "" {
		i18n.SetLocale(cfg.Preferences.Language)
	}

	logger.Debug(i18n.KeyLogReinitLogger)
	if err := logger.Configure(logger.Options{
		Verbose:     cfg.Logger.Verbose,
		NoColor:     cfg.Logger.NoColor,
		Level:       cfg.Logger.Level,
		Prefix:      cfg.Logger.Prefix,
		EnableFiles: cfg.Logger.EnableFiles,
		Dir:         cfg.Logger.Dir,
		LogName:     cfg.Logger.LogName,
		MaxSizeMb:   cfg.Logger.MaxSizeMb,
		MaxBackups:  cfg.Logger.MaxBackups,
		MaxAgeDays:  cfg.Logger.MaxAgeDays,
		Compression: cfg.Logger.Compression,
		LocalTime:   cfg.Logger.LocalTime,
	}); err != nil {
		logger.Error(i18n.KeyLogLoggerInitFailed, logger.Err(err))
		failed(startupFailure{err: err})
		return 1
	}

	if err := initDatastore(cfg.Storage.DBPath); err != nil {
		logger.Error(i18n.KeyLogDatastoreInitFailed, logger.Err(err))
		failed(startupFailure{err: err})
		return 1
	}

	if err := launch(); err != nil {
		logger.Error(i18n.KeyLogRunFailed, logger.Err(err))
		return 1
	}
	return 0
}

func cleanup() error {
	logger.Info(i18n.KeyLogCleaningUp)
	// Wails may have stopped services already; partial startup still needs every owner closed.
	var failures []error
	for _, operation := range []struct {
		name  string
		close func() error
	}{
		{"tuner", tunerSvc.ServiceShutdown},
		{"audio", audioSvc.ServiceShutdown},
		{"session", sessionSvc.ServiceShutdown},
		{"datastore", datastore.Close},
	} {
		if err := operation.close(); err != nil {
			logger.Error(i18n.KeyLogRunFailed, logger.Str("shutdown", operation.name), logger.Err(err))
			failures = append(failures, err)
		}
	}
	return errors.Join(append(failures, logger.Close())...)
}
