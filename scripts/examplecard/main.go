// Command examplecard writes a tuning card built from made-up readings, so the sheet can be
// looked at without an accordion, a microphone or a recorded pass. Nothing here touches the
// datastore: the session is assembled in memory and rendered by the same core/report the app
// exports through.
//
//	go run ./scripts/examplecard -out card.html
//	go run ./scripts/examplecard -shape musette -lang pl -letterhead -out karta.html
//
// Two shapes, because the sheet lays itself out differently for each. A three-voice musette fits
// one wide table; a four-voice box with a five-voice bass machine needs more columns than a page
// holds, and the sheet falls back to one narrow table per rank (see backend/core/report/layout.go).
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/report"
	"smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// The instrument the made-up pass was taken on: a four-voice box at A442 with a five-voice bass
// machine, which is the widest shape the card has to lay out.
const (
	a4    = 442.0
	reeds = 4
)

var (
	// A fixed date, so running this twice produces the same sheet.
	when = time.Date(2026, 8, 19, 14, 20, 0, 0, time.UTC)

	banks = []session.Bank{session.BankL, session.BankM1, session.BankM2, session.BankH}
)

// words is the made-up CONTENT of the card, in the language the sheet is written in - an English
// line under a Polish heading is exactly what an example must not show.
//
// It carries NO technician's note. That field is a human's own line about the instrument, and an
// example has no human: inventing one puts advice on the card that no measurement produced.
type words struct {
	date    string // how the pass's day is written into a session name
	company string
	address string
	site    string
}

var byLang = map[string]words{
	"en": {
		date:    "19 Aug 2026",
		company: "Bellows & Reed - accordion workshop",
		address: "12 Example Street, Kraków",
		site:    "example.invalid",
	},
	"pl": {
		date:    "19 sierpnia 2026",
		company: "Miech i Głos - warsztat akordeonowy",
		address: "ul. Przykładowa 12, 30-001 Kraków",
		site:    "przyklad.invalid",
	},
}

// say is the example text for the language the sheet is being written in, English when there is
// no block for it.
func say(lang string) words {
	if w, ok := byLang[lang]; ok {
		return w
	}
	return byLang["en"]
}

// text is the language's block, set once in main.
var text = byLang["en"]

