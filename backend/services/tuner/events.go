package tuner

import (
	"smegg.me/smeggtuner/backend/common/emit"
	"smegg.me/smeggtuner/backend/common/serviceerror"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/target"
)

const (
	// EventMeasurement carries a MeasurementDTO on two cadences: a ~20/s heartbeat (the screen state, empty Reeds/Beats) and two analyses a second.
	EventMeasurement = "tuner:measurement"
	// EventState carries a StateDTO on engine lifecycle transitions only, never per tick.
	EventState = "tuner:state"
	// EventSettings carries a SettingsDTO after every tuner setting change and when a session opens, closes or is edited.
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

// MeasurementDTO is a Measurement and the goal it is measured against, in reference units; with no session every Goal is zero and every Error the plain deviation from the scale.
type MeasurementDTO struct {
	dsp.Measurement
	// Equalizer, Spectrum and Waveform shadow the Measurement's float arrays so encoding/json emits these packed bytes, not the floats. See pixels.go.
	Equalizer pixels `json:"equalizer"`
	Spectrum  pixels `json:"spectrum"`
	Waveform  pixels `json:"waveform"`
	// ReedErrors is one row per sounding slot: what it reads, what the curve asks, the difference. Empty without fresh slot values.
	ReedErrors []target.ReedError `json:"reedErrors"`
	// BeatErrors is the same for the beat between each adjacent pair of sounding slots.
	BeatErrors []target.BeatError `json:"beatErrors"`
}

// SettingsDTO is what the engine is measuring with right now.
type SettingsDTO struct {
	A4 float64 `json:"a4"`
	// ReedCount is how many slots the engine measures: the session's register, or the panel's mode.
	ReedCount int `json:"reedCount"`
	// SessionReeds is what the open session's instrument sounds, 0 when none is open.
	SessionReeds  int     `json:"sessionReeds"`
	Tolerance     float64 `json:"tolerance"`
	BeatTolerance float64 `json:"beatTolerance"`
	// Mode, ReedBoxes and OctaveOffsets are the panel; they decide only while SessionReeds is 0.
	Mode          string `json:"mode"`
	ReedBoxes     []int  `json:"reedBoxes"`
	OctaveOffsets []int  `json:"octaveOffsets"`
	// Transpose is what the engine runs on: the saved value in tones mode, 0 in chords and octaves.
	Transpose int `json:"transpose"`
	// ManualNote is the pinned key, 0 when the detector chooses.
	ManualNote int `json:"manualNote"`
}

// ServiceError is the shared i18n-keyed error shape, so a Build failure travels through unchanged.
type ServiceError = serviceerror.ServiceError

var (
	// ErrInvalidA4 reports a reference pitch outside 430..450 Hz.
	ErrInvalidA4 = &ServiceError{Key: "tuner.error.invalidA4"}
	// ErrInvalidNote reports a note outside the tracked range (0 means auto).
	ErrInvalidNote = &ServiceError{Key: "tuner.error.invalidNote"}
	// ErrInvalidTranspose reports a transposition outside 0..11 semitones.
	ErrInvalidTranspose = &ServiceError{Key: "tuner.error.invalidTranspose"}
	// ErrTransposeMode reports a transposition asked for in chords or octaves mode, which run untransposed.
	ErrTransposeMode = &ServiceError{Key: "tuner.error.transposeMode"}
	// ErrA4Locked reports an A4 change while tuner.lock_a4 is on.
	ErrA4Locked = &ServiceError{Key: "tuner.error.a4Locked"}
	// ErrInvalidMode reports a mode other than tones, octaves or chords.
	ErrInvalidMode = &ServiceError{Key: "tuner.error.invalidMode"}
	// ErrSessionDecides reports a mode switch while a session is open: its register decides.
	ErrSessionDecides = &ServiceError{Key: "tuner.error.sessionDecides"}
	// ErrInvalidBox reports a reed box outside 1..3 or an octave other than 12 or 24.
	ErrInvalidBox = &ServiceError{Key: "tuner.error.invalidBox"}
	// ErrLastBox reports turning off the last box that is on.
	ErrLastBox = &ServiceError{Key: "tuner.error.lastBox"}
	// ErrChordsAuto reports a manual note asked for in chords mode, which picks its own notes.
	ErrChordsAuto = &ServiceError{Key: "tuner.error.chordsAuto"}
	// ErrToneNeedsManual reports the tone button pressed with no manual note.
	ErrToneNeedsManual = &ServiceError{Key: "tuner.error.toneNeedsManual"}
	// ErrPlaybackFailed reports a reference tone that could not be played.
	ErrPlaybackFailed = &ServiceError{Key: "tuner.error.playbackFailed"}
	// ErrSampleRate reports a source whose sample rate cannot carry the note table at this A4 and transposition.
	ErrSampleRate = &ServiceError{Key: "tuner.error.sampleRate"}
	// ErrDeviceLost reports a capture device that died mid-run.
	ErrDeviceLost = &ServiceError{Key: "tuner.error.deviceLost"}
	// ErrFileUnreadable reports a recording that stopped decoding mid-run.
	ErrFileUnreadable = &ServiceError{Key: "tuner.error.fileUnreadable"}
)

// emitEvent is the single door to the frontend and the seam tests replace.
var emitEvent = emit.Emit
