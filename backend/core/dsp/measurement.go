package dsp

import "smegg.me/smeggtuner/backend/core/tuning"

type EngineState string

const (
	StateInitializing EngineState = "initializing"
	StateRunning      EngineState = "running"
	StateFrozen       EngineState = "frozen"
	StateTooLoud      EngineState = "tooLoud" // the last level interval had a flat top
	StateTooQuiet     EngineState = "tooQuiet"
	StateDeviceLost   EngineState = "deviceLost"
)

// FrameKind says what an analysis event did; FrameNone is a heartbeat between analyses.
type FrameKind string

const (
	FrameNone       FrameKind = ""
	FrameWaiting    FrameKind = "waiting"    // not enough audio for a window yet
	FrameNoNote     FrameKind = "noNote"     // no note selected: readouts dim, numbers stay
	FrameHeld       FrameKind = "held"       // a result was acquired and StopAfterLock holds it
	FrameIneligible FrameKind = "ineligible" // a reed beyond the display range: readouts unchanged
	FrameMeasured   FrameKind = "measured"
)

// ReedMeasure is one reed slot of the analysis just made; a slot without a peak is not Present.
type ReedMeasure struct {
	Reed    int     `json:"reed,omitempty"` // the reference reed box, 1-based; the slot number in the extensions
	Present bool    `json:"present"`
	Freq    float64 `json:"freq"`
	Dev     float64 `json:"dev"`              // deviation units: 1 = 0.06 percent of the slot's note
	Octave  int     `json:"octave,omitempty"` // semitones from the key: -12 a 16', +12 a 4'
}

// BeatMeasure is one adjacent slot pair of the analysis just made, in Hz.
type BeatMeasure struct {
	Pair    string  `json:"pair"` // 1-based positions in Reeds, such as "1-2"
	Present bool    `json:"present"`
	Hz      float64 `json:"hz"`
}

// Panel is one numeric readout as the reference display holds it, in order reed, pair, reed, ...
type Panel struct {
	Value    float64 `json:"value"`
	Target   float64 `json:"target"`
	Signed   bool    `json:"signed"`
	Decimals int     `json:"decimals"`
	Dimmed   bool    `json:"dimmed"`
	Text     string  `json:"text,omitempty"`   // note annotation (compound modes)
	Shown    string  `json:"shown"`            // the value as the reference prints it, "-" when missing
	Absent   bool    `json:"absent,omitempty"` // no such readout: reeds 1 and 3, chord tones, Bass octaves in units
}

// Needle is one slot's retained frequency cache; it keeps its numbers when it turns invalid or dims.
type Needle struct {
	Dev   float64 `json:"dev"`
	Freq  float64 `json:"freq"`
	Valid bool    `json:"valid"`
}

// Lock keeps the reference's latches apart: the visible lock can outlast the internal one.
type Lock struct {
	Visible   bool `json:"visible"`
	Internal  bool `json:"internal"`
	Acquired  bool `json:"acquired"`
	Countdown int  `json:"countdown"`
	Recorded  bool `json:"recorded"` // a take was written; shown in place of the lock until the label changes
}

// InputInterval is the last completed 0.5 s level interval of the int16 capture, for the automatic volume.
type InputInterval struct {
	Seq     uint64  `json:"seq"`     // counts completed intervals in this run; 0 before the first
	RMS     float32 `json:"rms"`     // int16 units
	Peak    float32 `json:"peak"`    // largest |sample|, int16 units
	Clipped bool    `json:"clipped"` // a flat top: 8 equal consecutive samples at |x| >= 500
	Restart bool    `json:"restart"` // the capture ring restarted since the previous interval: run start, seek, note-off
}

type Measurement struct {
	State         EngineState   `json:"state"`
	Frame         FrameKind     `json:"frame"`
	Note          tuning.Note   `json:"note"`
	NoteName      string        `json:"noteName"`
	ScalePitch    float64       `json:"scalePitch"`      // the key's sounding reference frequency; 0 without a note
	Label         string        `json:"label,omitempty"` // the chord name
	Taken         bool          `json:"taken,omitempty"` // this analysis wrote a take while recording
	Reeds         []ReedMeasure `json:"reeds"`
	Beats         []BeatMeasure `json:"beats"`
	Panels        []Panel       `json:"panels"`
	Needles       []Needle      `json:"needles"`
	NeedlesDimmed bool          `json:"needlesDimmed"`
	Lock          Lock          `json:"lock"`
	Learning      bool          `json:"learning"`
	Equalizer     []float32     `json:"equalizer"`
	Spectrum      []float32     `json:"spectrum"`
	Spectra       [][]float32   `json:"spectra,omitempty"` // one picture per band or chord tone, on one scale
	InputLevel    float32       `json:"inputLevel"`
	Input         InputInterval `json:"input"`
	Unreliable    bool          `json:"unreliable"` // the capture lost audio within the last 0.1 s; display only
	Waveform      []float32     `json:"waveform"`
	SourceAt      float64       `json:"sourceAt"`
}
