package session

import (
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/dsp"
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

// A bass take never collides with a treble take of the same note: the two sides are different voices.
func TestABassTakeIsItsOwnVoice(t *testing.T) {
	s := &Session{Instrument: Instrument{
		BassReeds:     5,
		BassRegisters: []BassRegister{{Name: "Solo16", Feet: []int{16}}},
	}}
	s.UpsertTake(Take{Note: 45, Bass: true, Register: "Solo16", At: time.Unix(1, 0),
		Reeds: []dsp.ReedMeasure{{Present: true, Freq: 110}}})
	s.UpsertTake(Take{Note: 45, Register: "", At: time.Unix(2, 0),
		Reeds: []dsp.ReedMeasure{{Present: true, Freq: 110}}})

	if len(s.Takes) != 2 {
		t.Fatalf("a bass and a treble take of one note are different voices, got %d takes", len(s.Takes))
	}
}
