package config

#Config: {
	preferences: #Preferences
	logger:      #LoggerConfig
	audio:       #Audio
	tuner:       #Tuner
	report:      #Report
	storage:     #Storage
	engine:      #Engine
}

// #Engine is the audio engine's own tuning, mapped one-to-one onto
// core/dsp.EngineConfig at the services/tuner boundary. The engine depends on
// nothing in this package - these are the app's knobs on it - which is what
// keeps core/dsp liftable on its own. The defaults are the engine's own
// (core/dsp.DefaultEngineConfig); the ranges keep a typo from producing a
// window too short to resolve a reed or a hold no note could satisfy.
//
// This section is the engine's TIMINGS. The engine's deeper calibration (its
// detection, beat and pair thresholds) stays as documented constants in
// core/dsp: those are load-bearing and empirically fixed, and opening them to a
// config file would be a way to silently break reed detection.
#Engine: {
	// The analysis window, in milliseconds. The highest-leverage knob: it sets
	// the frequency resolution and how close two reeds of a musette may sit and
	// still be told apart, and the slowest beat the engine can honestly measure.
	// Three seconds is the balance the detector is calibrated around; move it far
	// and a merged pair stops looking like one.
	fine_window_ms: int & >=500 & <=8000 | *3000 @go(FineWindowMs)

	// How long a reading must hold still, in milliseconds, before the engine
	// reports a stable lock. A take is recorded on the lock; in calibration the
	// capture opens on it. Lower it and the tuner commits sooner but to a less
	// settled reading.
	lock_hold_ms: int & >=200 & <=5000 | *1250 @go(LockHoldMs)

	// How far any reed may drift, in hertz, between fine results and still count
	// as the same reading. Raise it to keep the lock through a wavering note;
	// lower it to demand a stiller one.
	lock_epsilon_hz: number & >0 & <=2 | *0.1 @go(LockEpsilonHz,type=float64)
}

// db_path defaults to a @{datadir:...} reference so one config file means the
// right thing everywhere: the platform data directory on an installed build,
// beside the config file on a portable one.
#Storage: {
	db_path: string & != "" | *"@{datadir:smeggtuner.db}" @go(DBPath)
}

#ThemeMode:  "auto" | "light" | "dark" | "lightHighContrast" | "darkHighContrast"
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
	clock_ppm:     number | *0.0 @go(ClockPPM,type=float64)
}

// tolerance is the reference technician's own figure: "najlepiej glosy czyste by
// byly z dokladnoscia do centa". beat_tolerance is deliberately looser, and for
// arithmetic rather than taste: a beat carries BOTH its reeds' errors, so a pair
// at +1 and -1 cents is two cents apart while each reed passes its own one cent
// window. See core/target.DefaultBeatTolerance.
#Tuner: {
	a4:                       number & >=430 & <=450 | *440.0               @go(A4,type=float64)
	unit:                     "cent" | "hz" | *"cent"                       @go(Unit)
	scale_naming:             "cdefgab" | "cdefgah" | "doremi" | "polish" | *"cdefgab" @go(ScaleNaming)
	error_reference:          "scale" | "goal" | *"scale"                   @go(ErrorReference)
	stop_after_lock:          bool | *false                                 @go(StopAfterLock)
	continuous_update_manual: bool | *false                                 @go(ContinuousUpdateManual)
	tolerance:                number & >0 & <=50 | *1.0                     @go(Tolerance,type=float64)
	beat_tolerance:           number & >0 & <=50 | *3.0                     @go(BeatTolerance,type=float64)

	// How long the reference tone plays, in milliseconds, when the note strip is
	// asked to sound a note. Ten seconds, per Dirk's "play the note for ten
	// seconds"; it is also how long the recorder treats the room as contaminated.
	tone_duration_ms: int & >=1000 & <=30000 | *10000 @go(ToneDurationMs)

	// The calibration capture window, in milliseconds: how long the Capture key
	// stays live after the engine locks onto a note while calibrating an
	// instrument, so the reading can be taken without racing a lock that comes
	// and goes. Longer than a lock hold on purpose - calibration is a deliberate,
	// one-note-at-a-time job.
	calibration_capture_ms: int & >=500 & <=8000 | *2000 @go(CalibrationCaptureMs)

	// note_sounds lets the note strip speak: hold a note and hear it, which is how
	// a technician finds the key he has pinned on an instrument whose layout he
	// does not know by heart (Dirk's manual, chapter 15).
	//
	// It defaults OFF, and that is not timidity. This tuner listens through a
	// microphone, so a tone out of the speakers is a tone the microphone hears.
	// The engine cannot tell it from the instrument, and a take recorded while it
	// sounds would put a reed on a printed report that does not exist. The RECORDER
	// is what refuses that (services/tuner.observe), but a feature that can quietly
	// contaminate a measurement should be reached for, not stumbled into.
	note_sounds: bool | *false @go(NoteSounds)
}

#Report: {
	company_name:    string | *"" @go(CompanyName)
	company_address: string | *"" @go(CompanyAddress)
	company_website: string | *"" @go(CompanyWebsite)
	logo_path:       string | *"" @go(LogoPath)
}

#LoggerConfig: {
	verbose:      bool | *false                      @go(Verbose)
	no_color:     bool | *false                      @go(NoColor)
	enable_files: bool | *false                      @go(EnableFiles)
	dir:          string | *"./logs"                 @go(Dir)
	prefix:       string | *""                       @go(Prefix)
	level:        string | *"INFO"                   @go(Level)
	max_size_mb:  int & >0 | *10                     @go(MaxSizeMb)
	max_backups:  int & >=0 | *5                     @go(MaxBackups)
	max_age_days: int & >0 | *30                     @go(MaxAgeDays)
	log_name:     string | *"smeggtuner.log"         @go(LogName)
	compression:  "none" | "gzip" | "zstd" | *"zstd" @go(Compression)
	local_time:   bool | *true                       @go(LocalTime)
}
