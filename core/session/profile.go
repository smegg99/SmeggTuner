package session

import (
	"sort"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

// Profile reads the rank voices this session's calibration takes recorded: a take of a single-bank
// register whose reading carries its own partial ratios (dsp.ReedMeasure.Harmonics, measured while
// the rank sounded alone) calibrates that rank at that note. A later take of the same note replaces
// the earlier - re-sweeping is how a profile is corrected. Entries key on the rank's octave and its
// own sounding note, which is what the engine's compound stage looks up.
func (s *Session) Profile() []dsp.RankProfile {
	type key struct {
		offset int
		note   int
	}
	found := map[key]dsp.RankProfile{}
	for _, t := range s.Takes {
		if t.Bass {
			continue // bass ranks key on their foot: see BassProfiles
		}
		r, ok := s.Instrument.Register(t.Register)
		if !ok || len(r.Banks) != 1 || len(t.Reeds) != 1 {
			continue
		}
		h := t.Reeds[0].Harmonics
		if len(h) == 0 {
			continue
		}
		p := dsp.RankProfile{
			Offset: r.Banks[0].Octave(),
			Note:   t.Note + tuning.Note(r.Banks[0].Octave()),
			R2:     h[0],
		}
		if len(h) > 1 {
			p.R4 = h[1]
		}
		found[key{p.Offset, int(p.Note)}] = p
	}
	if len(found) == 0 {
		return nil
	}
	out := make([]dsp.RankProfile, 0, len(found))
	for _, p := range found {
		out = append(out, p)
	}
	sort.Slice(out, func(a, b int) bool {
		if out[a].Offset != out[b].Offset {
			return out[a].Offset < out[b].Offset
		}
		return out[a].Note < out[b].Note
	})
	return out
}

// BassProfiles is the bass ranks' calibrated voices: takes of a single-rank bass register carrying
// partial ratios. A bass rank keys on its foot, not an octave offset - which octave slot it holds
// depends on which register is pulled, so the offset is resolved at impose time. The take's note is
// the rank's own sounding note: a solo rank is the layout's base.
func (s *Session) BassProfiles() []BassProfile {
	type key struct {
		foot int
		note int
	}
	found := map[key]BassProfile{}
	for _, t := range s.Takes {
		if !t.Bass {
			continue
		}
		r, ok := s.Instrument.BassRegister(t.Register)
		if !ok || len(r.Feet) != 1 || len(t.Reeds) != 1 {
			continue
		}
		h := t.Reeds[0].Harmonics
		if len(h) == 0 {
			continue
		}
		p := BassProfile{Foot: r.Feet[0], Note: t.Note, R2: h[0]}
		if len(h) > 1 {
			p.R4 = h[1]
		}
		found[key{p.Foot, int(p.Note)}] = p
	}
	if len(found) == 0 {
		return nil
	}
	out := make([]BassProfile, 0, len(found))
	for _, p := range found {
		out = append(out, p)
	}
	sort.Slice(out, func(a, b int) bool {
		if out[a].Foot != out[b].Foot {
			return out[a].Foot > out[b].Foot
		}
		return out[a].Note < out[b].Note
	})
	return out
}

// ProfileRev fingerprints the takes the profile is read from, so a consumer holding a comparable
// struct can tell when to re-read it. Replacing a take keeps the count but moves its At.
func (s *Session) ProfileRev() int64 {
	var newest int64
	for _, t := range s.Takes {
		if at := t.At.UnixNano(); at > newest {
			newest = at
		}
	}
	return newest ^ int64(len(s.Takes))
}
