// backend/internal/app/display.go
package app

import (
	"github.com/smegg99/s99wails/v2/platform/env"
	"github.com/smegg99/s99wails/v2/shell/crash"
)

func setEnv() {
	res := env.Setup(env.Options{AppImageRenderPath: env.RenderPathNoGBM})
	describeCrash(crash.Field{Name: "Render path", Value: res.RenderPath.String() + " (" + res.RenderPathFrom + ")"})
}
