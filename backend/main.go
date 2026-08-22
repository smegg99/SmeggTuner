package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/backend/internal/androidhost"
	"smegg.me/smeggtuner/backend/internal/app"
)

//go:embed all:frontend/public
var assets embed.FS

func main() {
	flush := func() {}
	if runtime.GOOS == "android" {
		flush = androidhost.CaptureOutput()
		// Relative config, database and log paths must stay in the app sandbox.
		path := application.Mobile.StoragePath()
		if path == "" {
			fmt.Fprintln(os.Stderr, "Android app storage path is unavailable")
			return
		}
		if err := os.Chdir(path); err != nil {
			fmt.Fprintln(os.Stderr, "open Android app storage:", err)
			return
		}
		cache, err := androidhost.CacheDirectory(context.Background())
		if err != nil || cache == "" {
			fmt.Fprintln(os.Stderr, "open Android cache:", err)
			return
		}
		for key, value := range map[string]string{
			"TMPDIR":          cache,
			"XDG_CONFIG_HOME": filepath.Join(path, "config"),
			"XDG_CACHE_HOME":  cache,
		} {
			if err := os.Setenv(key, value); err != nil {
				fmt.Fprintln(os.Stderr, "set Android storage environment:", err)
				return
			}
		}
	}
	code := app.Main(assets)
	flush()
	os.Exit(code)
}
