package session

import (
	"testing"
	"time"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

func profiledSession() *Session {
	s := &Session{
		Instrument: Instrument{
			Banks: []Bank{BankL, BankM1},
			Registers: []Register{
				{Name: "L", Banks: []Bank{BankL}},
				{Name: "LM", Banks: []Bank{BankL, BankM1}},
			},
		},
	}
	return s
}

// A take of a solo-rank register carrying its partial ratios calibrates that rank at that note; a
// register sounding several ranks never does, whatever its reading carries.
func TestProfileReadsSoloRegisterTakes(t *testing.T) {
	s := profiledSession()
	s.Takes = []Take{
		{Note: 69, Register: "L", At: time.Unix(1, 0),
			Reeds: []dsp.ReedMeasure{{Freq: 220, Octave: -12, Harmonics: []float64{0.21, 0.48}}}},
		{Note: 69, Register: "LM", At: time.Unix(2, 0),
			Reeds: []dsp.ReedMeasure{{Freq: 220, Octave: -12, Harmonics: []float64{0.9}}, {Freq: 440}}},
	}

	p := s.Profile()
	if len(p) != 1 {
		t.Fatalf("want the one solo-register take, got %+v", p)
	}
	want := dsp.RankProfile{Offset: -12, Note: tuning.Note(57), R2: 0.21, R4: 0.48}
	if p[0] != want {
		t.Errorf("profile %+v, want %+v: keyed on the rank's own sounding note", p[0], want)
	}
}

// Re-sweeping replaces: the later take of the same note wins, and the fingerprint moves with it.
func TestProfileTakesTheLatestTake(t *testing.T) {
	s := profiledSession()
	s.Takes = []Take{
		{Note: 69, Register: "L", At: time.Unix(1, 0),
			Reeds: []dsp.ReedMeasure{{Freq: 220, Octave: -12, Harmonics: []float64{0.30}}}},
	}
	rev1 := s.ProfileRev()
	s.Takes = []Take{
		{Note: 69, Register: "L", At: time.Unix(5, 0),
			Reeds: []dsp.ReedMeasure{{Freq: 220, Octave: -12, Harmonics: []float64{0.19}}}},
	}

	if p := s.Profile(); len(p) != 1 || p[0].R2 != 0.19 {
		t.Errorf("want the later take's ratio, got %+v", p)
	}
	if s.ProfileRev() == rev1 {
		t.Error("replacing a take must move the fingerprint")
	}
}
