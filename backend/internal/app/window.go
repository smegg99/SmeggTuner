package app

import (
	browserservice "github.com/smegg99/s99wails/v2/services/browser"
	loggerservice "github.com/smegg99/s99wails/v2/services/logger"
	themeservice "github.com/smegg99/s99wails/v2/services/theme"
	titleservice "github.com/smegg99/s99wails/v2/services/title"
	"github.com/smegg99/s99wails/v2/shell/window"
	"github.com/smegg99/s99wails/v2/shell/windowstate"
	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/core/datastore"
	"smegg.me/smeggtuner/backend/internal/trayicons"
	configservice "smegg.me/smeggtuner/backend/services/config"
	sessionservice "smegg.me/smeggtuner/backend/services/session"
)

func createApp(onSecondInstance func(application.SecondInstanceData)) *application.App {
	return application.New(applicationOptions(onSecondInstance))
}

func applicationOptions(onSecondInstance func(application.SecondInstanceData)) application.Options {
	return application.Options{
		Name:        i18n.AppName(i18n.Localizer()),
		Description: i18n.AppDescription(i18n.Localizer()),
		Icon:        trayicons.AppIcon,
		Logger:      logger.Wails(),
		Services: []application.Service{
			application.NewService(&configservice.Service{}),
			application.NewService(loggerservice.New(logger.Underlying())),
			application.NewService(&themeservice.Service{}),
			application.NewService(titleservice.New(titleservice.Options{
				Base: i18n.AppName(i18n.Localizer()),
			})),
			application.NewService(browserservice.New(logger.Underlying())),
			application.NewService(audioSvc),
			application.NewService(sessionSvc),
			application.NewService(recordSvc),
			application.NewService(reportSvc),
			application.NewService(tunerSvc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
			// Instrument images served at /instruments/<id>/image, streamed and cached, not base64'd through a binding. See ImageMiddleware.
			Middleware: sessionservice.ImageMiddleware(sessionSvc),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID:               "me.smegg.smeggtuner",
			OnSecondInstanceLaunch: onSecondInstance,
		},
		Linux: application.LinuxOptions{
			ProgramName: "smeggtuner",
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	}
}

func createWindow(app *application.App) *application.WebviewWindow {
	var store windowstate.Store
	if application.System.IsDesktop() {
		store = datastore.WindowStore{}
	}
	return app.Window.NewWithOptions(windowOptions(store))
}

func windowOptions(store windowstate.Store) application.WebviewWindowOptions {
	opts := application.WebviewWindowOptions{
		Title:  i18n.AppName(i18n.Localizer()),
		Width:  1100,
		Height: 720,
		URL:    "/",
	}
	if store != nil {
		if _, err := windowstate.Configure(&opts, store); err != nil {
			logger.Warn(i18n.KeyLogWindowStateLoadFailed, logger.Err(err))
		}
	}
	window.Apply(&opts, window.Limits{MinWidth: 960, MinHeight: 640})
	return opts
}
