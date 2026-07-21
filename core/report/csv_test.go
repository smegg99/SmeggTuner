package report

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
)

func TestCSV(t *testing.T) {
	r := sheet(t, musette(t))
	var buf bytes.Buffer
	if err := CSV(&buf, r); err != nil {
		t.Fatalf("CSV: %v", err)
	}
	raw := buf.String()

	for _, lobe := range lobes() {
		if strings.Contains(raw, lobe) {
			t.Fatalf("the merged pair's lobe %s reached the spreadsheet", lobe)
		}
	}

	rows, err := csv.NewReader(strings.NewReader(raw)).ReadAll()
	if err == nil {
		t.Fatal("a rectangular reader should have refused the metadata block; use FieldsPerRecord = -1")
	}

	rd := csv.NewReader(strings.NewReader(raw))
	rd.FieldsPerRecord = -1
	rows, err = rd.ReadAll()
	if err != nil {
		t.Fatalf("CSV does not parse: %v", err)
	}

	meta := map[string]string{}
	var header []string
	var body [][]string
	for _, rec := range rows {
		switch {
		case len(rec) == 2 && header == nil:
			meta[rec[0]] = rec[1]
		case len(rec) > 2 && header == nil:
			header = rec
		case len(rec) > 2:
			body = append(body, rec)
		}
	}

	if meta["reference_a4_hz"] != "435.00" {
		t.Errorf("reference_a4_hz = %q, want the session's 435.00", meta["reference_a4_hz"])
	}
	if meta["session"] != "Hohner Morino - Jan K." {
		t.Errorf("the metadata does not identify the instrument: %v", meta)
	}
	if len(body) != 5 {
		t.Fatalf("%d data rows, want 5", len(body))
	}

	at := func(rec []string, col string) string {
		for i, name := range header {
			if name == col && i < len(rec) {
				return rec[i]
			}
		}
		t.Fatalf("no column %q", col)
		return ""
	}

	for _, rec := range body {
		switch at(rec, "note") {
		case "64": // the merged note
			if got := at(rec, "reeds"); got != "merged" {
				t.Errorf("the merged row says reeds=%q", got)
			}
			for _, col := range []string{"reed1_curr_cents", "reed2_curr_cents", "reed1_error_cents"} {
				if got := at(rec, col); got != "" {
					t.Errorf("the merged row carries %s=%q; it has no per-reed reading, and a zero would "+
						"be read as one", col, got)
				}
			}
			if at(rec, "beat1_2_curr_cents") == "" || at(rec, "beat1_2_source") != "envelope" {
				t.Error("the merged row lost the beat, which is the only reading it has")
			}
		case "65": // recovered from the beat
			if got := at(rec, "reeds"); got != "from_beat" {
				t.Errorf("the derived row says reeds=%q", got)
			}
			if at(rec, "reed1_curr_cents") == "" {
				t.Error("the derived row dropped reeds that were measured")
			}
		case "60":
			if at(rec, "reed3_in_tol") != "no" {
				t.Error("the out-of-tolerance reed is not marked in the spreadsheet")
			}
		case "67": // the reed that never sounded
			if at(rec, "reed3_curr_cents") != "" {
				t.Error("a reed that did not sound is printed as a reading")
			}
			if at(rec, "manual") != "yes" {
				t.Error("the hand-edited row is not marked")
			}
		}
	}
}
