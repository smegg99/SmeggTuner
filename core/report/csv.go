package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// CSV writes the table as a spreadsheet with a leading key/value metadata block; a merged note has empty reed cells, never zeros.
func CSV(w io.Writer, r *Report) error {
	if r == nil {
		return ErrNoSession
	}
	out := csv.NewWriter(w)

	meta := [][]string{
		{"session", r.Identity.Session},
		{"make", r.Identity.Make},
		{"model", r.Identity.Model},
		{"serial", r.Identity.Serial},
		{"reed_count", strconv.Itoa(r.Identity.ReedCount)},
		{"reed_banks", banksList(r)},
		{"recorded", r.Session.At.Format("2006-01-02 15:04:05")},
		{"reference_a4_hz", fmt.Sprintf("%.2f", r.Session.A4)},
		{"goal_curve", curveName(r)},
		{"tolerance_cents", fmt.Sprintf("%.2f", r.Summary.Tolerance)},
		{"beat_tolerance_cents", fmt.Sprintf("%.2f", r.Summary.BeatTol)},
		{"generated", r.Generated.Format("2006-01-02 15:04:05")},
		{},
	}
	for _, row := range meta {
		if err := out.Write(row); err != nil {
			return err
		}
	}

	// Reed columns are named after the rank (m2_curr_cents) when known, else the position (reed2_curr_cents).
	header := []string{"note", "note_name", "register", "reeds", "manual"}
	for i, reed := range r.Reeds {
		key := fmt.Sprintf("reed%d", reed)
		if i < len(r.Banks) {
			key = strings.ToLower(string(r.Banks[i]))
		}
		header = append(header,
			key+"_curr_cents", key+"_goal_cents", key+"_error_cents", key+"_in_tol")
	}
	for _, p := range r.Pairs {
		key := fmt.Sprintf("beat%d_%d", p.Low, p.High)
		header = append(header,
			key+"_curr_cents", key+"_goal_cents", key+"_error_cents",
			key+"_curr_hz", key+"_goal_hz", key+"_in_tol", key+"_source")
	}
	if err := out.Write(header); err != nil {
		return err
	}

	for _, row := range r.Rows {
		rec := []string{
			strconv.Itoa(int(row.Note)),
			row.Name,
			row.Register,
			reedState(row),
			yesNo(row.Manual),
		}
		for i := range r.Reeds {
			c := ReedCell{}
			if i < len(row.Reeds) {
				c = row.Reeds[i]
			}
			if !c.Present {
				// Empty, never zero: a merged or unsounded reed has no reading at all.
				rec = append(rec, "", "", "", "")
				continue
			}
			rec = append(rec, num(c.Curr), num(c.Goal), num(c.Error), yesNo(c.InTol))
		}
		for i := range r.Pairs {
			b := BeatCell{}
			if i < len(row.Beats) {
				b = row.Beats[i]
			}
			if !b.Present {
				rec = append(rec, "", "", "", "", "", "", "")
				continue
			}
			rec = append(rec,
				num(b.Curr), num(b.Goal), num(b.Error),
				num(b.CurrHz), num(b.GoalHz), yesNo(b.InTol), beatSource(b))
		}
		if err := out.Write(rec); err != nil {
			return err
		}
	}

	out.Flush()
	return out.Error()
}

// reedState is the spreadsheet's version of the M and D marks on the printed sheet.
func reedState(row Row) string {
	switch {
	case row.Merged:
		return "merged"
	case row.Derived:
		return "from_beat"
	default:
		return "measured"
	}
}

func beatSource(b BeatCell) string {
	if b.FromEnvelope {
		return "envelope"
	}
	return "spectrum"
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
