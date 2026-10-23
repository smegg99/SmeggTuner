package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// The instrument sits near its own A4, so every note must land within this of its scale pitch.
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

// TestSamples checks a folder against the German notes in its file names: SMEGGTUNER_SAMPLES=<dir> SMEGGTUNER_SAMPLES_A4=442.
func TestSamples(t *testing.T) {
	t.Parallel()
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
	a4 := 440.0
	if raw := os.Getenv("SMEGGTUNER_SAMPLES_A4"); raw != "" {
		if a4, err = strconv.ParseFloat(raw, 64); err != nil {
			t.Fatalf("SMEGGTUNER_SAMPLES_A4=%q: %v", raw, err)
		}
	}
	found := map[int]map[int]tuning.Note{}
	for _, path := range paths {
		s, ok := parseSampleName(path)
		if !ok {
			t.Logf("skipping %s: no note in the file name", filepath.Base(path))
			continue
		}
		t.Run(s.name, func(t *testing.T) {
			all := runFile(t, s.path, dsp.EngineConfig{A4: a4, ReedCount: 1}).all
			note, share := heard(all)
			if note == 0 {
				t.Fatal("nothing measured")
			}
			if int(note)%12 != s.class || share < minShare {
				t.Errorf("reads %s, the file says %s", summary(all), s.name)
			}
			freq, _ := sustained(sounding(all, note), 0)
			if c := cents(freq, note.Freq(a4)); c > sampleMaxCents || c < -sampleMaxCents {
				t.Errorf("%s at %.2f Hz is %+.1f cents off its scale pitch: a wrong note, not a tuning error", name(note), freq, c)
			}
			if ts := takes(all); len(ts) == 0 {
				t.Logf("%-14s %-4s %8.2f Hz %+6.1f cents, never locked", s.name, name(note), freq, cents(freq, note.Freq(a4)))
			} else {
				t.Logf("%-14s %-4s %8.2f Hz %+6.1f cents, take %.2f Hz", s.name, name(note), freq, cents(freq, note.Freq(a4)), ts[0].Reeds[0].Freq)
			}
			if found[s.class] == nil {
				found[s.class] = map[int]tuning.Note{}
			}
			found[s.class][s.register] = note
		})
	}
	for class, by := range found {
		eight, e := by[8]
		sixteen, s := by[16]
		if e && s && eight-sixteen != 12 {
			t.Errorf("pitch class %d: 8' is %s and 16' is %s, %d semitones apart, not 12", class, name(eight), name(sixteen), eight-sixteen)
		}
	}
}
