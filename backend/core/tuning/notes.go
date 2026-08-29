// Package tuning holds note and cents math for the measurement core.
package tuning

import (
	"fmt"
	"math"
)

type Note int

const (
	MinNote  Note = 16
	MaxNote  Note = 120
	NoteA4   Note = 69
	NumNotes      = int(MaxNote-MinNote) + 1
)

type ScaleNaming int

const (
	NamingCDEFGAB ScaleNaming = iota
	NamingCDEFGAH
	NamingDoReMi
	// NamingPolish spells sharps out (Cis, Dis, Fis, Gis, Ais) and uses H for B natural; the recordings are named with this convention.
	NamingPolish
)

var sharpNames = [12]string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
var doReMi = [12]string{"Do", "Do#", "Re", "Re#", "Mi", "Fa", "Fa#", "Sol", "Sol#", "La", "La#", "Si"}

// H is B natural; A# uses the sharp spelling Ais rather than flat B.
var polishNames = [12]string{"C", "Cis", "D", "Dis", "E", "F", "Fis", "G", "Gis", "A", "Ais", "H"}

// ParseNaming reads the scale_naming config value; an unknown value falls back to C-to-B with sharps.
func ParseNaming(s string) ScaleNaming {
	switch s {
	case "cdefgah":
		return NamingCDEFGAH
	case "doremi":
		return NamingDoReMi
	case "polish":
		return NamingPolish
	default:
		return NamingCDEFGAB
	}
}

func (n Note) Valid() bool { return n >= MinNote && n <= MaxNote }

func (n Note) Freq(a4 float64) float64 {
	return a4 * math.Exp2(float64(n-NoteA4)/12)
}

func (n Note) Octave() int { return int(n)/12 - 1 }

func (n Note) Name(s ScaleNaming) string { return fmt.Sprintf("%s%d", n.PitchName(s), n.Octave()) }

// PitchName is the pitch-class name without an octave, as a chord root is written.
func (n Note) PitchName(s ScaleNaming) string {
	pc := int(n) % 12
	switch s {
	case NamingCDEFGAH:
		// German convention: B natural is H, A# is written B.
		switch pc {
		case 11:
			return "H"
		case 10:
			return "B"
		}
		return sharpNames[pc]
	case NamingDoReMi:
		return doReMi[pc]
	case NamingPolish:
		return polishNames[pc]
	default:
		return sharpNames[pc]
	}
}

// ManualStart is the note manual selection pins on entry: the shown note, else A4.
func ManualStart(shown Note) Note {
	if shown.Valid() {
		return shown
	}
	return NoteA4
}

// Step moves a manual note by semitones (12 an octave) within MinNote..MaxNote; an unset note pins A4 instead.
func (n Note) Step(semitones int) Note {
	if !n.Valid() {
		return NoteA4
	}
	return min(MaxNote, max(MinNote, n+Note(semitones)))
}

func Cents(f, ref float64) float64 { return 1200 * math.Log2(f/ref) }

func FreqAtCents(ref, cents float64) float64 { return ref * math.Exp2(cents/1200) }

func Nearest(freq, a4 float64) (Note, float64) {
	n := Note(math.Round(12*math.Log2(freq/a4))) + NoteA4
	if n < MinNote {
		n = MinNote
	}
	if n > MaxNote {
		n = MaxNote
	}
	return n, Cents(freq, n.Freq(a4))
}
