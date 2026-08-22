// backend/internal/app/failed.go
package app

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/internal/trayicons"
)

// startupFailure is why the app could not start; path names the settings file when that was the cause.
type startupFailure struct {
	path string
	err  error
}

var failurePage = template.Must(template.New("failed").Parse(`<!doctype html>
<html lang="{{.Lang}}"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width">
<style>
:root { color-scheme: light dark; }
body { background: Canvas; color: CanvasText; font: 14px/1.5 system-ui, sans-serif; margin: 0; padding: 24px; }
h1 { font-size: 17px; margin: 0 0 8px; }
p { margin: 0 0 12px; }
pre { font-size: 12px; margin: 0; opacity: 0.8; overflow-wrap: anywhere; white-space: pre-wrap; }
</style></head><body>
<h1>{{.Title}}</h1>
<p>{{.Message}}</p>
{{if .File}}<p>{{.File}}</p>{{end}}
<pre>{{.Detail}}</pre>
</body></html>`))

// absolute names the settings file the way the user finds it; a relative path is the working directory's.
func absolute(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

// renderFailure is the page a startup failure shows: a heading, what to do, the settings file and the error.
func renderFailure(f startupFailure) []byte {
	loc := i18n.Localizer()
	page := struct{ Lang, Title, Message, File, Detail string }{
		Lang: i18n.Current(), Title: i18n.AppStartupFailedTitle(loc), Message: i18n.AppStartupFailedOther(loc), Detail: f.err.Error(),
	}
	if f.path != "" {
		page.Message = i18n.AppStartupFailedConfig(loc)
		page.File = i18n.AppStartupFailedFile(loc, i18n.AppStartupFailedFileParams{Path: f.path})
	}
	var out bytes.Buffer
	if err := failurePage.Execute(&out, page); err != nil {
		return []byte(template.HTMLEscapeString(f.err.Error()))
	}
	return out.Bytes()
}

// showStartupFailure opens one plain window that says why the app could not start, and returns once it is closed.
func showStartupFailure(f startupFailure) {
	page := renderFailure(f)
	app := application.New(application.Options{
		Name:   i18n.AppName(i18n.Localizer()),
		Icon:   trayicons.AppIcon,
		Logger: logger.Wails(),
		Assets: application.AssetOptions{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(page)
		})},
		Linux: application.LinuxOptions{ProgramName: "smeggtuner"},
		Mac:   application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
	})
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: i18n.AppStartupFailedTitle(i18n.Localizer()), Width: 560, Height: 320, URL: "/",
	})
	window.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) { app.Quit() })
	if err := app.Run(); err != nil {
		logger.Error(i18n.KeyLogRunFailed, logger.Err(err))
	}
}
