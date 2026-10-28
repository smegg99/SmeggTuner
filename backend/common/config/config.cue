package config

#Config: {
	preferences: #Preferences
	logger:      #LoggerConfig
	audio:       #Audio
	tuner:       #Tuner
	report:      #Report
	storage:     #Storage
}

// db_path defaults to a @{datadir:...} reference so one config file means the
// right thing everywhere: the platform data directory on an installed build,
// beside the config file on a portable one.
#Storage: {
	db_path: string & !="" | *"@{datadir:smeggtuner.db}" @go(DBPath)
}

#ThemeMode:  "auto" | "light" | "dark"
#AccentMode: "auto" | "custom"
#Language:   "en" | "pl"

#Preferences: {
	theme: #ThemeMode | *"auto" @go(Theme)
	// accent_mode "auto" follows the desktop's own accent (via s99wails), falling
	// back to the app's built-in blue where the OS exposes none; "custom" uses
	// accent_color instead.
	accent_mode:   #AccentMode | *"auto" @go(AccentMode)
	accent_color:  string | *"#2563eb"   @go(AccentColor)
	language:      #Language | *"en"     @go(Language)
	close_to_tray: bool | *true          @go(CloseToTray)
}

#Audio: {
	device_id:     string | *""  @go(DeviceID)
	hum_filter_50: bool | *false @go(HumFilter50)
	hum_filter_60: bool | *false @go(HumFilter60)

	// devices is what is kept per capture device, keyed by its device id; an unlisted device has NewDevice's values.
	devices: [string]: #Device @go(Devices)
}

#Device: {
	// calibration is the sound-card clock calibration's outcome, shown to the user.
	calibration: "uncalibrated" | "calibrated" | "unreliable" | *"uncalibrated" @go(Calibration)

	// calibration_percent corrects the sample clock, rate = nominal*(1+p/100), rounded with meter.StoredCalibration;
	// -999999 (meter.CalibrationMissing) is no calibration: no correction, and no reference tone.
	calibration_percent: number & (>=-500 & <=500 | >=-999999 & <=-999999) | *-999999 @go(CalibrationPercent,type=float64)

	// auto_volume lets the tuner steer the device's OS input level while it captures; the OS level comes back at stop.
	auto_volume: bool | *true @go(AutoVolume)

	// volume is the OS input level, 0..1, automatic volume last left the device at; it starts there next time.
	volume: number & >=0 & <=1 | *0.5 @go(Volume,type=float64)
}

#Tuner: {
	a4:              number & >=430 & <=450 | *440.0                          @go(A4,type=float64)
	scale_naming:    "cdefgab" | "cdefgah" | "doremi" | "polish" | *"cdefgab" @go(ScaleNaming)
	error_reference: "scale" | "goal" | *"scale"                              @go(ErrorReference)

	// deviation_unit is display only: the engine, storage and goals carry reference units (0.06 percent of the note frequency).
	deviation_unit: "reference" | "cents" | *"reference" @go(DeviationUnit)

	// beat_unit maps onto the engine's BeatsInHertz; "reference" reads a beat in deviation units.
	beat_unit: "hz" | "reference" | *"hz" @go(BeatUnit)

	// detection trades how soon a new note reads and locks against how sure the engine is before it locks
	// (EngineConfig.Detection); "steady" is the reference engine with nothing added.
	detection: "fast" | "balanced" | "steady" | *"balanced" @go(Detection)

	// stop_after_lock gates the engine: an acquired result is held and fine analysis skipped (EngineConfig.StopAfterLock).
	stop_after_lock:          bool | *false @go(StopAfterLock)
	continuous_update_manual: bool | *false @go(ContinuousUpdateManual)

	// tolerance and beat_tolerance are in deviation units; a beat carries both reeds' errors, so its window is wider.
	tolerance:      number & >0 & <=50 | *1.0 @go(Tolerance,type=float64)
	beat_tolerance: number & >0 & <=50 | *3.0 @go(BeatTolerance,type=float64)

	// lock_a4 makes the A4 control refuse every change.
	lock_a4: bool | *false @go(LockA4)

	// transpose is how many semitones (0..11) the instrument's keys sound above their names. Tones mode only.
	transpose: int & >=0 & <=11 | *0 @go(Transpose,type=int)

	// mode, reed_boxes and octave_offsets decide what is measured while no session is open; a session's register decides otherwise.
	mode: "tones" | "octaves" | "chords" | *"tones" @go(Mode)

	// reed_boxes are the reeds tones mode reads, by box 1..3; never empty. The list builtins are avoided:
	// CUE v0.17 panics unifying a value equal to a list default guarded by them.
	reed_boxes: [int & >=1 & <=3, ...int & >=1 & <=3] | *[1] @go(ReedBoxes,type=[]int)

	// octave_offsets are the octaves above the base that octaves mode reads, in semitones; never empty.
	octave_offsets: [12 | 24, ...12 | 24] | *[12] @go(OctaveOffsets,type=[]int)

	// How long the reference tone sounds, in milliseconds, its 0.5 s fades included. The
	// reference generator sounds 43 s; this is also how long the recorder treats the room as contaminated.
	tone_duration_ms: int & >=1000 & <=60000 | *43000 @go(ToneDurationMs)

	// The calibration capture window, in milliseconds: how long the Capture key
	// stays live after the engine locks onto a note while calibrating an
	// instrument, so the reading can be taken without racing a lock that comes
	// and goes.
	calibration_capture_ms: int & >=500 & <=8000 | *2000 @go(CalibrationCaptureMs)
}

#Report: {
	company_name:    string | *"" @go(CompanyName)
	company_address: string | *"" @go(CompanyAddress)
	company_website: string | *"" @go(CompanyWebsite)
}

#LoggerConfig: {
	verbose:      bool | *false                      @go(Verbose)
	no_color:     bool | *false                      @go(NoColor)
	enable_files: bool | *true                       @go(EnableFiles)
	dir:          string | *"@{datadir:logs}"        @go(Dir)
	prefix:       string | *""                       @go(Prefix)
	level:        string | *"INFO"                   @go(Level)
	max_size_mb:  int & >0 | *2                      @go(MaxSizeMb)
	max_backups:  int & >=0 | *3                     @go(MaxBackups)
	max_age_days: int & >0 | *14                     @go(MaxAgeDays)
	log_name:     string | *"smeggtuner.log"         @go(LogName)
	compression:  "none" | "gzip" | "zstd" | *"zstd" @go(Compression)
	local_time:   bool | *true                       @go(LocalTime)
}
