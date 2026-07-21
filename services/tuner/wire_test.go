package tuner

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"smegg.me/smeggtuner/core/dsp"
)

// These tests pin the three pictures to bytes on the wire; if one reverts to a float
// array the app gets slow quietly (Wails compiles each event as JS ~12/s), so it fails here.

// sampleMeasurement is a full (non-heartbeat) measurement, as the engine emits.
func sampleMeasurement() dsp.Measurement {
	m := dsp.Measurement{
		Note:       69,
		NoteName:   "A4",
		Locked:     true,
		ScalePitch: 442,
		Reeds:      []dsp.ReedMeasure{{Freq: 442.3}},
		Spectrum:   make([]float32, dsp.SpectrumColumns),
		Waveform:   make([]float32, dsp.WaveformPoints),
		Equalizer:  make([]float32, 105),
	}
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
	b, err := json.Marshal(decorate(sampleMeasurement(), goal{}))
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

// Size is the bottleneck: 6.1 kB was 91% pictures, each byte compiled by WebKit ~12/s.
func TestMeasurementStaysSmall(t *testing.T) {
	b, err := json.Marshal(decorate(sampleMeasurement(), goal{}))
	if err != nil {
		t.Fatal(err)
	}

	// Well under the ~6.1 kB it was, with room for reed and beat rows.
	const budget = 2000
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

// A heartbeat carries no picture (a spectrum needs ~0.5s of audio); it must not become zeroes on the wire.
func TestHeartbeatCarriesNoPicture(t *testing.T) {
	dto := decorate(dsp.Measurement{Note: 69, NoteName: "A4"}, goal{})

	b, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"spectrum":""`) {
		t.Errorf("an empty spectrum should send as \"\", got:\n%s", b)
	}
	if len(b) > 300 {
		t.Errorf("a heartbeat is %d bytes; it carries no picture and should be tiny", len(b))
	}
	t.Logf("heartbeat on the wire: %d bytes", len(b))
}
