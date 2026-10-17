package tuner

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
)

// These tests pin the three pictures to bytes on the wire; if one reverts to a float
// array the app gets slow quietly (Wails compiles each event as JS ~12/s), so it fails here.

// screen is the reference display state every measurement carries, sized for the widest register the engine resolves.
func screen(m *dsp.Measurement) {
	for i := range 2*maxReeds - 1 {
		m.Panels = append(m.Panels, dsp.Panel{Value: -12.3, Target: -99, Signed: true, Decimals: 1, Dimmed: i%2 == 1})
	}
	for range maxReeds {
		m.Needles = append(m.Needles, dsp.Needle{Dev: -12.34567, Freq: 442.123456, Valid: true})
	}
	m.Lock = dsp.Lock{Visible: true, Internal: true, Acquired: true, Countdown: 3}
}

// sampleMeasurement is a measured analysis of the widest register, as the engine emits it.
func sampleMeasurement() dsp.Measurement {
	m := measured(69, -12.345678, -0.123456, 4.567891, 9.876543, 13.24680, 17.97531)
	m.NoteName = "A4"
	m.Spectrum = make([]float32, dsp.SpectrumColumns)
	m.Waveform = make([]float32, dsp.WaveformPoints)
	m.Equalizer = make([]float32, 105)
	screen(&m)
	for i := range m.Spectrum {
		m.Spectrum[i] = float32(i) / float32(len(m.Spectrum)-1) // 0..1
	}
	for i := range m.Waveform {
		m.Waveform[i] = float32(math.Sin(2 * math.Pi * float64(i) / 32)) // -1..1
	}
	for i := range m.Equalizer {
		m.Equalizer[i] = float32(i) / float32(len(m.Equalizer)-1) * dsp.EqualizerCeilingDB
	}
	return m
}

// The pictures must reach the frontend as base64, not float arrays; the packed fields
// shadow the embedded []float32 by being one level shallower in MeasurementDTO.
func TestPicturesGoOverAsBytes(t *testing.T) {
	b, err := json.Marshal(decorate(sampleMeasurement(), goal{a4: 440}))
	if err != nil {
		t.Fatal(err)
	}
	wire := string(b)

	var got struct {
		Spectrum  string `json:"spectrum"`
		Waveform  string `json:"waveform"`
		Equalizer string `json:"equalizer"`
	}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("the pictures are not strings any more: %v\n%s", err, wire)
	}

	for name, field := range map[string]string{
		"spectrum":  got.Spectrum,
		"waveform":  got.Waveform,
		"equalizer": got.Equalizer,
	} {
		raw, err := base64.StdEncoding.DecodeString(field)
		if err != nil {
			t.Fatalf("%s is not base64: %v", name, err)
		}
		if len(raw) == 0 {
			t.Fatalf("%s came over empty", name)
		}
	}

	// A float array on the wire looks like "spectrum":[0.161...; this is how a broken shadow shows.
	if strings.Contains(wire, `"spectrum":[`) || strings.Contains(wire, `"waveform":[`) {
		t.Fatalf("a picture is being sent as a float array again:\n%s", wire)
	}
}

// Size is the bottleneck: 6.1 kB was 91% pictures, each byte compiled by WebKit on every event.
func TestMeasurementStaysSmall(t *testing.T) {
	b, err := json.Marshal(decorate(sampleMeasurement(), goal{a4: 440}))
	if err != nil {
		t.Fatal(err)
	}

	// A six-slot analysis with its rows and the screen state measured 4627 bytes (2026-10-24); twice a second.
	const budget = 5000
	if len(b) > budget {
		t.Errorf("measurement is %d bytes, budget is %d - the transport is the "+
			"bottleneck, so this is a performance regression", len(b), budget)
	}
	t.Logf("measurement on the wire: %d bytes", len(b))
}

// Eight bits must be enough to draw with: decoded values held against the measured floats.
func TestQuantizationSurvivesTheRoundTrip(t *testing.T) {
	m := sampleMeasurement()

	t.Run("spectrum keeps its heights", func(t *testing.T) {
		for i, b := range pack(m.Spectrum, 0, 1) {
			got := float32(b) / 255
			if diff := math.Abs(float64(got - m.Spectrum[i])); diff > 1.0/255 {
				t.Fatalf("column %d: %v -> %v (off by %v)", i, m.Spectrum[i], got, diff)
			}
		}
	})

	t.Run("equalizer keeps its dB", func(t *testing.T) {
		for i, b := range pack(m.Equalizer, 0, dsp.EqualizerCeilingDB) {
			got := float32(b) * dsp.EqualizerCeilingDB / 255
			if diff := math.Abs(float64(got - m.Equalizer[i])); diff > dsp.EqualizerCeilingDB/255 {
				t.Fatalf("band %d: %v dB -> %v dB (off by %v)", i, m.Equalizer[i], got, diff)
			}
		}
	})

	t.Run("waveform keeps its shape and its sign", func(t *testing.T) {
		for i, b := range packSigned(m.Waveform) {
			got := (float32(b) - 128) / 127
			if diff := math.Abs(float64(got - m.Waveform[i])); diff > 1.0/127 {
				t.Fatalf("point %d: %v -> %v (off by %v)", i, m.Waveform[i], got, diff)
			}
		}
	})
}

// Silence must decode to exactly zero at 128, or the silent trace lifts off the centre line (a DC offset).
func TestSilenceDecodesToExactlyZero(t *testing.T) {
	silent := make([]float32, dsp.WaveformPoints)

	for i, b := range packSigned(silent) {
		if b != 128 {
			t.Fatalf("point %d: silence packed to %d, want 128", i, b)
		}
		if got := (float32(b) - 128) / 127; got != 0 {
			t.Fatalf("point %d: silence decoded to %v, want exactly 0", i, got)
		}
	}

	for i, b := range pack(make([]float32, 8), 0, dsp.EqualizerCeilingDB) {
		if b != 0 {
			t.Fatalf("band %d: a silent band packed to %d, want 0", i, b)
		}
	}
}

// A heartbeat carries the screen state but no spectrum; it must not become zeroes on the wire, and at ~20/s it stays small.
func TestHeartbeatCarriesNoPicture(t *testing.T) {
	m := dsp.Measurement{State: dsp.StateRunning, Reeds: []dsp.ReedMeasure{}, Beats: []dsp.BeatMeasure{},
		Spectrum: []float32{}, Equalizer: []float32{}, Waveform: make([]float32, dsp.WaveformPoints), InputLevel: 0.2}
	screen(&m)

	b, err := json.Marshal(decorate(m, goal{}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"spectrum":""`) {
		t.Errorf("an empty spectrum should send as \"\", got:\n%s", b)
	}
	// Six slots of panels and needles plus the waveform measured 1866 bytes (2026-10-24); twenty a second.
	const budget = 2000
	if len(b) > budget {
		t.Errorf("a heartbeat is %d bytes, budget is %d", len(b), budget)
	}
	t.Logf("heartbeat on the wire: %d bytes", len(b))
}
