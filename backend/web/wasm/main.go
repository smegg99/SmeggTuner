//go:build js

// backend/web/wasm/main.go

// Command wasm is the web edition's engine, run in a browser worker; globalThis.smeggtunerInit carries {config, rate}.
package main

import (
	"syscall/js"

	"smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/common/logger"
	configservice "smegg.me/smeggtuner/backend/services/config"
	"smegg.me/smeggtuner/backend/services/tuner"
	"smegg.me/smeggtuner/backend/web"
)

func main() {
	init := js.Global().Get("smeggtunerInit")
	logger.Initialize()
	config.Seed([]byte(init.Get("config").String()))
	if _, err := config.Initialize(); err != nil {
		logger.Warn(i18n.KeyLogConfigLoadFailed, logger.Err(err))
	}
	if lang := config.Get().Preferences.Language; lang != "" {
		i18n.SetLocale(lang)
	}

	src := web.NewBrowserSource(init.Get("rate").Int())
	tunerSvc := tuner.New(web.NewInput(src), nil, nil)
	web.Serve(web.NewDispatcher(tunerSvc, &configservice.Service{}), src)
	select {}
}
