package report

import (
	"testing"

	"smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/target"
)

func TestReedCounts(t *testing.T) {
	for _, tc := range []struct {
		reeds     int
		wantPairs []string
		grouped   bool
		landscape bool
	}{
		{reeds: 1, wantPairs: nil},
		{reeds: 2, wantPairs: []string{"1-2"}},
		{reeds: 3, wantPairs: []string{"1-2", "2-3"}},
		{reeds: 5, wantPairs: []string{"1-2", "2-3", "3-4", "4-5"}, landscape: true},
		{reeds: 8, wantPairs: []string{"1-2", "2-3", "3-4", "4-5", "5-6", "6-7", "7-8"}, grouped: true},
	} {
		s := session.New("Bench", session.Instrument{ReedCount: tc.reeds}, passA4)
		cents := make([]float64, tc.reeds)
		for i := range cents {
			cents[i] = float64(i) * 4
		}
		s.UpsertTake(reeds(60, cents...))

		r := sheet(t, s)
		if len(r.Reeds) != tc.reeds {
			t.Fatalf("%d reeds: report has %d reed columns", tc.reeds, len(r.Reeds))
		}
		if len(r.Rows[0].Reeds) != tc.reeds {
			t.Fatalf("%d reeds: row has %d cells", tc.reeds, len(r.Rows[0].Reeds))
		}
		var pairs []string
		for _, p := range r.Pairs {
			pairs = append(pairs, p.Key)
		}
		if !equal(pairs, tc.wantPairs) {
			t.Errorf("%d reeds: beat columns %v, want %v", tc.reeds, pairs, tc.wantPairs)
		}
		if r.Layout.Grouped != tc.grouped || r.Layout.Landscape != tc.landscape {
			t.Errorf("%d reeds (%d columns): layout %+v, want grouped=%v landscape=%v",
				tc.reeds, r.Columns(), r.Layout, tc.grouped, tc.landscape)
		}
	}
}

func TestMeasuredPairAlwaysHasAColumn(t *testing.T) {
	s := session.New("Bench", session.Instrument{ReedCount: 1}, passA4)
	s.UpsertTake(mergedTake(60))

	r := sheet(t, s)
	if len(r.Pairs) != 1 || r.Pairs[0].Key != "1-2" {
		t.Fatalf("beat columns %+v, want the measured 1-2", r.Pairs)
	}
	if !r.Rows[0].Beats[0].Present {
		t.Fatal("the measured beat has no cell to print in")
	}
	if len(r.Rows[0].Reeds) != 0 {
		t.Fatal("a merged row printed reed cells")
	}
}

func TestBuildRefusesNothing(t *testing.T) {
	if _, err := Build(nil, Options{}); err != ErrNoSession {
		t.Errorf("Build(nil) = %v, want ErrNoSession", err)
	}

	empty := session.New("Bench", session.Instrument{ReedCount: 2}, passA4)
	if _, err := Build(empty, Options{}); err != ErrNoReadings {
		t.Errorf("Build(empty pass) = %v, want ErrNoReadings", err)
	}
	if _, err := Build(session.New("Bench", session.Instrument{ReedCount: 2}, passA4), Options{}); err != ErrNoReadings {
		t.Error("a session with nothing recorded should refuse rather than print an empty sheet")
	}
}

func TestSummary(t *testing.T) {
	r := sheet(t, musette(t))
	s := r.Summary

	if s.Notes != 5 {
		t.Errorf("notes = %d, want 5", s.Notes)
	}
	if s.Merged != 1 || s.Derived != 1 || s.Manual != 1 {
		t.Errorf("merged/derived/manual = %d/%d/%d, want 1/1/1", s.Merged, s.Derived, s.Manual)
	}
	// Four rows of reeds: three full, one missing its third; the merged row contributes none.
	if s.Reeds != 11 {
		t.Errorf("reed readings = %d, want 11", s.Reeds)
	}
	if s.MinName == "" || s.MaxName == "" {
		t.Error("the range is not named")
	}
	if s.Tolerance != target.DefaultTolerance || s.BeatTol != target.DefaultBeatTolerance {
		t.Errorf("tolerances = %.1f/%.1f, want the defaults", s.Tolerance, s.BeatTol)
	}
}

func TestLayoutForColumns(t *testing.T) {
	for _, tc := range []struct {
		columns int
		want    Layout
	}{
		{4, Layout{}},                 // one reed
		{16, Layout{}},                // three reeds, last that fits portrait
		{17, Layout{Landscape: true}}, // four reeds
		{28, Layout{Landscape: true}}, // five
		{29, Layout{Grouped: true}},   // six reeds
		{100, Layout{Grouped: true}},
	} {
		if got := layoutFor(tc.columns); got != tc.want {
			t.Errorf("layoutFor(%d) = %+v, want %+v", tc.columns, got, tc.want)
		}
	}
}
