//go:build !js

// backend/services/tuner/events_desktop.go
package tuner

import "github.com/wailsapp/wails/v3/pkg/application"

// init registers the payloads so the binding generator emits TypeScript typings for the events.
func init() {
	application.RegisterEvent[MeasurementDTO](EventMeasurement)
	application.RegisterEvent[StateDTO](EventState)
	application.RegisterEvent[SettingsDTO](EventSettings)
	application.RegisterEvent[PlaybackDTO](EventPlayback)
}
