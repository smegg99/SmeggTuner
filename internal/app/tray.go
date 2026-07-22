package app

import (
	"sync/atomic"

	themeservice "github.com/smegg99/s99wails/services/theme"
	"github.com/smegg99/s99wails/tray"
	"github.com/wailsapp/wails/v3/pkg/application"

	trayicons "smegg.me/smeggtuner/assets/tray"
	"smegg.me/smeggtuner/common/config"
	"smegg.me/smeggtuner/common/i18n"
	"smegg.me/smeggtuner/common/logger"
	recordservice "smegg.me/smeggtuner/services/record"
	sessionservice "smegg.me/smeggtuner/services/session"
)

func setupSystemTray(app *application.App, window *application.WebviewWindow, quit func()) *tray.Manager {
	systemTray := app.SystemTray.New()

	manager := tray.NewManager(systemTray, tray.Options{
		Icons:   trayicons.PlatformIcons(trayicons.StatusIdle),
		Tooltip: func() string { return i18n.T("tray.tooltip") },
		OnError: func(err error) {
			logger.Error(logger.MsgTrayIconFailed, logger.Err(err))
		},
		ThemeChangedEvents: []string{themeservice.EventThemeChanged},
	})
	manager.RegisterEvents(app)

	manager.SetMenu(func() *application.Menu {
		menu := app.NewMenu()
		menu.Add(i18n.T("tray.show")).OnClick(func(_ *application.Context) {
			tray.Show(window)
		})
		menu.AddSeparator()
		menu.Add(i18n.T("tray.quit")).OnClick(func(_ *application.Context) {
			quit()
		})
		return menu
	})

	tray.OnClickToggle(systemTray, window)
	return manager
}

func setupCloseToTray(window *application.WebviewWindow, quitting func() bool) {
	tray.CloseToTray(window, func() bool {
		return config.Get().Preferences.CloseToTray
	}, quitting)
}

// setupRecordTrayState lights the tray badge while record is armed; record stays ignorant of the tray.
func setupRecordTrayState(app *application.App, manager *tray.Manager) {
	app.Event.On(recordservice.EventState, func(e *application.CustomEvent) {
		state, ok := e.Data.(recordservice.StateDTO)
		if !ok {
			return
		}
		// Armed with a session means readings are landing; warm-up saves nothing.
		status := trayicons.StatusIdle
		if state.SessionID != "" && state.Armed {
			status = trayicons.StatusRecording
		}
		manager.SetIcons(trayicons.PlatformIcons(status))
	})
}

// setupRecordRecalibrate restarts the engine's noise-floor warm-up on the rising edge of Armed; wired in the composition root so record stays ignorant of the tuner.
func setupRecordRecalibrate(app *application.App) {
	var wasArmed atomic.Bool
	app.Event.On(recordservice.EventState, func(e *application.CustomEvent) {
		state, ok := e.Data.(recordservice.StateDTO)
		if !ok {
			return
		}
		if prev := wasArmed.Swap(state.Armed); state.Armed && !prev {
			tunerSvc.Recalibrate()
		}
	})
}

// setupRecordSessionState calls SessionChanged (not PublishState) on session change, since a new session starts in warm-up.
func setupRecordSessionState(app *application.App) {
	app.Event.On(sessionservice.EventActive, func(*application.CustomEvent) {
		recordSvc.SessionChanged()
	})
}

func setupLocaleSync(trayMgr *tray.Manager) {
	i18n.OnChange(trayMgr.RebuildMenu)
}
