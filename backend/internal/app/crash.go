// backend/internal/app/crash.go
package app

import (
	"path/filepath"
	"runtime"

	"github.com/smegg99/s99wails/v2/shell/crash"
	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/internal/trayicons"
)

// reportEmail receives the error screen's "Email report".
const reportEmail = "dudak799@gmail.com"

// startupFailure is why the app could not start; path names the settings file when that was the cause.
type startupFailure struct {
	path string
	err  error
}

// crashOptions starts the error screen on the app's own frontend, with no other service.
func crashOptions() crash.Options {
	return crash.Options{
		App: application.Options{
			Name:   i18n.AppName(i18n.Localizer()),
			Icon:   trayicons.AppIcon,
			Logger: logger.Wails(),
			Assets: application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
			Linux:  application.LinuxOptions{ProgramName: "smeggtuner"},
		},
		Version:     version,
		ReportEmail: reportEmail,
		Width:       minWidth,
		Height:      minHeight,
	}
}

// watchCrashes runs first: in the watcher copy of the process it never returns.
func watchCrashes() {
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		return
	}
	if err := crash.Watch(crashOptions()); err != nil {
		logger.Warn(i18n.KeyLogCrashWatchFailed, logger.Err(err))
	}
}

// describeCrash tells the error screen the language, the log file and fields once config is loaded.
func describeCrash(fields ...crash.Field) {
	cfg := config.Get()
	d := crash.Details{Lang: i18n.Current(), Fields: fields}
	if cfg.Logger.EnableFiles && cfg.Logger.Dir != "" {
		d.LogFile = absolute(filepath.Join(cfg.Logger.Dir, cfg.Logger.LogName))
	}
	crash.Describe(d)
}

// absolute names a file the way the user finds it; a relative path is the working directory's.
func absolute(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

// showStartupFailure opens the error screen in this process and returns once it is closed.
func showStartupFailure(f startupFailure) {
	failure := crash.Failure{Kind: crash.KindStartup, Err: f.err.Error()}
	if f.path != "" {
		failure.Kind, failure.File = crash.KindConfig, f.path
	}
	if err := crash.Show(crashOptions(), failure); err != nil {
		logger.Error(i18n.KeyLogCrashScreenFailed, logger.Err(err))
	}
}
