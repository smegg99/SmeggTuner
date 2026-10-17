package tuner

import (
	"slices"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	coresession "smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// impose folds the active session over an engine config: its reference pitch and transpose become
// the engine's, its bench sets the reed count and the octave layout (the pulled register's banks, or facing the
// bass side the sounding ladder's feet), and its goal curve becomes the engine's slot targets. With
// no session open the panel decides.
func impose(c dsp.EngineConfig, pn panel, p imposed) dsp.EngineConfig {
	c.Chords, c.Reeds, c.Octaves, c.Bass = false, nil, nil, false
	c.Targets, c.TargetCoverage = nil, meter.NoteCoverage{}
	if p.a4 <= 0 {
		return pn.apply(c)
	}
	c.A4, c.Transpose = p.a4, p.transpose
	if p.reeds > 0 {
		c.ReedCount = clampReeds(p.reeds)
	}
	feet := splitFeet(p.bassFeet)
	c.Bass = len(feet) > 0
	if c.Bass {
		c.Octaves = coresession.OctavesOfFeet(feet)
	} else {
		banks := splitBanks(p.banks)
		c.Octaves = coresession.OctavesOf(banks)
		if c.Reeds = reedBoxes(splitBanks(p.cardBanks), banks); c.Reeds != nil {
			c.ReedCount = len(c.Reeds)
		}
	}
	slots := slotCount(c)
	c.Targets, c.TargetCoverage = targetsOf(p.curve, p.columns(slots), slots)
	return c
}

// reedBoxes maps a unison register of up to three banks onto the reference reed boxes: each bank's
// place among the instrument's banks at the same foot, so M2 and M3 of an MMM card are boxes 2 and 3.
// Nil when the register spans octaves or does not fit boxes 1..3.
func reedBoxes(card, register []coresession.Bank) []int {
	if len(register) == 0 || len(register) > 3 || coresession.OctavesOf(register) != nil {
		return nil
	}
	var unison []coresession.Bank
	for _, b := range card {
		if b.Octave() == 0 {
			unison = append(unison, b)
		}
	}
	var ids []int
	for _, b := range coresession.SlotBanks(register, len(register)) {
		id := slices.Index(unison, b) + 1
		if id < 1 || id > 3 {
			return nil
		}
		ids = append(ids, id)
	}
	if len(ids) != len(register) {
		return nil
	}
	return ids
}

// slotCount is how many slots the engine measures for c: the layout's reeds when it spans octaves,
// else the reed count.
func slotCount(c dsp.EngineConfig) int {
	n, spans := 0, false
	for _, o := range c.Octaves {
		n += o.Reeds
		spans = spans || o.Offset != 0
	}
	if !spans {
		return c.ReedCount
	}
	return n
}

// targetsOf is the goal curve as the engine takes it: per slot, the slot's column at every played key.
// The engine moves it onto the note it analyses. Notes the curve does not reach stay Missing and
// uncovered, as the reference leaves notes outside its anchor spans. Nil without anchors.
func targetsOf(c *target.Curve, cols []int, slots int) ([]meter.NoteCurve, meter.NoteCoverage) {
	var coverage meter.NoteCoverage
	if c == nil || len(c.Anchors) == 0 {
		return nil, coverage
	}
	out := make([]meter.NoteCurve, slots)
	for i := range meter.ActiveNotes {
		key := tuning.Note(i + meter.FirstActiveNote + 12)
		goal := c.At(key)
		coverage[i] = reaches(c, key)
		for slot := range out {
			out[slot][i] = meter.Missing
			if col := column(cols, slot); coverage[i] && col < len(goal) {
				out[slot][i] = goal[col]
			}
		}
	}
	return out, coverage
}

// reaches reports whether the curve sets a goal at key: inside its anchors, or past an end it extrapolates.
func reaches(c *target.Curve, key tuning.Note) bool {
	first, last := c.Anchors[0].Note, c.Anchors[len(c.Anchors)-1].Note
	switch {
	case !key.Valid():
		return false
	case key < first:
		return c.ExtrapolateLeft
	case key > last:
		return c.ExtrapolateRight
	}
	return true
}

func column(cols []int, slot int) int {
	if slot < len(cols) {
		return cols[slot]
	}
	return slot
}
