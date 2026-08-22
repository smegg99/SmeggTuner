// backend/internal/app/display.go
package app

import "github.com/smegg99/s99wails/v2/platform/env"

func setEnv() {
	env.Setup(env.Options{AppImageRenderPath: env.RenderPathNoGBM})
}
