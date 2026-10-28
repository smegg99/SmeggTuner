// Command analyze runs the measurement engine over a WAV or synth spec and prints its last locked measurement as JSON, or the last measured one when nothing locked.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/tuning"
)

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func parseNote(name string) (tuning.Note, error) {
	if name == "" {
		return 0, nil
	}
	for n := tuning.MinNote; n <= tuning.MaxNote; n++ {
		if n.Name(tuning.NamingCDEFGAB) == name {
			return n, nil
		}
	}
	return 0, fmt.Errorf("unknown note %q (use names like A4, C#5)", name)
}

func main() {
	wavPath := flag.String("wav", "", "input WAV file")
	synthPath := flag.String("synth", "", "SynthSpec JSON file")
	a4 := flag.Float64("a4", 440, "A4 reference frequency")
	noteName := flag.String("note", "", "manual note (e.g. A4); empty = auto")
	reeds := flag.Int("reeds", 1, "reed count; -banks and -bass set it themselves")
	banks := flag.String("banks", "", "register banks (e.g. LM, LMMM, MH); sets the compound layout and the reed count")
	bass := flag.Int("bass", 0, "bass machine voice count (2..6): tune the whole octave ladder")
	calibration := flag.Float64("calibration", 0, "sample-clock calibration, percent")
	trace := flag.Bool("trace", false, "print every measurement to stderr")
	flag.Parse()

	if (*wavPath == "") == (*synthPath == "") {
		fail("exactly one of -wav or -synth is required")
	}

	var octaves []dsp.OctaveRequest
	if *banks != "" {
		bs, err := session.ParseBanks(*banks)
		if err != nil {
			fail("%v", err)
		}
		octaves = session.OctavesOf(bs)
		*reeds = len(bs)
	}
	if *bass != 0 {
		feet := session.BassFeet(*bass)
		if feet == nil {
			fail("-bass must be %d..%d", session.MinBassReeds, session.MaxBassReeds)
		}
		octaves = session.OctavesOfFeet(feet)
		*reeds = len(feet)
	}
	// Checked after -banks and -bass, which overwrite it: what the engine runs with is what has to
	// be in range, and a four-rank register (LMMH) or a six-rank bass ladder is a normal instrument.
	if *reeds < session.MinReeds || *reeds > session.MaxReeds {
		fail("-reeds must be %d..%d", session.MinReeds, session.MaxReeds)
	}

	var src audio.Source
	if *wavPath != "" {
		s, err := audio.NewFileSource(*wavPath, false, false)
		if err != nil {
			fail("open wav: %v", err)
		}
		src = s
	} else {
		raw, err := os.ReadFile(*synthPath)
		if err != nil {
			fail("read synth spec: %v", err)
		}
		var spec audio.SynthSpec
		if err := json.Unmarshal(raw, &spec); err != nil {
			fail("parse synth spec: %v", err)
		}
		src = audio.NewSynthSource(spec)
	}

	manual, err := parseNote(*noteName)
	if err != nil {
		fail("%v", err)
	}

	var final, locked dsp.Measurement
	emit := func(m dsp.Measurement) {
		// Heartbeats carry no slot values; retain only summary candidates.
		if m.Frame == dsp.FrameMeasured {
			final = m
			if m.Lock.Visible {
				locked = m
			}
		}
		if *trace {
			line, _ := json.Marshal(m)
			fmt.Fprintln(os.Stderr, string(line))
		}
	}

	// No LearnOnStart: recordings start mid-note, and learning would fold the tone into the background.
	eng := dsp.NewEngine(dsp.EngineConfig{
		A4:          *a4,
		ReedCount:   *reeds,
		ManualNote:  manual,
		Calibration: *calibration,
		Octaves:     octaves,
		Bass:        *bass != 0,
	}, emit)

	if err := eng.Run(context.Background(), src); err != nil {
		fail("engine: %v", err)
	}

	if locked.Frame == dsp.FrameMeasured {
		final = locked
	}
	if final.Frame != dsp.FrameMeasured {
		fail("no measured frame (empty or silent input)")
	}
	out, err := json.MarshalIndent(final, "", "  ")
	if err != nil {
		fail("marshal: %v", err)
	}
	fmt.Println(string(out))
}
