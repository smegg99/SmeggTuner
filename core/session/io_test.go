package session

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
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

// writeLegacy writes a session as bare legacy JSON, bypassing Save validation.
func writeLegacy(t *testing.T, s *Session) string {
	t.Helper()
	data, err := json.Marshal(sessionFile{V: Version, Session: *s})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), s.ID+LegacyFileExt)
	if err := os.WriteFile(path, data, filePerm); err != nil {
		t.Fatal(err)
	}
	return path
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
	path := writeLegacy(t, s)

	got, err := Read(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Curve.Anchors[0].Note != 53 || got.Curve.Anchors[1].Note != 89 {
		t.Fatalf("anchors came back unsorted: %+v", got.Curve.Anchors)
	}
}

// A future or missing version is refused, not guessed at.
func TestVersionMismatch(t *testing.T) {
	for name, body := range map[string]string{
		"future":  `{"v":3,"id":"abc","a4":442,"instrument":{"reedCount":3}}`,
		"missing": `{"id":"abc","a4":442,"instrument":{"reedCount":3}}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "abc"+LegacyFileExt)
			if err := os.WriteFile(path, []byte(body), filePerm); err != nil {
				t.Fatal(err)
			}
			_, err := Read(path)
			if !errors.Is(err, ErrVersion) {
				t.Fatalf("read = %v, want ErrVersion", err)
			}
		})
	}
}

// Unknown fields keep the file loadable, not refused.
func TestUnknownFieldsLoad(t *testing.T) {
	s := recorded(t, 3)
	path := writeLegacy(t, s)

	var raw map[string]any
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	raw["temperature"] = 21.5
	raw["takes"].([]any)[0].(map[string]any)["humidity"] = 40
	patched, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, patched, filePerm); err != nil {
		t.Fatal(err)
	}

	got, err := Read(path)
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

// Absent interpolation flags mean the file predates them; loaded as on, not off.
func TestOldCurveFileLoadsWithInterpolationOn(t *testing.T) {
	old := `{
	  "v": 1,
	  "id": "` + NewID() + `",
	  "name": "Hohner Morino",
	  "instrument": {"reedCount": 3},
	  "a4": 442,
	  "curve": {
	    "name": "musette",
	    "reedCount": 3,
	    "refReed": 1,
	    "unit": "cent",
	    "anchors": [
	      {"note": 60, "reeds": [-4, 0, 4]},
	      {"note": 72, "reeds": [-8, 0, 8]}
	    ]
	  },
	  "created": "2024-01-01T00:00:00Z",
	  "updated": "2024-01-01T00:00:00Z"
	}`
	path := filepath.Join(t.TempDir(), "old"+LegacyFileExt)
	if err := os.WriteFile(path, []byte(old), filePerm); err != nil {
		t.Fatal(err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	c := got.Curve
	if c == nil {
		t.Fatal("the curve did not survive the load")
	}
	if !c.Interpolate || !c.ExtrapolateLeft || !c.ExtrapolateRight {
		t.Fatalf("an old file came back with the flags off: %+v", c)
	}
	if c.Asymmetry != 0 {
		t.Fatalf("asymmetry = %v, want 0", c.Asymmetry)
	}
	if v := c.At(66)[2]; math.Abs(v-6) > 1e-12 {
		t.Fatalf("reed 3 at note 66 = %v, want 6: the old curve stopped interpolating", v)
	}
	if v := c.At(30)[2]; math.Abs(v-4) > 1e-12 {
		t.Fatalf("reed 3 below the first anchor = %v, want 4: the old curve stopped extrapolating", v)
	}
	if v := c.At(100)[2]; math.Abs(v-8) > 1e-12 {
		t.Fatalf("reed 3 above the last anchor = %v, want 8: the old curve stopped extrapolating", v)
	}
}
