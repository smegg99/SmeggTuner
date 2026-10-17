package audio

import "github.com/wailsapp/wails/v3/pkg/application"

// Events the frontend follows.
const (
	EventCalibration = "audio:calibration" // CalibrationDTO: every estimator step of a session and its end
	EventVolume      = "audio:volume"      // AutoVolumeDTO: automatic volume switched, started or stopped
)

// emitEvent hands an event to the frontend; a variable so tests can listen.
var emitEvent = func(name string, data any) {
	if app := application.Get(); app != nil {
		app.Event.Emit(name, data)
	}
}
