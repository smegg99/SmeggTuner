package dsp

import "smegg.me/smeggtuner/core/tuning"

type EngineState string

const (
	StateInitializing EngineState = "initializing"
	StateRunning      EngineState = "running"
	StateFrozen       EngineState = "frozen"
	StateTooLoud      EngineState = "tooLoud"
	StateTooQuiet     EngineState = "tooQuiet"
	StateDeviceLost   EngineState = "deviceLost"
)

type ReedMeasure struct {
	Freq     float64 `json:"freq"`
	DevCents float64 `json:"devCents"` // vs the note's scale pitch at engine A4
}

type BeatMeasure struct {
	Pair  string  `json:"pair"` // "1-2", "2-3"
	Hz    float64 `json:"hz"`
	Cents float64 `json:"cents"` // beat expressed in cents at the note frequency
	// FromEnvelope means the reeds sat too close to separate and the beat was read off the amplitude.
	// Depth is the amplitude swing as a fraction of the carrier. Zero for a beat taken from peaks.
	FromEnvelope bool    `json:"fromEnvelope"`
	Depth        float64 `json:"depth"`
}

// SpectrumColumns and SpectrumCents size Measurement.Spectrum: 256 columns across +-50 cents, the
// same range the error ruler shows, so a reed sits at the same place in both.
const (
	SpectrumColumns = 256
	SpectrumCents   = 50.0
)

type Measurement struct {
	Note     tuning.Note `json:"note"`
	NoteName string      `json:"noteName"`
	Locked   bool        `json:"locked"`
	// LockProgress is how far the reading is toward a lock, 0..1: the settle for a UI to draw. The
	// lock itself is Locked; this only pictures the wait. 0 on a heartbeat.
	LockProgress float64 `json:"lockProgress"`
	// ScalePitch is the note's exact frequency at the engine's A4 and transposition: what every
	// DevCents is measured against, and where a display draws its centre line. 0 on a heartbeat.
	ScalePitch float64       `json:"scalePitch"`
	Reeds      []ReedMeasure `json:"reeds"`
	// ReedsSeparated reports whether the spectrum told the reeds apart. False means they sounded
	// closer than this window can resolve, so the peak picker's lines are not one reed each. False
	// with ReedsFromBeat also false is the case with no per-reed answer at all: show the beat, not the split.
	ReedsSeparated bool `json:"reedsSeparated"`
	// ReedsFromBeat reports that Reeds was recovered from the beat rather than the spectrum, and may
	// be trusted. Never true at the same time as ReedsSeparated: two routes to the same answer, kept
	// apart so a display can say which one it has. Below the beat floor, or where the pair could not
	// be told from three merged reeds or one reed on a moving bellows, both are false.
	ReedsFromBeat bool          `json:"reedsFromBeat"`
	Beats         []BeatMeasure `json:"beats"`
	Equalizer     []float32     `json:"equalizer"` // one dB-ish value per note, for the UI

	// SourceAt is where in the recording this reading's audio came from, in seconds: the end of the
	// window it was measured over. Zero for a microphone. The reading is latched, so the mark carries
	// its own provenance rather than being guessed from the playhead.
	SourceAt float64 `json:"sourceAt"`
	// Spectrum is the analysis band as the display draws it: SpectrumColumns heights in 0..1 across
	// +-SpectrumCents around ScalePitch, in dB under the loudest line. Empty on a heartbeat. A
	// picture, not a measurement: read values from Reeds and Beats, not from here.
	Spectrum   []float32 `json:"spectrum"`
	InputLevel float32   `json:"inputLevel"`
	// Waveform is the most recent block thinned to WaveformPoints signed peaks for the input strip. It
	// ships with every measurement so the trace keeps moving. Each point is the largest-magnitude
	// sample in its stride, sign and all, taken after the hum notches and highpass. Scaled so the
	// loudest point reaches full deflection, except below waveFloor where the scale is held.
	Waveform []float32   `json:"waveform"`
	State    EngineState `json:"state"`
}

// SpectrumFloorDB is the bottom of Measurement.Spectrum, in dB under the loudest line: a column's
// height is 1 + dB/SpectrumFloorDB. Exported because the display needs to label its axis.
const SpectrumFloorDB = 60.0

// WaveformPoints is how many points Measurement.Waveform carries.
const WaveformPoints = 256

// EqualizerCeilingDB is the top of Measurement.Equalizer: equalizerDB clamps every band to
// 0..EqualizerCeilingDB over its noise floor. Exported because the wire format needs the range.
const EqualizerCeilingDB = 60.0

// EqualizerFullScaleDB is how tall a band can actually get, and it is NOT EqualizerCeilingDB. Draw
// against this one: equalizerDB divides by a reference never below leakGuardRatio of the loudest note
// in the frame, so the loudest band is 20*log10(1/leakGuardRatio) and no band exceeds it.
// TestEqualizerFullScaleIsSetByTheLeakGuard keeps it honest if leakGuardRatio moves.
const EqualizerFullScaleDB = 27.9588
