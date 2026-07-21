// Command analyze runs the measurement engine over a WAV or synth spec and prints one measurement as JSON; -last reports the raw last tick instead of aggregating.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
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
	reeds := flag.Int("reeds", 1, "reed count 1..3")
	ppm := flag.Float64("ppm", 0, "clock correction ppm")
	window := flag.Float64("window", 3.0, "fine analysis window, seconds")
	trace := flag.Bool("trace", false, "print every measurement to stderr")
	last := flag.Bool("last", false, "report the last (locked if any) tick instead of aggregating")
	flag.Parse()

	if (*wavPath == "") == (*synthPath == "") {
		fail("exactly one of -wav or -synth is required")
	}
	if *reeds < 1 || *reeds > 3 {
		fail("-reeds must be 1..3")
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

	var all []dsp.Measurement
	emit := func(m dsp.Measurement) {
		all = append(all, m)
		if *trace {
			line, _ := json.Marshal(m)
			fmt.Fprintln(os.Stderr, string(line))
		}
	}

	eng := dsp.NewEngine(dsp.EngineConfig{
		A4:         *a4,
		ReedCount:  *reeds,
		ManualNote: manual,
		ClockPPM:   *ppm,
		// CalibSecs 0: recordings start mid-note; calibrating would fold the tone into the noise floor.
		FineWindow: time.Duration(*window * float64(time.Second)),
	}, emit)

	if err := eng.Run(context.Background(), src); err != nil {
		fail("engine: %v", err)
	}

	var final dsp.Measurement
	if *last {
		// Skip the reedless heartbeats between results, or the fallback reports no reeds.
		for _, m := range all {
			if m.Locked {
				final = m
			}
		}
		if !final.Locked {
			for _, m := range all {
				if len(m.Reeds) > 0 {
					final = m
				}
			}
		}
	} else {
		agg, ok := dsp.Aggregate(all, *a4, 0.5)
		if !ok {
			fail("no usable measurements (empty or silent input)")
		}
		final = agg
	}
	out, err := json.MarshalIndent(final, "", "  ")
	if err != nil {
		fail("marshal: %v", err)
	}
	fmt.Println(string(out))
}
