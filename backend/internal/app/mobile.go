// backend/internal/app/mobile.go
package app

import (
	"sync"

	"smegg.me/smeggtuner/backend/common/i18n"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"smegg.me/smeggtuner/backend/common/logger"
)

type mobileLifecycle struct {
	mu        sync.Mutex
	stop      func() error
	stopTone  func() error
	cancelCal func()
	save      func() error
	onError   func(error)
}

func (m *mobileLifecycle) suspend() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cancelCal()
	for _, operation := range []func() error{m.stopTone, m.stop, m.save} {
		if err := operation(); err != nil {
			m.onError(err)
		}
	}
}

func setupMobile(app *application.App) func() {
	if !application.System.IsPlatform(application.PlatformAndroid) {
		return func() {}
	}
	lifecycle := &mobileLifecycle{
		stop: tunerSvc.Stop, stopTone: tunerSvc.StopTone,
		cancelCal: audioSvc.CancelCalibration, save: sessionSvc.Save,
		onError: func(err error) { logger.Warn(i18n.KeyLogRunFailed, logger.Err(err)) },
	}
	remove := app.Event.OnApplicationEvent(events.Android.ActivityStopped, func(*application.ApplicationEvent) {
		lifecycle.suspend()
	})
	app.OnShutdown(lifecycle.suspend)
	return remove
}
