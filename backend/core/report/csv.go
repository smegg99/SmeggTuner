package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"smegg.me/smeggtuner/backend/core/target"
)

// CSV writes the table as a spreadsheet with a leading key/value metadata block; a slot that did not sound has empty cells, never zeros.
func CSV(w io.Writer, r *Report) error {
	if r == nil {
		return ErrNoSession
	}
	out := csv.NewWriter(w)
	unit := unitSuffix(r)

	meta := [][]string{
		{"session", r.Identity.Session},
		{"accordion", r.Identity.Accordion},
		{"serial", r.Identity.Serial},
		{"reed_count", strconv.Itoa(r.Identity.ReedCount)},
		{"reed_banks", banksList(r)},
		{"recorded", r.Session.At.Format("2006-01-02 15:04:05")},
		{"reference_a4_hz", fmt.Sprintf("%.2f", r.Session.A4)},
		{"transpose_semitones", strconv.Itoa(r.Session.Transpose)},
		{"goal_curve", curveName(r)},
		{"deviation_unit", unitMeta(r)},
		{"tolerance_" + unit, fmt.Sprintf("%.2f", r.Summary.Tolerance)},
		{"beat_tolerance_" + unit, fmt.Sprintf("%.2f", r.Summary.BeatTol)},
		{"generated", r.Generated.Format("2006-01-02 15:04:05")},
		{},
	}
	for _, row := range meta {
		if err := out.Write(row); err != nil {
			return err
		}
	}

	// Reed columns are named after the rank (m2_curr_ref) when known, else the position (reed2_curr_ref).
	header := []string{"note", "note_name", "register", "manual"}
	for i, reed := range r.Reeds {
		key := fmt.Sprintf("reed%d", reed)
		if i < len(r.Banks) {
			key = strings.ToLower(string(r.Banks[i]))
		}
		header = append(header, deviationColumns(key, unit)...)
	}
	for _, p := range r.Pairs {
		header = append(header, beatColumns(fmt.Sprintf("beat%d_%d", p.Low, p.High), unit)...)
	}
	if err := out.Write(header); err != nil {
		return err
	}

	if err := writeRows(out, r.Rows, len(r.Reeds), len(r.Pairs)); err != nil {
		return err
	}

	// The bass side follows as its own block: a blank line, its own header (columns keyed by
	// foot), then its rows - a second table in one file, the way the printed sheet lays it.
	if r.Bass != nil {
		if err := out.Write(nil); err != nil {
			return err
		}
		header := []string{"note", "note_name", "register", "manual"}
		for i, reed := range r.Bass.Reeds {
			key := fmt.Sprintf("bass_reed%d", reed)
			if i < len(r.Bass.Feet) {
				key = fmt.Sprintf("bass_%dft", r.Bass.Feet[i])
			}
			header = append(header, deviationColumns(key, unit)...)
		}
		for _, p := range r.Bass.Pairs {
			header = append(header, beatColumns(fmt.Sprintf("bass_beat%d_%d", p.Low, p.High), unit)...)
		}
		if err := out.Write(header); err != nil {
			return err
		}
		if err := writeRows(out, r.Bass.Rows, len(r.Bass.Reeds), len(r.Bass.Pairs)); err != nil {
			return err
		}
	}

	out.Flush()
	return out.Error()
}

func writeRows(out *csv.Writer, rows []Row, reeds, pairs int) error {
	for _, row := range rows {
		rec := []string{
			strconv.Itoa(int(row.Note)),
			row.Name,
			row.Register,
			yesNo(row.Manual),
		}
		for i := 0; i < reeds; i++ {
			c := ReedCell{}
			if i < len(row.Reeds) {
				c = row.Reeds[i]
			}
			if !c.Present {
				// Empty, never zero: a slot that did not sound has no reading at all.
				rec = append(rec, "", "", "", "")
				continue
			}
			rec = append(rec, num(c.Curr), goalNum(c.Goal, c.HasGoal), num(c.Error), yesNo(c.InTol))
		}
		for i := 0; i < pairs; i++ {
			b := BeatCell{}
			if i < len(row.Beats) {
				b = row.Beats[i]
			}
			if !b.Present {
				rec = append(rec, "", "", "", "", "", "")
				continue
			}
			rec = append(rec,
				num(b.Curr), goalNum(b.Goal, b.HasGoal), num(b.Error),
				num(b.CurrHz), goalNum(b.GoalHz, b.HasGoal), yesNo(b.InTol))
		}
		if err := out.Write(rec); err != nil {
			return err
		}
	}
	return nil
}

// unitSuffix names the deviation columns' unit: ref for reference units, cents for cents.
func unitSuffix(r *Report) string {
	if r.Unit == target.UnitCents {
		return "cents"
	}
	return "ref"
}

func unitMeta(r *Report) string {
	if r.Unit == target.UnitCents {
		return "cents"
	}
	return "reference (1 = 0.06% of the note frequency)"
}

func deviationColumns(key, unit string) []string {
	return []string{key + "_curr_" + unit, key + "_goal_" + unit, key + "_error_" + unit, key + "_in_tol"}
}

func beatColumns(key, unit string) []string {
	return []string{key + "_curr_" + unit, key + "_goal_" + unit, key + "_error_" + unit,
		key + "_curr_hz", key + "_goal_hz", key + "_in_tol"}
}

func curveName(r *Report) string {
	if !r.Identity.HasCurve {
		return "none"
	}
	if r.Identity.CurveName == "" {
		return "unnamed"
	}
	return r.Identity.CurveName
}

func num(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) }

// goalNum prints a goal cell, empty where the curve sets none.
func goalNum(v float64, has bool) string {
	if !has {
		return ""
	}
	return num(v)
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func banksList(r *Report) string {
	names := make([]string, 0, len(r.Banks))
	for _, b := range r.Banks {
		names = append(names, string(b))
	}
	return strings.Join(names, ",")
}
