package report

import (
	"smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// judging is what every row of the sheet is scored and printed by.
type judging struct {
	curve        *target.Curve
	a4           float64
	transpose    int
	tol, beatTol float64 // reference units
	unit         target.Unit
}

func buildRows(
	display []session.DisplayRow, j judging, banks []session.Bank,
	reedCount int, naming tuning.ScaleNaming, intoFor func(session.Take) []int,
) ([]Row, []Pair) {
	rows := make([]Row, 0, len(display))
	// Beats are keyed by the columns they join, not by the slots: a partial register's slots and
	// the card's columns do not line up, and keying by slot puts an octave beat in the musette column.
	beats := make([]map[[2]int]target.BeatError, 0, len(display))

	for _, d := range display {
		reading := d.Take.Reading(j.transpose)
		row := Row{
			Note:     d.Note,
			Name:     d.Note.Name(naming),
			Register: d.Register,
			At:       d.Take.At,
			Manual:   d.Take.Manual,
		}

		// Which column each slot answers to. The goal is read there too, or a register short of
		// the instrument's full one is judged against another rank's goal.
		into := intoFor(d.Take)

		errs := target.Errors(reading, j.curve, j.a4, j.tol, into)
		row.Reeds = reedCells(errs, reedCount, banks, into, j.unit)
		for _, cell := range row.Reeds {
			if cell.Present && !cell.InTol {
				row.OutOfTol++
			}
		}

		byCols := make(map[[2]int]target.BeatError)
		for _, b := range target.BeatErrors(reading, j.curve, j.a4, j.beatTol, into) {
			byCols[[2]int{colOf(into, b.Low), colOf(into, b.High)}] = b.In(j.unit)
		}
		beats = append(beats, byCols)
		rows = append(rows, row)
	}

	pairs := pairsOf(reedCount, beats)
	for i := range rows {
		rows[i].Beats = beatCells(beats[i], pairs)
		for _, cell := range rows[i].Beats {
			if cell.Present && !cell.InTol {
				rows[i].OutOfTol++
			}
		}
	}
	return rows, pairs
}

func manyRegisters(rows []Row) bool {
	first := ""
	for i, r := range rows {
		if i == 0 {
			first = r.Register
			continue
		}
		if r.Register != first {
			return true
		}
	}
	return false
}
