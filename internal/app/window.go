package app

import (
	browserservice "github.com/smegg99/s99wails/services/browser"
	loggerservice "github.com/smegg99/s99wails/services/logger"
	themeservice "github.com/smegg99/s99wails/services/theme"
	titleservice "github.com/smegg99/s99wails/services/title"
	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/common/i18n"
	"smegg.me/smeggtuner/common/logger"
	configservice "smegg.me/smeggtuner/services/config"
	sessionservice "smegg.me/smeggtuner/services/session"
)

func createApp(onSecondInstance func(application.SecondInstanceData), window func() *application.WebviewWindow) *application.App {
	return application.New(application.Options{
		Name:        i18n.T("app.name"),
		Description: i18n.T("app.description"),
		Icon:        appIcon,
		Services: []application.Service{
			application.NewService(&configservice.Service{}),
			application.NewService(loggerservice.New(logger.Underlying())),
			application.NewService(&themeservice.Service{}),
			application.NewService(titleservice.New(titleservice.Options{
				Base: i18n.T("app.name"),
				// Window does not exist yet (built from this app), so pass a getter to reach it later.
				Window: window,
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
	})
}

func createWindow(app *application.App) *application.WebviewWindow {
	return app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     i18n.T("app.name"),
		Width:     1100,
		Height:    720,
		MinWidth:  960,
		MinHeight: 640,
		URL:       "/",
	})
}
