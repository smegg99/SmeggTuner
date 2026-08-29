package session

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
)

func recorded(t *testing.T, reedCount int) *Session {
	t.Helper()
	s := New("Hohner Morino - Jan K.", Instrument{
		Serial:    "12345",
		ReedCount: reedCount,
		Banks:     Banks[:reedCount],
		Registers: []Register{{Name: "Musette", Banks: Banks[:reedCount]}},
	}, 442)
	s.Notes = "left hand block reglued"
	now := time.Now()
	for i, n := range []tuning.Note{69, 60, 69} {
		s.UpsertTake(take(n, reedCount, now.Add(time.Duration(i)*time.Second)))
	}
	return s
}

// writeStsf packs a raw session body behind a manifest of version v, bypassing WriteSessionFile's validation.
func writeStsf(t *testing.T, v int, body []byte) string {
	t.Helper()
	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	if err := writeEntry(z, entryManifest, mustJSON(manifest{V: v, Kind: sessionFileKind})); err != nil {
		t.Fatal(err)
	}
	if err := writeEntry(z, entrySession, body); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "raw"+SessionFileExt)
	if err := writeFileAtomic(path, buf.Bytes()); err != nil {
		t.Fatal(err)
	}
	return path
}

func body(t *testing.T, s *Session) []byte {
	t.Helper()
	data, err := json.Marshal(sessionFile{V: Version, Session: *s})
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// A hand-edited file may have its anchors in any order; a curve is only valid sorted.
func TestCurveAnchorsAreSortedOnLoad(t *testing.T) {
	s := recorded(t, 3)
	c := target.NewCurve("musette", 3)
	c.Anchors = []target.Anchor{
		{Note: 89, Reeds: []float64{-9, 0, 9}},
		{Note: 53, Reeds: []float64{-3, 0, 3}},
	}
	s.Curve = c

	got, err := ReadSessionFile(writeStsf(t, SessionFileVersion, body(t, s)))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Curve.Anchors[0].Note != 53 || got.Curve.Anchors[1].Note != 89 {
		t.Fatalf("anchors came back unsorted: %+v", got.Curve.Anchors)
	}
}

// Any schema version but this one is refused, not guessed at.
func TestVersionMismatch(t *testing.T) {
	for name, raw := range map[string]string{
		"zero":    `{"v":0,"id":"abc","a4":442,"instrument":{"reedCount":3}}`,
		"future":  `{"v":2,"id":"abc","a4":442,"instrument":{"reedCount":3}}`,
		"missing": `{"id":"abc","a4":442,"instrument":{"reedCount":3}}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ReadSessionFile(writeStsf(t, SessionFileVersion, []byte(raw)))
			if !errors.Is(err, ErrVersion) {
				t.Fatalf("read = %v, want ErrVersion", err)
			}
		})
	}
}

// Unknown fields keep the file loadable, not refused.
func TestUnknownFieldsLoad(t *testing.T) {
	var raw map[string]any
	if err := json.Unmarshal(body(t, recorded(t, 3)), &raw); err != nil {
		t.Fatal(err)
	}
	raw["temperature"] = 21.5
	raw["takes"].([]any)[0].(map[string]any)["humidity"] = 40
	patched, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}

	got, err := ReadSessionFile(writeStsf(t, SessionFileVersion, patched))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got.Takes) != 2 {
		t.Fatalf("readings = %d, want 2", len(got.Takes))
	}
}

// The ID comes from the frontend and ends up in a query and a URL, so it is validated.
func TestValidID(t *testing.T) {
	for _, id := range []string{"", "..", "../../etc/passwd", "a/b", "a.b", strings.Repeat("a", 65)} {
		if ValidID(id) {
			t.Fatalf("ValidID(%q) = true, want false", id)
		}
	}
	if !ValidID(NewID()) {
		t.Fatalf("NewID produced an id ValidID rejects: %q", NewID())
	}
}

// A curve with no anchors is no beat list: the summary does not offer it as one.
func TestSummaryHasCurveOnlyWithAnchors(t *testing.T) {
	s := newTestSession(t, 3)
	if s.Summarize().HasCurve {
		t.Fatal("a session with no curve has one")
	}
	s.Curve = target.NewCurve("empty", 3)
	if s.Summarize().HasCurve {
		t.Fatal("an anchorless curve is listed as a goal")
	}
	if err := s.Curve.Set(60, 0, 1); err != nil {
		t.Fatal(err)
	}
	if !s.Summarize().HasCurve {
		t.Fatal("a curve with an anchor is not listed")
	}
}
