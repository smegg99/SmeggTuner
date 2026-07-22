// assets/tray/platform.go

package trayicons

import (
	"fmt"
	"runtime"

	"github.com/smegg99/s99wails/tray"
)

// PlatformIcons is ForStatus with the size the platform's tray actually wants.
// Windows GDI (CreateIconFromResourceEx) rescales whatever PNG it gets to the
// small-icon metric with no filtering, so it must be handed the 16px render;
// everywhere else the 64px primary stays sharp under the panel's own scaler.
func PlatformIcons(status Status) tray.IconsFunc {
	if runtime.GOOS != "windows" {
		return ForStatus(status)
	}
	return func(dark bool) ([]byte, error) {
		theme := "light"
		if dark {
			theme = "dark"
		}
		return FS.ReadFile(fmt.Sprintf("%s/linux-png-fallback/smeggtuner-tray-%s-%s-16px.png", theme, theme, status))
	}
}
