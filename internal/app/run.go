package app

import (
	"path/filepath"
	"sync/atomic"

	themeservice "github.com/smegg99/s99wails/services/theme"
	"github.com/smegg99/s99wails/tray"
	"github.com/smegg99/s99wails/windowstate"
	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/common/config"
	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/datastore"
	"smegg.me/smeggtuner/core/repositories"
)

func runApplication() {
	setEnv()
	seedAudioSource()

	var window *application.WebviewWindow
	app := createApp(func(data application.SecondInstanceData) {
		logger.Info(logger.MsgSecondInstance, logger.Any("args", data.Args))
		tray.Show(window)
	}, func() *application.WebviewWindow { return window })

	window = createWindow(app)

	// Restore defers to the window's first show; geometry calls need the main thread Run() creates.
	stopTracking, err := windowstate.Manage(window, datastore.WindowStore{}, windowstate.Options{
		OnError: func(err error) {
			logger.Warn(logger.MsgWindowStateSaveFailed, logger.Err(err))
		},
	})
	if err != nil {
		logger.Warn(logger.MsgWindowStateLoadFailed, logger.Err(err))
	}
	defer stopTracking()

	var quitting atomic.Bool
	trayMgr := setupSystemTray(app, window, func() {
		quitting.Store(true)
		app.Quit()
	})
	defer trayMgr.Close()

	setupCloseToTray(window, quitting.Load)
	setupLocaleSync(trayMgr)
	setupRecordTrayState(app, trayMgr)
	setupRecordSessionState(app)
	setupRecordRecalibrate(app)

	stopThemeWatcher := themeservice.RegisterThemeWatcher(app)
	defer stopThemeWatcher()

	if err := app.Run(); err != nil {
		logger.Fatal(logger.MsgRunFailed, logger.Err(err))
	}
}

// seedAudioSource restores the last-picked capture device; a now-unplugged one falls back to the default, not an error.
func seedAudioSource() {
	id := config.Get().Audio.DeviceID
	if id == "" {
		return
	}
	if err := audioSvc.SelectMic(id); err != nil {
		logger.Warn(logger.MsgAudioDeviceGone, logger.Str("device_id", id), logger.Err(err))
	}
}

// initDatastore opens the DB and, on first run, imports the pre-DB legacy files, which are left in place.
func initDatastore(dbPath, configPath string) {
	if err := datastore.Initialize(dbPath); err != nil {
		logger.Fatal(logger.MsgDatastoreInitFailed, logger.Err(err))
	}
	logger.Debug(logger.MsgDatastoreReady, logger.Str("db_path", dbPath))

	configDir := filepath.Dir(configPath)
	stats, err := repositories.ImportLegacy(configDir)
	if err != nil {
		logger.Warn(logger.MsgLegacyImportFailed, logger.Err(err))
	} else if !stats.Empty() {
		logger.Info(logger.MsgLegacyImported,
			logger.Int("sessions", stats.Sessions),
			logger.Int("instruments", stats.Instruments),
			logger.Int("skipped", stats.Skipped))
	}
	if err := datastore.ImportLegacyWindowState(filepath.Join(configDir, "windowstate.json")); err != nil {
		logger.Warn(logger.MsgLegacyImportFailed, logger.Err(err))
	}
}
