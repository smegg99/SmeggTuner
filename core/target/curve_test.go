package target

import (
	"encoding/json"
	"math"
	"testing"

	"smegg.me/smeggtuner/core/tuning"
)

func almost(t *testing.T, got, want, tol float64, msg string) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Fatalf("%s: got %v want %v (tol %v)", msg, got, want, tol)
	}
}

func mustSet(t *testing.T, c *Curve, note tuning.Note, reed int, value, a4 float64) {
	t.Helper()
	if err := c.Set(note, reed, value, a4); err != nil {
		t.Fatalf("Set(%d,%d,%v): %v", note, reed, value, err)
	}
}

func TestSetKeepsSortOrder(t *testing.T) {
	c := NewCurve("sorted", 2)
	for _, n := range []tuning.Note{84, 60, 120, 16, 72, 60} {
		mustSet(t, c, n, 0, float64(n), 440)
	}
	want := []tuning.Note{16, 60, 72, 84, 120}
	if len(c.Anchors) != len(want) {
		t.Fatalf("got %d anchors, want %d", len(c.Anchors), len(want))
	}
	for i, n := range want {
		if c.Anchors[i].Note != n {
			t.Fatalf("anchor %d is note %d, want %d", i, c.Anchors[i].Note, n)
		}
		if len(c.Anchors[i].Reeds) != 2 {
			t.Fatalf("anchor %d has %d reeds, want 2", i, len(c.Anchors[i].Reeds))
		}
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	c.Clear(72)
	c.Clear(72) // clearing what is not there is a no-op
	c.Clear(200)
	if len(c.Anchors) != 4 || c.Anchors[2].Note != 84 {
		t.Fatalf("clear left %v", c.Anchors)
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("validate after clear: %v", err)
	}
}

func TestReedCounts(t *testing.T) {
	one := NewCurve("clarinet", 1)
	mustSet(t, one, tuning.NoteA4, 0, 3, 440)
	if got := one.At(60); len(got) != 1 || got[0] != 3 {
		t.Fatalf("1-reed At = %v", got)
	}
	if err := one.Set(tuning.NoteA4, 1, 0, 440); err == nil {
		t.Fatal("reed 1 of a 1-reed curve must be an error")
	}

	five := NewCurve("bass", 5)
	five.RefReed = 2
	mustSet(t, five, 40, 4, -2, 440)
	mustSet(t, five, 80, 4, 6, 440)
	got := five.At(60)
	if len(got) != 5 {
		t.Fatalf("5-reed At gave %d values", len(got))
	}
	almost(t, got[4], 2, 1e-12, "reed 5 halfway")
	if err := five.Validate(); err != nil {
		t.Fatalf("5-reed validate: %v", err)
	}
	if err := five.Set(60, 5, 0, 440); err == nil {
		t.Fatal("reed 5 of a 5-reed curve must be an error")
	}
}

func TestOutOfRangeIsErrorNotPanic(t *testing.T) {
	c := NewCurve("range", 3)
	if err := c.Set(tuning.MaxNote+1, 0, 1, 440); err == nil {
		t.Fatal("note above the range must be an error")
	}
	if err := c.Set(tuning.MinNote-1, 0, 1, 440); err == nil {
		t.Fatal("note below the range must be an error")
	}
	if err := c.Set(tuning.NoteA4, -1, 1, 440); err == nil {
		t.Fatal("negative reed must be an error")
	}
	if len(c.Anchors) != 0 {
		t.Fatalf("a rejected Set must not anchor anything: %v", c.Anchors)
	}

	// A note outside the range still answers flat rather than panicking.
	mustSet(t, c, tuning.NoteA4, 1, 5, 440)
	almost(t, c.At(200)[1], 5, 1e-12, "above the range")
	almost(t, c.At(0)[1], 5, 1e-12, "below the range")

	c.Unit = UnitHz
	if err := c.Set(tuning.NoteA4, 0, -1000, 440); err == nil {
		t.Fatal("-1000 Hz at A4 must be an error")
	}
}

func TestValidate(t *testing.T) {
	bad := []*Curve{
		{ReedCount: 0, RefReed: NoRefReed, Unit: UnitCents},
		{ReedCount: MaxReeds + 1, RefReed: NoRefReed, Unit: UnitCents},
		{ReedCount: 3, RefReed: 3, Unit: UnitCents},
		{ReedCount: 3, RefReed: -2, Unit: UnitCents},
		{ReedCount: 3, RefReed: NoRefReed, Unit: "semitone"},
		{ReedCount: 3, RefReed: NoRefReed, Unit: UnitCents, Anchors: []Anchor{{Note: 200}}},
		{ReedCount: 2, RefReed: NoRefReed, Unit: UnitCents,
			Anchors: []Anchor{{Note: 60, Reeds: []float64{0, 0, 0}}}},
		{ReedCount: 2, RefReed: NoRefReed, Unit: UnitCents,
			Anchors: []Anchor{{Note: 72}, {Note: 60}}},
		{ReedCount: 2, RefReed: NoRefReed, Unit: UnitCents,
			Anchors: []Anchor{{Note: 60}, {Note: 60}}},
	}
	for i, c := range bad {
		if err := c.Validate(); err == nil {
			t.Fatalf("curve %d should not validate: %+v", i, c)
		}
	}

	// Sort fixes an out-of-order list; the duplicate below stays an error.
	c := &Curve{ReedCount: 2, RefReed: NoRefReed, Unit: UnitCents,
		Anchors: []Anchor{{Note: 84}, {Note: 60}, {Note: 72}}}
	c.Sort()
	if err := c.Validate(); err != nil {
		t.Fatalf("sorted curve should validate: %v", err)
	}
	if c.Anchors[0].Note != 60 || c.Anchors[2].Note != 84 {
		t.Fatalf("sort left %v", c.Anchors)
	}
}

// The three flags default to true; a file written before they existed decodes with them on.
func TestFlagsDefaultTrueOnDecode(t *testing.T) {
	var old Curve
	if err := json.Unmarshal([]byte(`{
		"name": "musette",
		"reedCount": 3,
		"refReed": 1,
		"unit": "cent",
		"anchors": [
			{"note": 60, "reeds": [-4, 0, 4]},
			{"note": 72, "reeds": [-8, 0, 8]}
		]
	}`), &old); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !old.Interpolate || !old.ExtrapolateLeft || !old.ExtrapolateRight {
		t.Fatalf("an old curve came back with the flags off: %+v", old)
	}
	if old.Asymmetry != 0 {
		t.Fatalf("asymmetry = %v, want 0", old.Asymmetry)
	}
	almost(t, old.At(66)[2], 6, 1e-12, "an old curve still interpolates")
	almost(t, old.At(30)[2], 4, 1e-12, "an old curve still holds flat below")
	almost(t, old.At(100)[2], 8, 1e-12, "an old curve still holds flat above")

	// A written false survives the round-trip.
	off := NewCurve("off", 3)
	off.Interpolate, off.ExtrapolateLeft, off.ExtrapolateRight = false, false, false
	off.Asymmetry = -25
	data, err := json.Marshal(off)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	var back Curve
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if back.Interpolate || back.ExtrapolateLeft || back.ExtrapolateRight {
		t.Fatalf("flags turned themselves back on: %+v", back)
	}
	if back.Asymmetry != -25 {
		t.Fatalf("asymmetry = %v, want -25", back.Asymmetry)
	}
}
