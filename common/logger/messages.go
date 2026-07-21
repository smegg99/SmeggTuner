// Catalog of log event identifiers; each value is the human-readable message.

package logger

const (
	// Application lifecycle
	MsgAppStarting    MessageID = "starting smeggtuner"
	MsgAppInitialized MessageID = "application initialized"
	MsgCleaningUp     MessageID = "cleaning up"

	// Environment
	MsgAppImageEnv MessageID = "running in AppImage environment, applying webkit workaround"
	MsgX11Fallback MessageID = "Plasma Wayland session detected, defaulting GDK_BACKEND to x11"
	MsgGPUPath     MessageID = "SMEGGTUNER_GPU=1: keeping the DMABUF renderer and staying on native Wayland"

	// Configuration
	MsgLoadingConfig    MessageID = "loading configuration"
	MsgConfigLoaded     MessageID = "configuration loaded"
	MsgConfigLoadFailed MessageID = "failed to load configuration"
	MsgDefaultConfigGen MessageID = "default configuration generated"
	MsgReinitLogger     MessageID = "reconfiguring logger from configuration"
	MsgLoggerInitFailed MessageID = "failed to configure logger"

	// Datastore
	MsgDatastoreReady      MessageID = "datastore ready"
	MsgDatastoreInitFailed MessageID = "failed to initialize datastore"
	MsgLegacyImported      MessageID = "imported pre-database data"
	MsgLegacyImportFailed  MessageID = "failed to import pre-database data"
	MsgConfigSaveFailed    MessageID = "failed to save configuration"
	MsgPreferencesSaved    MessageID = "preferences saved"

	// Audio input selection
	MsgAudioDevicesListed    MessageID = "capture devices enumerated"
	MsgAudioDeviceEnumFailed MessageID = "failed to enumerate capture devices"
	MsgAudioNoDevices        MessageID = "no capture devices available"
	MsgAudioDeviceGone       MessageID = "selected capture device is no longer present"
	MsgAudioSourceSelected   MessageID = "input source selected"
	MsgAudioFileUnreadable   MessageID = "recording could not be opened"
	// warning, not error: measurement still works, only playback is lost.
	MsgAudioSpeakerUnavailable MessageID = "no output device, file playback will be silent"
	MsgAudioDialogUnavailable  MessageID = "file dialog requested without a running application"
	MsgAudioDialogFailed       MessageID = "file dialog failed"

	// Tuner engine lifecycle
	MsgTunerStarted         MessageID = "measurement engine started"
	MsgTunerStopped         MessageID = "measurement engine stopped"
	MsgTunerStartFailed     MessageID = "failed to build the input source"
	MsgTunerRunFailed       MessageID = "measurement engine stopped with an error"
	MsgTunerEventsDropped   MessageID = "measurement events dropped, the frontend is behind"
	MsgTunerSettingRejected MessageID = "rejected an out-of-range setting"
	MsgTunerTonePlaying     MessageID = "playing reference tone"
	MsgTunerTonePlayFailed  MessageID = "reference tone playback failed"

	// Tuning sessions
	MsgSessionCreated       MessageID = "session created"
	MsgSessionOpened        MessageID = "session opened"
	MsgSessionClosed        MessageID = "session closed"
	MsgSessionDeleted       MessageID = "session deleted"
	MsgSessionSaveFailed    MessageID = "failed to save the session"
	MsgSessionLoadFailed    MessageID = "failed to load the session"
	MsgSessionRejected      MessageID = "rejected an invalid session change"
	MsgSessionReedsClamped  MessageID = "instrument sounds more reeds than the engine resolves, clamping the engine"
	MsgSessionCurveFitted   MessageID = "goal curve fitted from a pass"
	MsgSessionCurveImported MessageID = "goal curve imported"
	MsgSessionCurveDropped  MessageID = "goal curve dropped"

	// Record mode
	MsgRecordTake     MessageID = "take recorded"
	MsgRecordUndone   MessageID = "last take undone"
	MsgRecordCleared  MessageID = "pass cleared"
	MsgRecordEdited   MessageID = "take edited by hand"
	MsgRecordRejected MessageID = "rejected a record request"

	// Tuning report
	MsgReportWritten    MessageID = "tuning report written"
	MsgReportOpened     MessageID = "tuning report handed to the browser"
	MsgReportFailed     MessageID = "failed to write the tuning report"
	MsgReportOpenFailed MessageID = "failed to open the tuning report"
	MsgReportRejected   MessageID = "rejected a report request"

	// Window / tray
	MsgSecondInstance        MessageID = "second instance launched, focusing existing window"
	MsgTrayIconFailed        MessageID = "failed to load tray icon"
	MsgWindowStateLoadFailed MessageID = "failed to load window state"
	MsgWindowStateSaveFailed MessageID = "failed to save window state"
	MsgRunFailed             MessageID = "application run failed"
)
