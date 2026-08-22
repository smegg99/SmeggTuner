package app

import (
	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/backend/core/dsp"
	audioservice "smegg.me/smeggtuner/backend/services/audio"
	tunerservice "smegg.me/smeggtuner/backend/services/tuner"
)

// observers hands the engine's stream to every service that follows it besides the frontend: the
// recorder and the automatic volume.
type observers []tunerservice.Recorder

func (o observers) OnMeasurement(m dsp.Measurement) {
	for _, r := range o {
		r.OnMeasurement(m)
	}
}

// setupCalibration hands a running engine its device's correction whenever a calibration starts,
// steps or ends; the tuner ignores a correction it already runs on.
func setupCalibration(app *application.App) {
	app.Event.On(audioservice.EventCalibration, func(*application.CustomEvent) { tunerSvc.AdoptCorrection() })
}
