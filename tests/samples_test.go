// Checks real recordings against the German note name in the file ("Ais '16.wav" = A#, 16'). Run: SMEGGTUNER_SAMPLES=sounds/wav go test ./tests/ -run TestSamples -v
package tests

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

// The instrument sits near A442, so every note must land within this of its scale pitch.
const sampleMaxCents = 35.0

// German note names to pitch class: B is B flat, H is B natural.
var germanPitchClass = map[string]int{
	"C": 0, "CIS": 1, "D": 2, "DIS": 3, "E": 4, "F": 5,
	"FIS": 6, "G": 7, "GIS": 8, "A": 9, "AIS": 10, "B": 10, "H": 11,
}

// "Cis (C#) '16" -> note "CIS", register 16.
var sampleName = regexp.MustCompile(`^([A-Za-z]+)\s*(?:\([^)]*\))?\s*'\s*(\d+)`)

type sample struct {
	path     string
	name     string
	class    int // pitch class from the file name
	register int // 8 or 16
}

func parseSampleName(path string) (sample, bool) {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	m := sampleName.FindStringSubmatch(base)
	if m == nil {
		return sample{}, false
	}
	class, ok := germanPitchClass[strings.ToUpper(m[1])]
	if !ok {
		return sample{}, false
	}
	register := 8
	if m[2] == "16" {
		register = 16
	}
	return sample{path: path, name: base, class: class, register: register}, true
}

// The app shows the last tick, so every tick that carried a reed must be right, not just the average.
func liveTicks(all []dsp.Measurement) []dsp.Measurement {
	var out []dsp.Measurement
	for _, m := range all {
		if len(m.Reeds) > 0 {
			out = append(out, m)
		}
	}
	return out
}

func measureSample(t *testing.T, path string, a4 float64) (dsp.Measurement, []dsp.Measurement) {
	t.Helper()

	src, err := audio.NewFileSource(path, false, false)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	var all []dsp.Measurement
	eng := dsp.NewEngine(dsp.EngineConfig{
		A4:         a4,
		ReedCount:  1,
		FineWindow: 3 * time.Second,
	}, func(m dsp.Measurement) { all = append(all, m) })

	if err := eng.Run(context.Background(), src); err != nil {
		t.Fatalf("engine: %v", err)
	}

	// Hand-worked bellows never hold still enough to lock, so aggregate over the sustained middle.
	m, ok := dsp.Aggregate(all, a4, 0.5)
	if !ok {
		t.Fatalf("no usable measurement in %s", filepath.Base(path))
	}
	return m, liveTicks(all)
}

func TestSamples(t *testing.T) {
	dir := os.Getenv("SMEGGTUNER_SAMPLES")
	if dir == "" {
		t.Skip("set SMEGGTUNER_SAMPLES to a folder of recordings")
	}

	paths, err := filepath.Glob(filepath.Join(dir, "*.wav"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatalf("no .wav files in %s", dir)
	}

	// The reference pitch is the instrument's own; judging against A440 would just report it sharp.
	a4 := 440.0
	if raw := os.Getenv("SMEGGTUNER_SAMPLES_A4"); raw != "" {
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			t.Fatalf("SMEGGTUNER_SAMPLES_A4=%q: %v", raw, err)
		}
		a4 = parsed
	}

	// Keyed by pitch class, to check the two registers against each other.
	measured := map[int]map[int]tuning.Note{}

	for _, path := range paths {
		s, ok := parseSampleName(path)
		if !ok {
			t.Logf("skipping %s: no note in the file name", filepath.Base(path))
			continue
		}

		t.Run(s.name, func(t *testing.T) {
			m, ticks := measureSample(t, s.path, a4)

			if len(m.Reeds) == 0 {
				t.Fatalf("no reed measured")
			}

			// The screen shows the latest reading, so every reading must be right, not just the average.
			for i, tick := range ticks {
				if int(tick.Note)%12 != s.class {
					t.Errorf("tick %d of %d says %s (%.2f Hz), but the file says %s",
						i+1, len(ticks), tick.NoteName, tick.Reeds[0].Freq, s.name)
				}
			}
			gotClass := int(m.Note) % 12
			if gotClass != s.class {
				t.Errorf("note is %s, but the file says %s (heard %.2f Hz)",
					m.NoteName, s.name, m.Reeds[0].Freq)
			}

			// A wrong octave is usually a harmonic winning over the sounding reed - the failure worth catching.
			if dev := m.Reeds[0].DevCents; dev > sampleMaxCents || dev < -sampleMaxCents {
				t.Errorf("%s is %.1f cents off its scale pitch (%.2f Hz): that is not a tuning error, "+
					"it is the wrong note", m.NoteName, dev, m.Reeds[0].Freq)
			}

			t.Logf("%-14s %-4s %8.2f Hz %+6.1f cents", s.name, m.NoteName, m.Reeds[0].Freq, m.Reeds[0].DevCents)

			if measured[s.class] == nil {
				measured[s.class] = map[int]tuning.Note{}
			}
			measured[s.class][s.register] = m.Note
		})
	}

	// The 16' reed is an octave below the 8'; a harmonic winning on one shows here.
	for class, byRegister := range measured {
		eight, hasEight := byRegister[8]
		sixteen, hasSixteen := byRegister[16]
		if !hasEight || !hasSixteen {
			continue
		}
		if eight-sixteen != 12 {
			t.Errorf("pitch class %d: 8' is %s and 16' is %s, which are %d semitones apart, not 12",
				class, eight.Name(tuning.NamingCDEFGAB), sixteen.Name(tuning.NamingCDEFGAB), eight-sixteen)
		}
	}
}
