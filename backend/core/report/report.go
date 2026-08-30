// Package report renders a recording pass as the bench sheet: per note and reed, the
// reading, the goal, the difference, and the beat between neighbouring ranks.
package report

import (
	"errors"
	"time"

	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
)

var (
	// ErrNoSession reports a report asked of nothing.
	ErrNoSession = errors.New("report: no session")
	// ErrNoReadings reports a session with nothing recorded in it.
	ErrNoReadings = errors.New("report: session holds no readings")
	// ErrUnit reports a sheet asked for in a unit it cannot print deviations in.
	ErrUnit = errors.New("report: deviations print in reference units or cents")
)

// Letterhead is the technician's own heading, off unless the caller asks for it. It carries
// words and no artwork: the card is branded, and an image slot here is a place to paint over it.
type Letterhead struct {
	CompanyName    string
	CompanyAddress string
	CompanyWebsite string
}

// Options is everything the caller decides about a report.
type Options struct {
	Letterhead *Letterhead
	// Date printed on the sheet; zero means the pass's own date.
	Date time.Time
	// Now is the generation stamp, the seam tests pin time at; zero means time.Now().
	Now time.Time
	// Naming is the note-name convention the tuner screen is set to.
	Naming tuning.ScaleNaming
	// Tolerance and BeatTolerance are the judging windows in reference units; not positive means
	// core/target defaults, and InTol travels as computed, never re-derived here.
	Tolerance     float64
	BeatTolerance float64
	// Unit is what the sheet prints deviations in: target.UnitReference (the default) or target.UnitCents.
	Unit target.Unit
}

// Identity is the instrument the sheet is about.
type Identity struct {
	Session string
	// Accordion is the instrument's own name, the one the technician calls it by.
	Accordion  string
	Serial     string
	ReedCount  int
	Notes      string
	Registers  []session.Register
	CurveName  string
	HasCurve   bool
	CurveReeds int
}

// SessionInfo is what the sheet says about the sitting. A4 and Transpose are the pass's own
// frozen reference: the instrument's C sounded as C+Transpose.
type SessionInfo struct {
	At        time.Time
	A4        float64
	Transpose int
	// Key and Sounds name a transposition on the sheet: the instrument's C and the pitch it sounded as.
	Key, Sounds string
	Readings    int
}

// ReedCell is one rank of one note in the sheet's unit; Present is false where the take had no
// such slot or it did not sound, and the cell prints absent rather than as zero.
type ReedCell struct {
	Reed int // 1-based, as it prints
	// Bank is the rank this column stands for; empty when columns are numbered.
	Bank    session.Bank
	Present bool
	// HasGoal is false where the curve sets no goal; Goal is then 0 and prints blank.
	HasGoal bool
	Curr    float64
	Goal    float64
	Error   float64
	InTol   bool
}

// Pair is one beat column: the two reeds it sits between.
type Pair struct {
	Key  string // "1-2", as dsp.BeatMeasure spells it
	Low  int    // 1-based, as it prints
	High int
}

// BeatCell is one rank pair of one note: Curr, Goal and Error in the sheet's unit, the rest in Hz.
type BeatCell struct {
	Present bool
	HasGoal bool
	Curr    float64
	Goal    float64
	Error   float64
	CurrHz  float64
	GoalHz  float64
	ErrorHz float64
	InTol   bool
}

// Row is one note of the pass.
type Row struct {
	Note tuning.Note
	Name string
	// Register is the switch this note was played on; its own column when there's more than one.
	Register string
	At       time.Time
	// Manual marks a row whose value was typed rather than heard.
	Manual bool
	Reeds  []ReedCell
	Beats  []BeatCell
	// OutOfTol counts this row's cells the backend judged out of tolerance.
	OutOfTol int
}

// Summary is the sheet's top line: how much of this instrument is done.
type Summary struct {
	Notes     int
	Reeds     int // reed readings printed
	ReedsOut  int
	Beats     int
	BeatsOut  int
	Manual    int
	NotesOut  int // notes with at least one cell out of tolerance
	MinNote   tuning.Note
	MaxNote   tuning.Note
	MinName   string
	MaxName   string
	Tolerance float64 // in the sheet's unit, as printed
	BeatTol   float64
}

// Report is a rendered pass: everything the templates print.
type Report struct {
	Identity   Identity
	Session    SessionInfo
	Letterhead *Letterhead
	Date       time.Time
	Generated  time.Time
	// Unit is what every deviation on the sheet is printed in.
	Unit target.Unit

	// Reeds is the columns: 1..Identity.ReedCount.
	Reeds []int

	// Banks names the columns (L, M1, M2, M3, H) where the instrument says so; same length as Reeds, or empty.
	Banks []session.Bank

	// MultiRegister says the pass caught more than one switch.
	MultiRegister bool
	Pairs         []Pair
	Rows          []Row

	// Bass is the bass side's own table, when the pass recorded any: columns are the machine's
	// ranks by foot, rows its buttons at their real pitches. Nil on an all-treble pass.
	Bass *BassPart

	Summary Summary
	Graph   *Graph
	Layout  Layout
}

// Columns is the width of the bench table - the wider of the two keyboards' - and it drives the
// page layout, so both tables share one orientation and grouping.
func (r *Report) Columns() int {
	main := 1 + r.VoiceColumns() + 3*len(r.Reeds) + 3*len(r.Pairs)
	if r.Bass == nil {
		return main
	}
	return max(main, 1+r.Bass.VoiceColumns()+3*len(r.Bass.Reeds)+3*len(r.Bass.Pairs))
}

// VoiceColumns is the columns needed to label a reading, zero unless multi-register.
func (r *Report) VoiceColumns() int {
	n := 0
	if r.MultiRegister {
		n++
	}
	return n
}

// Head is a reed column's heading: the rank's own name (untranslated) or the translated reed number.
func (r *Report) Head(i int) string {
	if i >= 0 && i < len(r.Banks) {
		return string(r.Banks[i])
	}
	return i18n.ReportHeadReed(i18n.Localizer(), i18n.ReportHeadReedParams{Index: i + 1})
}
