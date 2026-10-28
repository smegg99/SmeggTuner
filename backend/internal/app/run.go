// backend/internal/app/run.go
package app

import (
	"sync/atomic"

	themeservice "github.com/smegg99/s99wails/v2/services/theme"
	"github.com/smegg99/s99wails/v2/shell/crash"
	"github.com/smegg99/s99wails/v2/shell/editkeys"
	"github.com/smegg99/s99wails/v2/shell/tray"
	"github.com/smegg99/s99wails/v2/shell/windowstate"
	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/core/datastore"
)

func runApplication() error {
	setEnv()
	seedAudioSource()

	var window *application.WebviewWindow
	app := createApp(func(data application.SecondInstanceData) {
		logger.Info(i18n.KeyLogSecondInstance, logger.Any("args", data.Args))
		tray.Show(window)
	})

	window = createWindow(app)
	stopDesktop := setupDesktop(app, window)
	defer stopDesktop()
	stopMobile := setupMobile(app)
	defer stopMobile()
	setupRecordSessionState(app)
	setupRecordEngine(app)
	setupCalibration(app)

	return app.Run()
}

func setupDesktop(app *application.App, window *application.WebviewWindow) func() {
	if !application.System.IsDesktop() {
		return func() {}
	}
	editkeys.Watch(window, func(err error) { logger.Warn(i18n.KeyLogRunFailed, logger.Err(err)) })
	stopBeat := crash.Beat()

	stopTracking := windowstate.Track(window, datastore.WindowStore{}, windowstate.Options{
		OnError: func(err error) {
			logger.Warn(i18n.KeyLogWindowStateSaveFailed, logger.Err(err))
		},
	})

	var quitting atomic.Bool
	trayMgr := setupSystemTray(app, window, func() {
		quitting.Store(true)
		app.Quit()
	})

	setupCloseToTray(window, quitting.Load)
	setupLocaleSync(trayMgr)
	stopThemeWatcher := themeservice.RegisterThemeWatcher(app)
	return func() {
		stopBeat()
		stopThemeWatcher()
		trayMgr.Close()
		stopTracking()
	}
}

// seedAudioSource restores the last-picked capture device; a now-unplugged one falls back to the default, not an error.
func seedAudioSource() {
	id := config.Get().Audio.DeviceID
	if id == "" {
		return
	}
	if err := audioSvc.SelectMic(id); err != nil {
		logger.Warn(i18n.KeyLogAudioDeviceGone, logger.Str("device_id", id), logger.Err(err))
	}
}

// initDatastore opens the DB; its schema is the GORM models.
func initDatastore(dbPath string) error {
	if err := datastore.Initialize(dbPath); err != nil {
		return err
	}
	logger.Debug(i18n.KeyLogDatastoreReady, logger.Str("db_path", dbPath))
	return nil
}
