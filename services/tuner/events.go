package tuner

import (
	"github.com/wailsapp/wails/v3/pkg/application"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/target"
	audiosvc "smegg.me/smeggtuner/services/audio"
)

const (
	// EventMeasurement carries a MeasurementDTO on two cadences: a ~12/s heartbeat (empty Reeds/Beats) and a ~4/s fine result.
	EventMeasurement = "tuner:measurement"
	// EventState carries a StateDTO on engine lifecycle transitions only, never per tick.
	EventState = "tuner:state"
	// EventSettings carries a SettingsDTO when the engine adopts a pitch or reed count the UI did not set (session opened, closed or edited).
	EventSettings = "tuner:settings"
	// EventPlayback carries a PlaybackDTO ~30/s while reading a file; its own event because it runs on a different clock than measurements.
	EventPlayback = "playback:position"
)

// PlaybackDTO is the file transport's playhead, in seconds.
type PlaybackDTO struct {
	Position float64 `json:"position"`
	Paused   bool    `json:"paused"`
	// Moving: sound is actually leaving the speakers now; the needle coasts on this alone.
	Moving bool `json:"moving"`
}

// StateDTO is the engine's lifecycle as the UI sees it; Error is an i18n key, never a raw Go error.
type StateDTO struct {
	Running bool   `json:"running"`
	Source  string `json:"source"` // display name of the current input
	Error   string `json:"error"`  // i18n key, empty when fine
}

// MeasurementDTO is a Measurement and the goal it is measured against; with no session the goal is the empty curve, every Goal zero and every Error the plain deviation from the scale.
type MeasurementDTO struct {
	dsp.Measurement
	// Equalizer, Spectrum and Waveform shadow the Measurement's float arrays so encoding/json emits these packed bytes, not the floats. See pixels.go.
	Equalizer pixels `json:"equalizer"`
	Spectrum  pixels `json:"spectrum"`
	Waveform  pixels `json:"waveform"`
	// ReedErrors is one row per reed: what it reads, what the curve asks, the difference. Empty on a heartbeat.
	ReedErrors []target.ReedError `json:"reedErrors"`
	// BeatErrors is the same for the beat between every pair of reeds.
	BeatErrors []target.BeatError `json:"beatErrors"`
}

// SettingsDTO is what the engine is measuring with right now.
type SettingsDTO struct {
	A4 float64 `json:"a4"`
	// ReedCount is the pulled register's bank count when a session is open, one otherwise.
	ReedCount int `json:"reedCount"`
	// SessionReeds is what the open session's instrument sounds, 0 when none is open.
	SessionReeds  int     `json:"sessionReeds"`
	Tolerance     float64 `json:"tolerance"`
	BeatTolerance float64 `json:"beatTolerance"`
}

// init registers the payloads so the binding generator emits TypeScript typings for the events.
func init() {
	application.RegisterEvent[MeasurementDTO](EventMeasurement)
	application.RegisterEvent[StateDTO](EventState)
	application.RegisterEvent[SettingsDTO](EventSettings)
	application.RegisterEvent[PlaybackDTO](EventPlayback)
}

// ServiceError is the i18n-keyed error shape both services hand the frontend; it is the audio service's type, so a Build failure travels through unchanged.
type ServiceError = audiosvc.ServiceError

var (
	// ErrInvalidA4 reports a reference pitch outside 430..450 Hz.
	ErrInvalidA4 = &ServiceError{Key: "tuner.error.invalidA4"}
	// ErrInvalidNote reports a note outside the tracked range (0 means auto).
	ErrInvalidNote = &ServiceError{Key: "tuner.error.invalidNote"}
	// ErrInvalidTranspose reports a transposition beyond two octaves.
	ErrInvalidTranspose = &ServiceError{Key: "tuner.error.invalidTranspose"}
	// ErrPlaybackFailed reports a reference tone that could not be played.
	ErrPlaybackFailed = &ServiceError{Key: "tuner.error.playbackFailed"}
	// ErrDeviceLost reports a capture device that died mid-run.
	ErrDeviceLost = &ServiceError{Key: "tuner.error.deviceLost"}
)

// emitEvent is the single door to the frontend and the seam tests replace; a no-op when no application is running.
var emitEvent = func(name string, data any) {
	if app := application.Get(); app != nil {
		app.Event.Emit(name, data)
	}
}
