package app

import (
	"os"

	"github.com/smegg99/s99wails/env"

	"smegg.me/smeggtuner/common/logger"
)

// setEnv applies WebKitGTK display workarounds; DMABUF and XWayland flicker together on amdgpu, so SMEGGTUNER_GPU=1 toggles both, and an explicit WEBKIT_DISABLE_DMABUF_RENDERER overrides.
func setEnv() {
	// AppImage: host GPU stack is unknown, so disable DMABUF; a blank window is worse.
	if env.LaunchedViaAppImage() {
		os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
		logger.Info(logger.MsgAppImageEnv)
		env.PreferX11OnPlasmaWayland()
		return
	}

	if _, overridden := os.LookupEnv("WEBKIT_DISABLE_DMABUF_RENDERER"); overridden {
		env.PreferX11OnPlasmaWayland()
		return
	}

	// Unset, not "0": WebKit tests for presence.
	if os.Getenv("SMEGGTUNER_GPU") == "1" {
		os.Unsetenv("WEBKIT_DISABLE_DMABUF_RENDERER")
		logger.Info(logger.MsgGPUPath)
		return
	}

	os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	if env.PreferX11OnPlasmaWayland() {
		logger.Info(logger.MsgX11Fallback)
	}
}
