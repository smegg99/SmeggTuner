package app

import (
	"embed"

	"smegg.me/smeggtuner/common/config"
	"smegg.me/smeggtuner/common/i18n"
	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/datastore"
	audioservice "smegg.me/smeggtuner/services/audio"
	recordservice "smegg.me/smeggtuner/services/record"
	reportservice "smegg.me/smeggtuner/services/report"
	sessionservice "smegg.me/smeggtuner/services/session"
	tunerservice "smegg.me/smeggtuner/services/tuner"
)

// Embedded at the module root (go:embed can't escape its dir) and handed in by main.
var (
	assets  embed.FS
	appIcon []byte
)

// Constructors are side-effect free, safe before config.Initialize; wiring is one-way, nothing points back.
var (
	audioSvc   = audioservice.New()
	sessionSvc = sessionservice.New()
	recordSvc  = recordservice.New(sessionSvc)
	reportSvc  = reportservice.New(sessionSvc)
	tunerSvc   = tunerservice.New(audioSvc, sessionSvc, recordSvc)
)

// Main wires and runs the desktop app, returning the process exit code.
func Main(embeddedAssets embed.FS, icon []byte) (exitCode int) {
	assets = embeddedAssets
	appIcon = icon

	defer func() {
		if err := cleanup(); err != nil {
			exitCode = 1
		}
	}()

	logger.Initialize()
	logger.Info(logger.MsgAppStarting)

	logger.Debug(logger.MsgLoadingConfig)
	path, err := config.Initialize()
	if err != nil {
		logger.Fatal(logger.MsgConfigLoadFailed, logger.Err(err))
	}
	logger.Debug(logger.MsgConfigLoaded, logger.Str("config_path", path))

	cfg := config.Get()

	if cfg.Preferences.Language != "" {
		i18n.SetLocale(cfg.Preferences.Language)
	}

	logger.Debug(logger.MsgReinitLogger)
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
		logger.Fatal(logger.MsgLoggerInitFailed, logger.Err(err))
	}

	initDatastore(cfg.Storage.DBPath, path)

	runApplication()
	return 0
}

func cleanup() error {
	logger.Info(logger.MsgCleaningUp)
	// Stop the engine before the logger closes: a running Run loop logs, and Stop waits for it.
	if err := tunerSvc.Stop(); err != nil {
		logger.Warn(logger.MsgTunerRunFailed, logger.Err(err))
	}
	if err := datastore.Close(); err != nil {
		logger.Warn(logger.MsgDatastoreInitFailed, logger.Err(err))
	}
	return logger.Close()
}