func main() {
	out := flag.String("out", "example-card.html", "file to write")
	lang := flag.String("lang", "en", "language the sheet is written in")
	head := flag.Bool("letterhead", false, "print a technician's letterhead at the top")
	shape := flag.String("shape", "full", "full (LMMH + bass machine) or musette (three 8' ranks)")
	unit := flag.String("unit", "reference", "deviation unit the sheet prints: reference or cent")
	flag.Parse()

	build := example
	if *shape == "musette" {
		build = musetteExample
	}

	i18n.SetLocale(*lang)
	text = say(*lang)

	opts := report.Options{Date: when, Now: when, Naming: tuning.NamingCDEFGAB, Unit: target.Unit(*unit)}
	if *head {
		opts.Letterhead = &report.Letterhead{
			CompanyName:    text.company,
			CompanyAddress: text.address,
			CompanyWebsite: text.site,
		}
	}

	rep, err := report.Build(build(), opts)
	if err != nil {
		fail("build: %v", err)
	}

	f, err := os.Create(*out)
	if err != nil {
		fail("create: %v", err)
	}
	defer f.Close()

	if err := report.HTML(f, rep); err != nil {
		fail("render: %v", err)
	}
	fmt.Printf("wrote %s - %d treble rows, bass table: %v, %d columns, layout %q\n",
		*out, len(rep.Rows), rep.Bass != nil, rep.Columns(), rep.Layout.Class())
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// example is the made-up pass. It carries one of every row the card can print, because the point
// of the sheet is what it does with the awkward ones.
func example() *session.Session {
	s := session.New("Bugari 508/ARS - "+text.date, instrument(), a4)
	s.Curve = curve()

	// A gentle musette: the two 8' ranks straddle the key, the 16' and 4' sit on it.
	for _, t := range []session.Take{
		// Clean rows, everything inside tolerance.
		full(53, -0.4, -2.6, 2.4, 0.3),
		full(58, 0.2, -2.4, 2.7, -0.2),
		full(60, -0.3, -2.5, 2.6, 0.4),
		full(65, 0.5, -2.7, 2.3, -0.4),

		// One reed out: the 4' is 6 units sharp, which is what a card is for.
		full(69, -0.2, -2.4, 2.5, 6.1),

		// A note played on the two 8' ranks alone: two reeds, and the 16'/4' columns stay empty
		// because they were not sounding, not because they failed.
		middles(62, -2.5, 2.6),
		middles(67, -2.9, 3.4),

		// The 16' alone, an octave under the key.
		lone(48),

		// A pair where one rank did not sound: its cell prints absent and its beats with it.
		silent(64, -2.6),

		// Hand-edited after the fact, and the 4' never sounded.
		manual(74, -0.1, -2.5, 2.4),

		// The bass side: the whole ladder on one button, then a switch sounding two ranks of it.
		bassLadder(29),
		bassPartial(45),
	} {
		s.UpsertTake(t)
	}
	return s
}

func instrument() session.Instrument {
	m := func(b ...session.Bank) []session.Bank { return b }
	return session.Instrument{
		Name:      "Bugari 508/ARS",
		Serial:    "508-ARS-1174",
		ReedCount: reeds,
		Banks:     banks,
		Registers: []session.Register{
			{Name: "LMMH", Banks: m(session.BankL, session.BankM1, session.BankM2, session.BankH)},
			{Name: "MM", Banks: m(session.BankM1, session.BankM2)},
			{Name: "L", Banks: m(session.BankL)},
		},
		BassReeds:     5,
		BassRegisters: []session.BassRegister{{Name: "Soft", Feet: []int{16, 8}}},
		A4:            a4,
		Tolerance:     3,
		BeatTolerance: 4,
	}
}

// curve is the goal: the 8' ranks two and a half units either side of the key, the 16' and 4' on it.
func curve() *target.Curve {
	c := target.NewCurve("Musette", reeds)
	for _, note := range []tuning.Note{48, 60, 72, 84} {
		for reed, dev := range []float64{0, -2.5, 2.5, 0} {
			if err := c.Set(note, reed, dev); err != nil {
				fail("curve.Set(%d, %d): %v", note, reed, err)
			}
		}
	}
	return c
}

// reed builds one measured rank: dev reference units off the key, in the octave its rank sounds in
// (-12 for the 16', +12 for the 4').
func reed(note tuning.Note, octave int, dev float64) dsp.ReedMeasure {
	ref := note.Step(octave).Freq(a4)
	return dsp.ReedMeasure{Present: true, Freq: meter.FrequencyHz(ref, dev), Dev: dev, Octave: octave}
}

// quiet is a slot that did not sound, as the engine reports one.
func quiet(octave int) dsp.ReedMeasure {
	return dsp.ReedMeasure{Dev: meter.Missing, Octave: octave}
}

// take is a note's slots with the adjacent beats the engine derives from them.
func take(note tuning.Note, register string, slots ...dsp.ReedMeasure) session.Take {
	t := session.Take{Note: note, At: when, Register: register, Reeds: slots}
	for i := 1; i < len(slots); i++ {
		lo, hi := slots[i-1], slots[i]
		b := dsp.BeatMeasure{Pair: fmt.Sprintf("%d-%d", i, i+1)}
		if lo.Present && hi.Present {
			b.Present, b.Hz = true, meter.HarmonicBeat(lo.Freq, hi.Freq, note.Step(lo.Octave).Freq(a4), true)
		}
		t.Beats = append(t.Beats, b)
	}
	return t
}

// full is a note under the whole LMMH switch.
func full(note tuning.Note, l, m1, m2, h float64) session.Take {
	return take(note, "LMMH", reed(note, -12, l), reed(note, 0, m1), reed(note, 0, m2), reed(note, 12, h))
}

// middles is a note under the two 8' ranks alone.
func middles(note tuning.Note, m1, m2 float64) session.Take {
	return take(note, "MM", reed(note, 0, m1), reed(note, 0, m2))
}

// lone is the 16' rank on its own.
func lone(note tuning.Note) session.Take {
	return take(note, "L", reed(note, -12, -1.2))
}

// silent is the two 8' ranks with the second one not heard.
func silent(note tuning.Note, m1 float64) session.Take {
	return take(note, "MM", reed(note, 0, m1), quiet(0))
}

// manual is a row somebody typed over, with the 4' rank never sounding.
func manual(note tuning.Note, l, m1, m2 float64) session.Take {
	t := take(note, "LMMH", reed(note, -12, l), reed(note, 0, m1), reed(note, 0, m2), quiet(12))
	t.Manual = true
	return t
}

// bassLadder is one bass button with the whole machine sounding: five ranks an octave apart.
func bassLadder(note tuning.Note) session.Take {
	var slots []dsp.ReedMeasure
	for i, dev := range []float64{1.4, -0.8, 0.6, -1.9, 2.2} {
		slots = append(slots, reed(note, 12*i, dev))
	}
	t := take(note, "", slots...)
	t.Bass = true
	return t
}

// bassPartial is a bass switch sounding two ranks of the ladder, which must land under their own
// columns rather than the first two.
func bassPartial(note tuning.Note) session.Take {
	t := take(note, "Soft", reed(note, 0, 0.7), reed(note, 12, -1.1))
	t.Bass = true
	return t
}

// musetteExample is the common shape: one register, three 8' ranks, no bass machine. It fits the
// sheet's ordinary wide table, where the full example above overflows into the grouped fallback.
func musetteExample() *session.Session {
	m := func(b ...session.Bank) []session.Bank { return b }
	inst := session.Instrument{
		Name:      "Hohner Morino V N",
		Serial:    "M-88214",
		ReedCount: 3,
		Banks:     m(session.BankM1, session.BankM2, session.BankM3),
		Registers: []session.Register{
			{Name: "MMM", Banks: m(session.BankM1, session.BankM2, session.BankM3)},
		},
		A4:            a4,
		Tolerance:     3,
		BeatTolerance: 4,
	}

	s := session.New("Hohner Morino V N - "+text.date, inst, a4)

	c := target.NewCurve("Musette", 3)
	for _, note := range []tuning.Note{53, 65, 77} {
		for reedIdx, dev := range []float64{-2.5, 0, 2.5} {
			if err := c.Set(note, reedIdx, dev); err != nil {
				fail("curve.Set(%d, %d): %v", note, reedIdx, err)
			}
		}
	}
	s.Curve = c

	three := func(note tuning.Note, a, b, d float64) session.Take {
		return take(note, "MMM", reed(note, 0, a), reed(note, 0, b), reed(note, 0, d))
	}

	for _, t := range []session.Take{
		three(53, -2.4, 0.2, 2.6),
		three(57, -2.7, -0.1, 2.4),
		three(60, -2.5, 0.3, 2.5),
		three(62, -2.6, 0.1, 2.8),
		three(65, -2.3, -0.2, 2.7),
		three(69, -2.8, 0.4, 5.9), // the third rank out of tolerance
		three(72, -2.5, 0.1, 6.4), // and again, higher up: the drift the note describes
	} {
		s.UpsertTake(t)
	}
	return s
}
