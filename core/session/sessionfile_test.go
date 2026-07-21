package session

import (
	"archive/zip"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
)

func stsfPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), name+SessionFileExt)
}

// A session, out to a .stsf and back in, with everything a report is written from intact.
func TestASessionSurvivesTheRoundTrip(t *testing.T) {
	s := recorded(t, 5)
	s.Takes[1].ReedsMerged = true
	s.Curve = &target.Curve{
		Name:      "musette",
		ReedCount: 5,
		RefReed:   1,
		Unit:      target.UnitCents,
		Anchors: []target.Anchor{
			{Note: 53, Reeds: []float64{-3, 0, 3, 6, 9}},
			{Note: 89, Reeds: []float64{-9, 0, 9, 18, 27}},
		},
	}

	path := stsfPath(t, "morino")
	if err := WriteSessionFile(path, s); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := ReadSessionFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Name != s.Name || got.Notes != s.Notes || got.A4 != s.A4 {
		t.Fatalf("identity lost: %+v", got)
	}
	if got.Instrument.ReedCount != 5 || len(got.Instrument.Registers) != 1 {
		t.Fatalf("instrument lost: %+v", got.Instrument)
	}
	// 69, 60, 69 was played: the second 69 replaces the first, so two voices survive in capture order.
	want := []tuning.Note{69, 60}
	if len(got.Takes) != len(want) {
		t.Fatalf("readings = %d, want %d", len(got.Takes), len(want))
	}
	for i, n := range want {
		if got.Takes[i].Note != n {
			t.Fatalf("reading %d = %d, want %d", i, got.Takes[i].Note, n)
		}
		if len(got.Takes[i].Reeds) != 5 {
			t.Fatalf("reading %d carries %d reeds", i, len(got.Takes[i].Reeds))
		}
	}
	// A merged pair has to survive the trip: it is the flag the report checks before a per-reed row.
	if got.Takes[0].ReedsMerged || !got.Takes[1].ReedsMerged {
		t.Fatalf("merged flags came back as %v / %v",
			got.Takes[0].ReedsMerged, got.Takes[1].ReedsMerged)
	}
	if got.Curve == nil {
		t.Fatal("the goal curve did not survive the round trip")
	}
	if got.Curve.Name != "musette" || got.Curve.ReedCount != 5 || got.Curve.RefReed != 1 {
		t.Fatalf("curve header lost: %+v", got.Curve)
	}
	if len(got.Curve.Anchors) != 2 {
		t.Fatalf("anchors = %d, want 2", len(got.Curve.Anchors))
	}
	a := got.Curve.Anchors[1]
	if a.Note != 89 || len(a.Reeds) != 5 || a.Reeds[4] != 27 {
		t.Fatalf("anchor lost: %+v", a)
	}
}

// ReadSessionAny takes either format: the .stsf now, or the bare legacy JSON it used to.
func TestReadSessionAnyTakesBothFormats(t *testing.T) {
	s := recorded(t, 3)

	stsf := stsfPath(t, "morino")
	if err := WriteSessionFile(stsf, s); err != nil {
		t.Fatal(err)
	}
	legacy := writeLegacy(t, s)

	for _, path := range []string{stsf, legacy} {
		got, err := ReadSessionAny(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if got.ID != s.ID || len(got.Takes) != 2 {
			t.Fatalf("%s came back as %+v", path, got)
		}
	}
}

// The session inside a .stsf is versioned and readable: anybody can unzip it and find indented JSON.
func TestTheFileIsVersionedAndReadable(t *testing.T) {
	s := recorded(t, 3)
	path := stsfPath(t, "morino")
	if err := WriteSessionFile(path, s); err != nil {
		t.Fatalf("write: %v", err)
	}

	z, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("it is not a zip: %v", err)
	}
	defer z.Close()

	data, err := readEntry(&z.Reader, entrySession, 16<<20)
	if err != nil {
		t.Fatalf("no session in it: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if raw["v"] != float64(Version) {
		t.Fatalf("v = %v, want %d", raw["v"], Version)
	}
	if !strings.Contains(string(data), "\n  ") {
		t.Fatal("the file is meant to be human-readable, so it is indented")
	}
}
