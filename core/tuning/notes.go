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

func (n Note) Name(s ScaleNaming) string {
	pc := int(n) % 12
	switch s {
	case NamingCDEFGAH:
		// German convention: B natural is H, A# is written B.
		name := sharpNames[pc]
		if pc == 11 {
			name = "H"
		} else if pc == 10 {
			name = "B"
		}
		return fmt.Sprintf("%s%d", name, n.Octave())
	case NamingDoReMi:
		return fmt.Sprintf("%s%d", doReMi[pc], n.Octave())
	case NamingPolish:
		return fmt.Sprintf("%s%d", polishNames[pc], n.Octave())
	default:
		return fmt.Sprintf("%s%d", sharpNames[pc], n.Octave())
	}
}

func (n Note) Transpose(semitones int) Note {
	t := n + Note(semitones)
	if t < MinNote {
		return MinNote
	}
	if t > MaxNote {
		return MaxNote
	}
	return t
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
