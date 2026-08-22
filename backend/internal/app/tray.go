package app

import (
	themeservice "github.com/smegg99/s99wails/v2/services/theme"
	"github.com/smegg99/s99wails/v2/shell/tray"
	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/internal/trayicons"
)

func setupSystemTray(app *application.App, window *application.WebviewWindow, quit func()) *tray.Manager {
	manager := tray.NewManager(app, tray.Options{
		Icons:              trayicons.ForStatus(trayicons.StatusIdle),
		TemplateIcon:       trayicons.TemplateIcon,
		Tooltip:            func() string { return i18n.TrayTooltip(i18n.Localizer()) },
		OnClick:            func() { tray.Toggle(window) },
		OnError:            func(err error) { logger.Error(i18n.KeyLogTrayIconFailed, logger.Err(err)) },
		ThemeChangedEvents: []string{themeservice.EventThemeChanged},
	})
	manager.RegisterEvents(app)
	manager.SetMenu(func() []tray.Item {
		return []tray.Item{
			{Label: i18n.TrayShow(i18n.Localizer()), OnClick: func() { tray.Show(window) }},
			{Separator: true},
			{Label: i18n.TrayQuit(i18n.Localizer()), OnClick: quit},
		}
	})

	return manager
}

func setupCloseToTray(window *application.WebviewWindow, quitting func() bool) {
	tray.CloseToTray(window, func() bool {
		return config.Get().Preferences.CloseToTray
	}, quitting)
}

func setupLocaleSync(trayMgr *tray.Manager) {
	i18n.OnChange(trayMgr.RebuildMenu)
}
