package session

import (
	"testing"
	"time"

	"smegg.me/smeggtuner/core/dsp"
)

func TestBassFeetNamesTheLadder(t *testing.T) {
	cases := map[int][]int{
		4: {16, 8, 4, 2},
		5: {32, 16, 8, 4, 2},
		6: {64, 32, 16, 8, 4, 2},
		7: nil,
		1: nil,
	}
	for count, want := range cases {
		got := BassFeet(count)
		if len(got) != len(want) {
			t.Fatalf("BassFeet(%d) = %v, want %v", count, got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("BassFeet(%d) = %v, want %v", count, got, want)
			}
		}
	}
}

func TestOctavesOfFeetClimbFromTheLargestFoot(t *testing.T) {
	reqs := OctavesOfFeet([]int{32, 8, 16})
	wantOff := []int{0, 12, 24}
	if len(reqs) != 3 {
		t.Fatalf("want 3 bands, got %+v", reqs)
	}
	for i, r := range reqs {
		if r.Offset != wantOff[i] || r.Reeds != 1 {
			t.Errorf("band %d: got %+v, want offset %d", i, r, wantOff[i])
		}
	}
	if OctavesOfFeet([]int{8}) != nil {
		t.Error("a single rank is the plain tuner's job, not a compound layout")
	}
}

func TestBassValidation(t *testing.T) {
	i := Instrument{ReedCount: 3, BassReeds: 5, BassRegisters: []BassRegister{
		{Name: "Soft", Feet: []int{16, 8}},
	}}
	if err := i.validate(); err != nil {
		t.Fatalf("a 5-voice machine with a 16.8 switch is legal: %v", err)
	}

	bad := []Instrument{
		{ReedCount: 3, BassReeds: 9},
		{ReedCount: 3, BassRegisters: []BassRegister{{Name: "X", Feet: []int{8}}}},
		{ReedCount: 3, BassReeds: 4, BassRegisters: []BassRegister{{Name: "X", Feet: []int{64}}}},
		{ReedCount: 3, BassReeds: 4, BassRegisters: []BassRegister{{Name: "X", Feet: []int{8, 8}}}},
	}
	for n, i := range bad {
		if err := i.validate(); err == nil {
			t.Errorf("case %d must be refused: %+v", n, i)
		}
	}
}

// A bass take of a solo-rank switch calibrates that rank by its foot; the same note's treble take
// stays out of it, and a bass take never collides with a treble take of the same note.
func TestBassProfilesKeyOnTheFoot(t *testing.T) {
	s := &Session{Instrument: Instrument{
		BassReeds:     5,
		BassRegisters: []BassRegister{{Name: "Solo16", Feet: []int{16}}},
	}}
	s.UpsertTake(Take{Note: 45, Bass: true, Register: "Solo16", At: time.Unix(1, 0),
		Reeds: []dsp.ReedMeasure{{Freq: 110, Harmonics: []float64{0.4, 0.1}}}})
	s.UpsertTake(Take{Note: 45, Register: "", At: time.Unix(2, 0),
		Reeds: []dsp.ReedMeasure{{Freq: 110}}})

	if len(s.Takes) != 2 {
		t.Fatalf("a bass and a treble take of one note are different voices, got %d takes", len(s.Takes))
	}
	p := s.BassProfiles()
	if len(p) != 1 || p[0].Foot != 16 || p[0].Note != 45 || p[0].R2 != 0.4 {
		t.Errorf("bass profile %+v, want foot 16 at note 45 with R2 0.4", p)
	}
	if s.Profile() != nil {
		t.Error("a bass take must not enter the treble profile")
	}
}
